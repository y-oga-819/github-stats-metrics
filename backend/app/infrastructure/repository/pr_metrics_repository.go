package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github-stats-metrics/domain/analytics"
	prDomain "github-stats-metrics/domain/pull_request"
)

// PRMetricsRepository はPRメトリクスの永続化を担当するリポジトリ
type PRMetricsRepository struct {
	db *sql.DB
}

// NewPRMetricsRepository は新しいPRメトリクスリポジトリを作成
func NewPRMetricsRepository(db *sql.DB) *PRMetricsRepository {
	return &PRMetricsRepository{
		db: db,
	}
}

// Save はPRメトリクスを保存
func (repo *PRMetricsRepository) Save(ctx context.Context, metrics *prDomain.PRMetrics) error {
	storage, err := repo.convertToStorage(metrics)
	if err != nil {
		return fmt.Errorf("failed to convert metrics to storage model: %w", err)
	}

	// PRメトリクス本体の保存
	if err := repo.savePRMetrics(ctx, storage); err != nil {
		return fmt.Errorf("failed to save pr metrics: %w", err)
	}

	// ファイル変更情報の保存
	if err := repo.saveFileChanges(ctx, storage.ID, metrics.SizeMetrics.FileChanges); err != nil {
		return fmt.Errorf("failed to save file changes: %w", err)
	}

	return nil
}

// SaveBatch は複数のPRメトリクスを一括保存
func (repo *PRMetricsRepository) SaveBatch(ctx context.Context, metricsList []*prDomain.PRMetrics) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, metrics := range metricsList {
		storage, err := repo.convertToStorage(metrics)
		if err != nil {
			return fmt.Errorf("failed to convert metrics to storage model: %w", err)
		}

		// PRメトリクス本体の保存
		if err := repo.savePRMetricsWithTx(ctx, tx, storage); err != nil {
			return fmt.Errorf("failed to save pr metrics: %w", err)
		}

		// ファイル変更情報の保存
		if err := repo.saveFileChangesWithTx(ctx, tx, storage.ID, metrics.SizeMetrics.FileChanges); err != nil {
			return fmt.Errorf("failed to save file changes: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByID はIDによりPRメトリクスを取得
func (repo *PRMetricsRepository) FindByID(ctx context.Context, id string) (*prDomain.PRMetrics, error) {
	storage, err := repo.findStorageByID(ctx, id)
	if err != nil {
		return nil, err
	}

	metrics, err := repo.convertFromStorage(storage)
	if err != nil {
		return nil, fmt.Errorf("failed to convert from storage model: %w", err)
	}

	return metrics, nil
}

// FindByPRID はPR IDによりPRメトリクスを取得
func (repo *PRMetricsRepository) FindByPRID(ctx context.Context, prID string) (*prDomain.PRMetrics, error) {
	query := `
		SELECT id, pr_id, pr_number, title, author, repository, created_at, merged_at, collected_at,
			   size_metrics_json, total_cycle_time_seconds, time_to_first_review_seconds,
			   time_to_approval_seconds, time_to_merge_seconds, time_metrics_json,
			   review_comment_count, review_round_count, reviewer_count, first_review_pass_rate,
			   quality_metrics_json, complexity_score, size_category,
			   year_month, week_of_year, day_of_year
		FROM pr_metrics
		WHERE pr_id = $1
	`

	var storage analytics.PRMetricsStorage
	err := repo.db.QueryRowContext(ctx, query, prID).Scan(
		&storage.ID, &storage.PRID, &storage.PRNumber, &storage.Title, &storage.Author,
		&storage.Repository, &storage.CreatedAt, &storage.MergedAt, &storage.CollectedAt,
		&storage.SizeMetricsJSON, &storage.TotalCycleTimeSeconds, &storage.TimeToFirstReviewSeconds,
		&storage.TimeToApprovalSeconds, &storage.TimeToMergeSeconds, &storage.TimeMetricsJSON,
		&storage.ReviewCommentCount, &storage.ReviewRoundCount, &storage.ReviewerCount,
		&storage.FirstReviewPassRate, &storage.QualityMetricsJSON, &storage.ComplexityScore,
		&storage.SizeCategory, &storage.YearMonth, &storage.WeekOfYear, &storage.DayOfYear,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find pr metrics by pr_id: %w", err)
	}

	return repo.convertFromStorage(&storage)
}

// FindByDateRange は日付範囲によりPRメトリクスを取得
func (repo *PRMetricsRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time, developers []string, repositories []string) ([]*prDomain.PRMetrics, error) {
	query := `
		SELECT id, pr_id, pr_number, title, author, repository, created_at, merged_at, collected_at,
			   size_metrics_json, total_cycle_time_seconds, time_to_first_review_seconds,
			   time_to_approval_seconds, time_to_merge_seconds, time_metrics_json,
			   review_comment_count, review_round_count, reviewer_count, first_review_pass_rate,
			   quality_metrics_json, complexity_score, size_category,
			   year_month, week_of_year, day_of_year
		FROM pr_metrics
		WHERE created_at >= $1 AND created_at <= $2
	`

	args := []interface{}{startDate, endDate}
	argIndex := 3

	// 開発者フィルタ
	if len(developers) > 0 {
		query += fmt.Sprintf(" AND author = ANY($%d)", argIndex)
		args = append(args, developers)
		argIndex++
	}

	// リポジトリフィルタ
	if len(repositories) > 0 {
		query += fmt.Sprintf(" AND repository = ANY($%d)", argIndex)
		args = append(args, repositories)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pr metrics by date range: %w", err)
	}
	defer rows.Close()

	var metricsList []*prDomain.PRMetrics
	for rows.Next() {
		var storage analytics.PRMetricsStorage
		err := rows.Scan(
			&storage.ID, &storage.PRID, &storage.PRNumber, &storage.Title, &storage.Author,
			&storage.Repository, &storage.CreatedAt, &storage.MergedAt, &storage.CollectedAt,
			&storage.SizeMetricsJSON, &storage.TotalCycleTimeSeconds, &storage.TimeToFirstReviewSeconds,
			&storage.TimeToApprovalSeconds, &storage.TimeToMergeSeconds, &storage.TimeMetricsJSON,
			&storage.ReviewCommentCount, &storage.ReviewRoundCount, &storage.ReviewerCount,
			&storage.FirstReviewPassRate, &storage.QualityMetricsJSON, &storage.ComplexityScore,
			&storage.SizeCategory, &storage.YearMonth, &storage.WeekOfYear, &storage.DayOfYear,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pr metrics row: %w", err)
		}

		metrics, err := repo.convertFromStorage(&storage)
		if err != nil {
			return nil, fmt.Errorf("failed to convert from storage model: %w", err)
		}

		metricsList = append(metricsList, metrics)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate pr metrics rows: %w", err)
	}

	return metricsList, nil
}

// FindByDeveloper は開発者によりPRメトリクスを取得
func (repo *PRMetricsRepository) FindByDeveloper(ctx context.Context, developer string, startDate, endDate time.Time) ([]*prDomain.PRMetrics, error) {
	return repo.FindByDateRange(ctx, startDate, endDate, []string{developer}, nil)
}

// FindByRepository はリポジトリによりPRメトリクスを取得
func (repo *PRMetricsRepository) FindByRepository(ctx context.Context, repository string, startDate, endDate time.Time) ([]*prDomain.PRMetrics, error) {
	return repo.FindByDateRange(ctx, startDate, endDate, nil, []string{repository})
}

// Update はPRメトリクスを更新
func (repo *PRMetricsRepository) Update(ctx context.Context, metrics *prDomain.PRMetrics) error {
	storage, err := repo.convertToStorage(metrics)
	if err != nil {
		return fmt.Errorf("failed to convert metrics to storage model: %w", err)
	}

	query := `
		UPDATE pr_metrics SET
			pr_number = $2, title = $3, author = $4, repository = $5,
			created_at = $6, merged_at = $7, collected_at = $8,
			size_metrics_json = $9, total_cycle_time_seconds = $10,
			time_to_first_review_seconds = $11, time_to_approval_seconds = $12,
			time_to_merge_seconds = $13, time_metrics_json = $14,
			review_comment_count = $15, review_round_count = $16,
			reviewer_count = $17, first_review_pass_rate = $18,
			quality_metrics_json = $19, complexity_score = $20,
			size_category = $21, year_month = $22, week_of_year = $23, day_of_year = $24
		WHERE id = $1
	`

	_, err = repo.db.ExecContext(ctx, query,
		storage.ID, storage.PRNumber, storage.Title, storage.Author, storage.Repository,
		storage.CreatedAt, storage.MergedAt, storage.CollectedAt,
		storage.SizeMetricsJSON, storage.TotalCycleTimeSeconds,
		storage.TimeToFirstReviewSeconds, storage.TimeToApprovalSeconds,
		storage.TimeToMergeSeconds, storage.TimeMetricsJSON,
		storage.ReviewCommentCount, storage.ReviewRoundCount,
		storage.ReviewerCount, storage.FirstReviewPassRate,
		storage.QualityMetricsJSON, storage.ComplexityScore,
		storage.SizeCategory, storage.YearMonth, storage.WeekOfYear, storage.DayOfYear,
	)

	if err != nil {
		return fmt.Errorf("failed to update pr metrics: %w", err)
	}

	return nil
}

// Delete はPRメトリクスを削除
func (repo *PRMetricsRepository) Delete(ctx context.Context, id string) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 関連データの削除
	if err := repo.deleteFileChanges(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to delete file changes: %w", err)
	}

	if err := repo.deleteReviewEvents(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to delete review events: %w", err)
	}

	// メインデータの削除
	if err := repo.deletePRMetrics(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to delete pr metrics: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteOldData は古いデータを削除
func (repo *PRMetricsRepository) DeleteOldData(ctx context.Context, retentionDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	
	query := `DELETE FROM pr_metrics WHERE collected_at < $1`
	result, err := repo.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetStatistics はリポジトリの統計情報を取得
func (repo *PRMetricsRepository) GetStatistics(ctx context.Context) (*RepositoryStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_records,
			COUNT(DISTINCT author) as unique_developers,
			COUNT(DISTINCT repository) as unique_repositories,
			MIN(created_at) as oldest_record,
			MAX(created_at) as newest_record,
			AVG(total_cycle_time_seconds) as avg_cycle_time,
			AVG(complexity_score) as avg_complexity
		FROM pr_metrics
	`

	var stats RepositoryStatistics
	err := repo.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalRecords,
		&stats.UniqueDevelopers,
		&stats.UniqueRepositories,
		&stats.OldestRecord,
		&stats.NewestRecord,
		&stats.AvgCycleTime,
		&stats.AvgComplexity,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get repository statistics: %w", err)
	}

	return &stats, nil
}

// RepositoryStatistics はリポジトリの統計情報
type RepositoryStatistics struct {
	TotalRecords        int64     `json:"totalRecords"`
	UniqueDevelopers    int64     `json:"uniqueDevelopers"`
	UniqueRepositories  int64     `json:"uniqueRepositories"`
	OldestRecord        time.Time `json:"oldestRecord"`
	NewestRecord        time.Time `json:"newestRecord"`
	AvgCycleTime        *float64  `json:"avgCycleTime"`
	AvgComplexity       *float64  `json:"avgComplexity"`
}

// プライベートメソッド

func (repo *PRMetricsRepository) convertToStorage(metrics *prDomain.PRMetrics) (*analytics.PRMetricsStorage, error) {
	// サイズメトリクスをJSONに変換
	sizeMetricsJSON, err := json.Marshal(metrics.SizeMetrics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal size metrics: %w", err)
	}

	// 時間メトリクスをJSONに変換
	timeMetricsJSON, err := json.Marshal(metrics.TimeMetrics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal time metrics: %w", err)
	}

	// 品質メトリクスをJSONに変換
	qualityMetricsJSON, err := json.Marshal(metrics.QualityMetrics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal quality metrics: %w", err)
	}

	// 時間を秒に変換
	var totalCycleTimeSeconds *int64
	if metrics.TimeMetrics.TotalCycleTime != nil {
		seconds := int64(metrics.TimeMetrics.TotalCycleTime.Seconds())
		totalCycleTimeSeconds = &seconds
	}

	var timeToFirstReviewSeconds *int64
	if metrics.TimeMetrics.TimeToFirstReview != nil {
		seconds := int64(metrics.TimeMetrics.TimeToFirstReview.Seconds())
		timeToFirstReviewSeconds = &seconds
	}

	var timeToApprovalSeconds *int64
	if metrics.TimeMetrics.TimeToApproval != nil {
		seconds := int64(metrics.TimeMetrics.TimeToApproval.Seconds())
		timeToApprovalSeconds = &seconds
	}

	var timeToMergeSeconds *int64
	if metrics.TimeMetrics.TimeToMerge != nil {
		seconds := int64(metrics.TimeMetrics.TimeToMerge.Seconds())
		timeToMergeSeconds = &seconds
	}

	// インデックス用フィールドを生成
	yearMonth := metrics.CreatedAt.Format("2006-01")
	year, week := metrics.CreatedAt.ISOWeek()
	weekOfYear := fmt.Sprintf("%d-W%02d", year, week)
	dayOfYear := metrics.CreatedAt.Format("2006-002")

	return &analytics.PRMetricsStorage{
		ID:          fmt.Sprintf("pr_%s_%d", metrics.PRID, time.Now().Unix()),
		PRID:        metrics.PRID,
		PRNumber:    metrics.PRNumber,
		Title:       metrics.Title,
		Author:      metrics.Author,
		Repository:  metrics.Repository,
		CreatedAt:   metrics.CreatedAt,
		MergedAt:    metrics.MergedAt,
		CollectedAt: time.Now(),

		SizeMetricsJSON: string(sizeMetricsJSON),

		TotalCycleTimeSeconds:    totalCycleTimeSeconds,
		TimeToFirstReviewSeconds: timeToFirstReviewSeconds,
		TimeToApprovalSeconds:    timeToApprovalSeconds,
		TimeToMergeSeconds:       timeToMergeSeconds,
		TimeMetricsJSON:          string(timeMetricsJSON),

		ReviewCommentCount:   metrics.QualityMetrics.ReviewCommentCount,
		ReviewRoundCount:     metrics.QualityMetrics.ReviewRoundCount,
		ReviewerCount:        metrics.QualityMetrics.ReviewerCount,
		FirstReviewPassRate:  metrics.QualityMetrics.FirstReviewPassRate,
		QualityMetricsJSON:   string(qualityMetricsJSON),

		ComplexityScore: metrics.ComplexityScore,
		SizeCategory:    metrics.SizeCategory,

		YearMonth:  yearMonth,
		WeekOfYear: weekOfYear,
		DayOfYear:  dayOfYear,
	}, nil
}

func (repo *PRMetricsRepository) convertFromStorage(storage *analytics.PRMetricsStorage) (*prDomain.PRMetrics, error) {
	// JSONからサイズメトリクスを復元
	var sizeMetrics prDomain.PRSizeMetrics
	if err := json.Unmarshal([]byte(storage.SizeMetricsJSON), &sizeMetrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal size metrics: %w", err)
	}

	// JSONから時間メトリクスを復元
	var timeMetrics prDomain.PRTimeMetrics
	if err := json.Unmarshal([]byte(storage.TimeMetricsJSON), &timeMetrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal time metrics: %w", err)
	}

	// JSONから品質メトリクスを復元
	var qualityMetrics prDomain.PRQualityMetrics
	if err := json.Unmarshal([]byte(storage.QualityMetricsJSON), &qualityMetrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal quality metrics: %w", err)
	}

	return &prDomain.PRMetrics{
		PRID:           storage.PRID,
		PRNumber:       storage.PRNumber,
		Title:          storage.Title,
		Author:         storage.Author,
		Repository:     storage.Repository,
		CreatedAt:      storage.CreatedAt,
		MergedAt:       storage.MergedAt,
		SizeMetrics:    sizeMetrics,
		TimeMetrics:    timeMetrics,
		QualityMetrics: qualityMetrics,
		ComplexityScore: storage.ComplexityScore,
		SizeCategory:   storage.SizeCategory,
	}, nil
}

func (repo *PRMetricsRepository) savePRMetrics(ctx context.Context, storage *analytics.PRMetricsStorage) error {
	query := `
		INSERT INTO pr_metrics (
			id, pr_id, pr_number, title, author, repository, created_at, merged_at, collected_at,
			size_metrics_json, total_cycle_time_seconds, time_to_first_review_seconds,
			time_to_approval_seconds, time_to_merge_seconds, time_metrics_json,
			review_comment_count, review_round_count, reviewer_count, first_review_pass_rate,
			quality_metrics_json, complexity_score, size_category,
			year_month, week_of_year, day_of_year
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25
		) ON CONFLICT (pr_id) DO UPDATE SET
			pr_number = EXCLUDED.pr_number,
			title = EXCLUDED.title,
			author = EXCLUDED.author,
			repository = EXCLUDED.repository,
			created_at = EXCLUDED.created_at,
			merged_at = EXCLUDED.merged_at,
			collected_at = EXCLUDED.collected_at,
			size_metrics_json = EXCLUDED.size_metrics_json,
			total_cycle_time_seconds = EXCLUDED.total_cycle_time_seconds,
			time_to_first_review_seconds = EXCLUDED.time_to_first_review_seconds,
			time_to_approval_seconds = EXCLUDED.time_to_approval_seconds,
			time_to_merge_seconds = EXCLUDED.time_to_merge_seconds,
			time_metrics_json = EXCLUDED.time_metrics_json,
			review_comment_count = EXCLUDED.review_comment_count,
			review_round_count = EXCLUDED.review_round_count,
			reviewer_count = EXCLUDED.reviewer_count,
			first_review_pass_rate = EXCLUDED.first_review_pass_rate,
			quality_metrics_json = EXCLUDED.quality_metrics_json,
			complexity_score = EXCLUDED.complexity_score,
			size_category = EXCLUDED.size_category,
			year_month = EXCLUDED.year_month,
			week_of_year = EXCLUDED.week_of_year,
			day_of_year = EXCLUDED.day_of_year
	`

	_, err := repo.db.ExecContext(ctx, query,
		storage.ID, storage.PRID, storage.PRNumber, storage.Title, storage.Author,
		storage.Repository, storage.CreatedAt, storage.MergedAt, storage.CollectedAt,
		storage.SizeMetricsJSON, storage.TotalCycleTimeSeconds,
		storage.TimeToFirstReviewSeconds, storage.TimeToApprovalSeconds,
		storage.TimeToMergeSeconds, storage.TimeMetricsJSON,
		storage.ReviewCommentCount, storage.ReviewRoundCount,
		storage.ReviewerCount, storage.FirstReviewPassRate,
		storage.QualityMetricsJSON, storage.ComplexityScore,
		storage.SizeCategory, storage.YearMonth, storage.WeekOfYear, storage.DayOfYear,
	)

	return err
}

func (repo *PRMetricsRepository) savePRMetricsWithTx(ctx context.Context, tx *sql.Tx, storage *analytics.PRMetricsStorage) error {
	query := `
		INSERT INTO pr_metrics (
			id, pr_id, pr_number, title, author, repository, created_at, merged_at, collected_at,
			size_metrics_json, total_cycle_time_seconds, time_to_first_review_seconds,
			time_to_approval_seconds, time_to_merge_seconds, time_metrics_json,
			review_comment_count, review_round_count, reviewer_count, first_review_pass_rate,
			quality_metrics_json, complexity_score, size_category,
			year_month, week_of_year, day_of_year
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25
		) ON CONFLICT (pr_id) DO NOTHING
	`

	_, err := tx.ExecContext(ctx, query,
		storage.ID, storage.PRID, storage.PRNumber, storage.Title, storage.Author,
		storage.Repository, storage.CreatedAt, storage.MergedAt, storage.CollectedAt,
		storage.SizeMetricsJSON, storage.TotalCycleTimeSeconds,
		storage.TimeToFirstReviewSeconds, storage.TimeToApprovalSeconds,
		storage.TimeToMergeSeconds, storage.TimeMetricsJSON,
		storage.ReviewCommentCount, storage.ReviewRoundCount,
		storage.ReviewerCount, storage.FirstReviewPassRate,
		storage.QualityMetricsJSON, storage.ComplexityScore,
		storage.SizeCategory, storage.YearMonth, storage.WeekOfYear, storage.DayOfYear,
	)

	return err
}

func (repo *PRMetricsRepository) saveFileChanges(ctx context.Context, prMetricsID string, fileChanges []prDomain.FileChangeMetrics) error {
	if len(fileChanges) == 0 {
		return nil
	}

	// 既存のファイル変更データを削除
	deleteQuery := `DELETE FROM file_changes WHERE pr_metrics_id = $1`
	if _, err := repo.db.ExecContext(ctx, deleteQuery, prMetricsID); err != nil {
		return fmt.Errorf("failed to delete existing file changes: %w", err)
	}

	// 新しいファイル変更データを挿入
	insertQuery := `
		INSERT INTO file_changes (
			id, pr_metrics_id, file_name, file_type, lines_added, lines_deleted,
			is_new_file, is_deleted, is_renamed, collected_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	for i, fileChange := range fileChanges {
		id := fmt.Sprintf("%s_file_%d", prMetricsID, i)
		_, err := repo.db.ExecContext(ctx, insertQuery,
			id, prMetricsID, fileChange.FileName, fileChange.FileType,
			fileChange.LinesAdded, fileChange.LinesDeleted,
			fileChange.IsNewFile, fileChange.IsDeleted, fileChange.IsRenamed,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert file change: %w", err)
		}
	}

	return nil
}

func (repo *PRMetricsRepository) saveFileChangesWithTx(ctx context.Context, tx *sql.Tx, prMetricsID string, fileChanges []prDomain.FileChangeMetrics) error {
	if len(fileChanges) == 0 {
		return nil
	}

	insertQuery := `
		INSERT INTO file_changes (
			id, pr_metrics_id, file_name, file_type, lines_added, lines_deleted,
			is_new_file, is_deleted, is_renamed, collected_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO NOTHING
	`

	for i, fileChange := range fileChanges {
		id := fmt.Sprintf("%s_file_%d", prMetricsID, i)
		_, err := tx.ExecContext(ctx, insertQuery,
			id, prMetricsID, fileChange.FileName, fileChange.FileType,
			fileChange.LinesAdded, fileChange.LinesDeleted,
			fileChange.IsNewFile, fileChange.IsDeleted, fileChange.IsRenamed,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert file change: %w", err)
		}
	}

	return nil
}

func (repo *PRMetricsRepository) findStorageByID(ctx context.Context, id string) (*analytics.PRMetricsStorage, error) {
	query := `
		SELECT id, pr_id, pr_number, title, author, repository, created_at, merged_at, collected_at,
			   size_metrics_json, total_cycle_time_seconds, time_to_first_review_seconds,
			   time_to_approval_seconds, time_to_merge_seconds, time_metrics_json,
			   review_comment_count, review_round_count, reviewer_count, first_review_pass_rate,
			   quality_metrics_json, complexity_score, size_category,
			   year_month, week_of_year, day_of_year
		FROM pr_metrics
		WHERE id = $1
	`

	var storage analytics.PRMetricsStorage
	err := repo.db.QueryRowContext(ctx, query, id).Scan(
		&storage.ID, &storage.PRID, &storage.PRNumber, &storage.Title, &storage.Author,
		&storage.Repository, &storage.CreatedAt, &storage.MergedAt, &storage.CollectedAt,
		&storage.SizeMetricsJSON, &storage.TotalCycleTimeSeconds, &storage.TimeToFirstReviewSeconds,
		&storage.TimeToApprovalSeconds, &storage.TimeToMergeSeconds, &storage.TimeMetricsJSON,
		&storage.ReviewCommentCount, &storage.ReviewRoundCount, &storage.ReviewerCount,
		&storage.FirstReviewPassRate, &storage.QualityMetricsJSON, &storage.ComplexityScore,
		&storage.SizeCategory, &storage.YearMonth, &storage.WeekOfYear, &storage.DayOfYear,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find pr metrics by id: %w", err)
	}

	return &storage, nil
}

func (repo *PRMetricsRepository) deleteFileChanges(ctx context.Context, tx *sql.Tx, prMetricsID string) error {
	query := `DELETE FROM file_changes WHERE pr_metrics_id = $1`
	_, err := tx.ExecContext(ctx, query, prMetricsID)
	return err
}

func (repo *PRMetricsRepository) deleteReviewEvents(ctx context.Context, tx *sql.Tx, prMetricsID string) error {
	query := `DELETE FROM review_events WHERE pr_metrics_id = $1`
	_, err := tx.ExecContext(ctx, query, prMetricsID)
	return err
}

func (repo *PRMetricsRepository) deletePRMetrics(ctx context.Context, tx *sql.Tx, id string) error {
	query := `DELETE FROM pr_metrics WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	return err
}