package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	analyticsApp "github-stats-metrics/application/analytics"
	"github-stats-metrics/domain/analytics"
)

// AggregatedMetricsRepository は集計済みメトリクスの永続化を担当するリポジトリ
type AggregatedMetricsRepository struct {
	db *sql.DB
}

// NewAggregatedMetricsRepository は新しい集計メトリクスリポジトリを作成
func NewAggregatedMetricsRepository(db *sql.DB) *AggregatedMetricsRepository {
	return &AggregatedMetricsRepository{
		db: db,
	}
}

// SaveTeamMetrics はチームメトリクスを保存
func (repo *AggregatedMetricsRepository) SaveTeamMetrics(ctx context.Context, metrics *analyticsApp.TeamMetrics) error {
	storage, err := repo.convertTeamMetricsToStorage(metrics)
	if err != nil {
		return fmt.Errorf("failed to convert team metrics to storage: %w", err)
	}

	return repo.saveAggregatedMetrics(ctx, storage)
}

// SaveDeveloperMetrics は開発者メトリクスを保存
func (repo *AggregatedMetricsRepository) SaveDeveloperMetrics(ctx context.Context, metrics *analyticsApp.DeveloperMetrics) error {
	storage, err := repo.convertDeveloperMetricsToStorage(metrics)
	if err != nil {
		return fmt.Errorf("failed to convert developer metrics to storage: %w", err)
	}

	return repo.saveAggregatedMetrics(ctx, storage)
}

// SaveRepositoryMetrics はリポジトリメトリクスを保存
func (repo *AggregatedMetricsRepository) SaveRepositoryMetrics(ctx context.Context, metrics *analyticsApp.RepositoryMetrics) error {
	storage, err := repo.convertRepositoryMetricsToStorage(metrics)
	if err != nil {
		return fmt.Errorf("failed to convert repository metrics to storage: %w", err)
	}

	return repo.saveAggregatedMetrics(ctx, storage)
}

// FindTeamMetrics はチームメトリクスを取得
func (repo *AggregatedMetricsRepository) FindTeamMetrics(ctx context.Context, period analyticsApp.AggregationPeriod, startDate, endDate time.Time) ([]*analyticsApp.TeamMetrics, error) {
	storageList, err := repo.findAggregatedMetrics(ctx, "team", string(period), "", startDate, endDate)
	if err != nil {
		return nil, err
	}

	var metricsList []*analyticsApp.TeamMetrics
	for _, storage := range storageList {
		metrics, err := repo.convertStorageToTeamMetrics(storage)
		if err != nil {
			return nil, fmt.Errorf("failed to convert storage to team metrics: %w", err)
		}
		metricsList = append(metricsList, metrics)
	}

	return metricsList, nil
}

// FindDeveloperMetrics は開発者メトリクスを取得
func (repo *AggregatedMetricsRepository) FindDeveloperMetrics(ctx context.Context, developer string, period analyticsApp.AggregationPeriod, startDate, endDate time.Time) ([]*analyticsApp.DeveloperMetrics, error) {
	storageList, err := repo.findAggregatedMetrics(ctx, "developer", string(period), developer, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var metricsList []*analyticsApp.DeveloperMetrics
	for _, storage := range storageList {
		metrics, err := repo.convertStorageToDeveloperMetrics(storage)
		if err != nil {
			return nil, fmt.Errorf("failed to convert storage to developer metrics: %w", err)
		}
		metricsList = append(metricsList, metrics)
	}

	return metricsList, nil
}

// FindRepositoryMetrics はリポジトリメトリクスを取得
func (repo *AggregatedMetricsRepository) FindRepositoryMetrics(ctx context.Context, repository string, period analyticsApp.AggregationPeriod, startDate, endDate time.Time) ([]*analyticsApp.RepositoryMetrics, error) {
	storageList, err := repo.findAggregatedMetrics(ctx, "repository", string(period), repository, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var metricsList []*analyticsApp.RepositoryMetrics
	for _, storage := range storageList {
		metrics, err := repo.convertStorageToRepositoryMetrics(storage)
		if err != nil {
			return nil, fmt.Errorf("failed to convert storage to repository metrics: %w", err)
		}
		metricsList = append(metricsList, metrics)
	}

	return metricsList, nil
}

// FindAllDeveloperMetrics は全開発者のメトリクスを取得
func (repo *AggregatedMetricsRepository) FindAllDeveloperMetrics(ctx context.Context, period analyticsApp.AggregationPeriod, startDate, endDate time.Time) (map[string]*analyticsApp.DeveloperMetrics, error) {
	storageList, err := repo.findAggregatedMetrics(ctx, "developer", string(period), "", startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*analyticsApp.DeveloperMetrics)
	for _, storage := range storageList {
		metrics, err := repo.convertStorageToDeveloperMetrics(storage)
		if err != nil {
			return nil, fmt.Errorf("failed to convert storage to developer metrics: %w", err)
		}
		result[metrics.Developer] = metrics
	}

	return result, nil
}

// FindAllRepositoryMetrics は全リポジトリのメトリクスを取得
func (repo *AggregatedMetricsRepository) FindAllRepositoryMetrics(ctx context.Context, period analyticsApp.AggregationPeriod, startDate, endDate time.Time) (map[string]*analyticsApp.RepositoryMetrics, error) {
	storageList, err := repo.findAggregatedMetrics(ctx, "repository", string(period), "", startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*analyticsApp.RepositoryMetrics)
	for _, storage := range storageList {
		metrics, err := repo.convertStorageToRepositoryMetrics(storage)
		if err != nil {
			return nil, fmt.Errorf("failed to convert storage to repository metrics: %w", err)
		}
		result[metrics.Repository] = metrics
	}

	return result, nil
}

// FindLatestMetrics は最新の集計メトリクスを取得
func (repo *AggregatedMetricsRepository) FindLatestMetrics(ctx context.Context, aggregationLevel, targetID string, period analyticsApp.AggregationPeriod) (*analytics.AggregatedMetricsStorage, error) {
	query := `
		SELECT id, aggregation_level, aggregation_period, target_id, target_name,
			   period_start, period_end, total_prs, merged_prs, closed_prs,
			   avg_cycle_time_seconds, median_cycle_time_seconds, p95_cycle_time_seconds,
			   avg_review_time_seconds, median_review_time_seconds, p95_review_time_seconds,
			   avg_approval_time_seconds, median_approval_time_seconds, p95_approval_time_seconds,
			   avg_lines_changed, median_lines_changed, avg_files_changed, median_files_changed,
			   avg_review_comments, avg_review_rounds, first_pass_rate,
			   avg_complexity_score, median_complexity_score,
			   prs_per_day, lines_per_day, throughput,
			   cycle_time_trend, review_time_trend, quality_trend,
			   generated_at, updated_at, version, detailed_stats_json,
			   year_month, week_of_year, day_of_year
		FROM aggregated_metrics
		WHERE aggregation_level = $1 AND aggregation_period = $2
	`

	args := []interface{}{aggregationLevel, string(period)}
	argIndex := 3

	if targetID != "" {
		query += fmt.Sprintf(" AND target_id = $%d", argIndex)
		args = append(args, targetID)
		argIndex++
	}

	query += " ORDER BY generated_at DESC LIMIT 1"

	var storage analytics.AggregatedMetricsStorage
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(
		&storage.ID, &storage.AggregationLevel, &storage.AggregationPeriod,
		&storage.TargetID, &storage.TargetName, &storage.PeriodStart, &storage.PeriodEnd,
		&storage.TotalPRs, &storage.MergedPRs, &storage.ClosedPRs,
		&storage.AvgCycleTimeSeconds, &storage.MedianCycleTimeSeconds, &storage.P95CycleTimeSeconds,
		&storage.AvgReviewTimeSeconds, &storage.MedianReviewTimeSeconds, &storage.P95ReviewTimeSeconds,
		&storage.AvgApprovalTimeSeconds, &storage.MedianApprovalTimeSeconds, &storage.P95ApprovalTimeSeconds,
		&storage.AvgLinesChanged, &storage.MedianLinesChanged, &storage.AvgFilesChanged, &storage.MedianFilesChanged,
		&storage.AvgReviewComments, &storage.AvgReviewRounds, &storage.FirstPassRate,
		&storage.AvgComplexityScore, &storage.MedianComplexityScore,
		&storage.PRsPerDay, &storage.LinesPerDay, &storage.Throughput,
		&storage.CycleTimeTrend, &storage.ReviewTimeTrend, &storage.QualityTrend,
		&storage.GeneratedAt, &storage.UpdatedAt, &storage.Version, &storage.DetailedStatsJSON,
		&storage.YearMonth, &storage.WeekOfYear, &storage.DayOfYear,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find latest metrics: %w", err)
	}

	return &storage, nil
}

// DeleteOldAggregatedData は古い集計データを削除
func (repo *AggregatedMetricsRepository) DeleteOldAggregatedData(ctx context.Context, retentionPolicy analytics.DataRetentionPolicy) (int64, error) {
	totalDeleted := int64(0)

	// 日次集計データの削除
	if retentionPolicy.DailyAggregationRetentionDays > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -retentionPolicy.DailyAggregationRetentionDays)
		deleted, err := repo.deleteOldAggregatedDataByPeriod(ctx, "daily", cutoffDate)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to delete old daily aggregated data: %w", err)
		}
		totalDeleted += deleted
	}

	// 週次集計データの削除
	if retentionPolicy.WeeklyAggregationRetentionDays > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -retentionPolicy.WeeklyAggregationRetentionDays)
		deleted, err := repo.deleteOldAggregatedDataByPeriod(ctx, "weekly", cutoffDate)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to delete old weekly aggregated data: %w", err)
		}
		totalDeleted += deleted
	}

	// 月次集計データの削除
	if retentionPolicy.MonthlyAggregationRetentionDays > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -retentionPolicy.MonthlyAggregationRetentionDays)
		deleted, err := repo.deleteOldAggregatedDataByPeriod(ctx, "monthly", cutoffDate)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to delete old monthly aggregated data: %w", err)
		}
		totalDeleted += deleted
	}

	return totalDeleted, nil
}

// GetAggregatedStatistics は集計データの統計情報を取得
func (repo *AggregatedMetricsRepository) GetAggregatedStatistics(ctx context.Context) (*AggregatedRepositoryStatistics, error) {
	query := `
		SELECT 
			aggregation_level,
			aggregation_period,
			COUNT(*) as record_count,
			MIN(period_start) as oldest_period,
			MAX(period_end) as newest_period,
			COUNT(DISTINCT target_id) as unique_targets
		FROM aggregated_metrics
		GROUP BY aggregation_level, aggregation_period
		ORDER BY aggregation_level, aggregation_period
	`

	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregated statistics: %w", err)
	}
	defer rows.Close()

	stats := &AggregatedRepositoryStatistics{
		LevelStats: make(map[string]*AggregationLevelStats),
	}

	for rows.Next() {
		var level, period string
		var recordCount, uniqueTargets int64
		var oldestPeriod, newestPeriod time.Time

		err := rows.Scan(&level, &period, &recordCount, &oldestPeriod, &newestPeriod, &uniqueTargets)
		if err != nil {
			return nil, fmt.Errorf("failed to scan aggregated statistics row: %w", err)
		}

		if stats.LevelStats[level] == nil {
			stats.LevelStats[level] = &AggregationLevelStats{
				PeriodStats: make(map[string]*PeriodStats),
			}
		}

		stats.LevelStats[level].PeriodStats[period] = &PeriodStats{
			RecordCount:   recordCount,
			UniqueTargets: uniqueTargets,
			OldestPeriod:  oldestPeriod,
			NewestPeriod:  newestPeriod,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate aggregated statistics rows: %w", err)
	}

	return stats, nil
}

// AggregatedRepositoryStatistics は集計データリポジトリの統計情報
type AggregatedRepositoryStatistics struct {
	LevelStats map[string]*AggregationLevelStats `json:"levelStats"`
}

// AggregationLevelStats は集計レベル別の統計情報
type AggregationLevelStats struct {
	PeriodStats map[string]*PeriodStats `json:"periodStats"`
}

// PeriodStats は期間別の統計情報
type PeriodStats struct {
	RecordCount   int64     `json:"recordCount"`
	UniqueTargets int64     `json:"uniqueTargets"`
	OldestPeriod  time.Time `json:"oldestPeriod"`
	NewestPeriod  time.Time `json:"newestPeriod"`
}

// プライベートメソッド

func (repo *AggregatedMetricsRepository) saveAggregatedMetrics(ctx context.Context, storage *analytics.AggregatedMetricsStorage) error {
	query := `
		INSERT INTO aggregated_metrics (
			id, aggregation_level, aggregation_period, target_id, target_name,
			period_start, period_end, total_prs, merged_prs, closed_prs,
			avg_cycle_time_seconds, median_cycle_time_seconds, p95_cycle_time_seconds,
			avg_review_time_seconds, median_review_time_seconds, p95_review_time_seconds,
			avg_approval_time_seconds, median_approval_time_seconds, p95_approval_time_seconds,
			avg_lines_changed, median_lines_changed, avg_files_changed, median_files_changed,
			avg_review_comments, avg_review_rounds, first_pass_rate,
			avg_complexity_score, median_complexity_score,
			prs_per_day, lines_per_day, throughput,
			cycle_time_trend, review_time_trend, quality_trend,
			generated_at, updated_at, version, detailed_stats_json,
			year_month, week_of_year, day_of_year
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33, $34, $35, $36, $37, $38, $39, $40
		) ON CONFLICT (aggregation_level, aggregation_period, target_id, period_start, period_end) 
		DO UPDATE SET
			target_name = EXCLUDED.target_name,
			total_prs = EXCLUDED.total_prs,
			merged_prs = EXCLUDED.merged_prs,
			closed_prs = EXCLUDED.closed_prs,
			avg_cycle_time_seconds = EXCLUDED.avg_cycle_time_seconds,
			median_cycle_time_seconds = EXCLUDED.median_cycle_time_seconds,
			p95_cycle_time_seconds = EXCLUDED.p95_cycle_time_seconds,
			avg_review_time_seconds = EXCLUDED.avg_review_time_seconds,
			median_review_time_seconds = EXCLUDED.median_review_time_seconds,
			p95_review_time_seconds = EXCLUDED.p95_review_time_seconds,
			avg_approval_time_seconds = EXCLUDED.avg_approval_time_seconds,
			median_approval_time_seconds = EXCLUDED.median_approval_time_seconds,
			p95_approval_time_seconds = EXCLUDED.p95_approval_time_seconds,
			avg_lines_changed = EXCLUDED.avg_lines_changed,
			median_lines_changed = EXCLUDED.median_lines_changed,
			avg_files_changed = EXCLUDED.avg_files_changed,
			median_files_changed = EXCLUDED.median_files_changed,
			avg_review_comments = EXCLUDED.avg_review_comments,
			avg_review_rounds = EXCLUDED.avg_review_rounds,
			first_pass_rate = EXCLUDED.first_pass_rate,
			avg_complexity_score = EXCLUDED.avg_complexity_score,
			median_complexity_score = EXCLUDED.median_complexity_score,
			prs_per_day = EXCLUDED.prs_per_day,
			lines_per_day = EXCLUDED.lines_per_day,
			throughput = EXCLUDED.throughput,
			cycle_time_trend = EXCLUDED.cycle_time_trend,
			review_time_trend = EXCLUDED.review_time_trend,
			quality_trend = EXCLUDED.quality_trend,
			updated_at = EXCLUDED.updated_at,
			version = aggregated_metrics.version + 1,
			detailed_stats_json = EXCLUDED.detailed_stats_json,
			year_month = EXCLUDED.year_month,
			week_of_year = EXCLUDED.week_of_year,
			day_of_year = EXCLUDED.day_of_year
	`

	_, err := repo.db.ExecContext(ctx, query,
		storage.ID, storage.AggregationLevel, storage.AggregationPeriod,
		storage.TargetID, storage.TargetName, storage.PeriodStart, storage.PeriodEnd,
		storage.TotalPRs, storage.MergedPRs, storage.ClosedPRs,
		storage.AvgCycleTimeSeconds, storage.MedianCycleTimeSeconds, storage.P95CycleTimeSeconds,
		storage.AvgReviewTimeSeconds, storage.MedianReviewTimeSeconds, storage.P95ReviewTimeSeconds,
		storage.AvgApprovalTimeSeconds, storage.MedianApprovalTimeSeconds, storage.P95ApprovalTimeSeconds,
		storage.AvgLinesChanged, storage.MedianLinesChanged, storage.AvgFilesChanged, storage.MedianFilesChanged,
		storage.AvgReviewComments, storage.AvgReviewRounds, storage.FirstPassRate,
		storage.AvgComplexityScore, storage.MedianComplexityScore,
		storage.PRsPerDay, storage.LinesPerDay, storage.Throughput,
		storage.CycleTimeTrend, storage.ReviewTimeTrend, storage.QualityTrend,
		storage.GeneratedAt, storage.UpdatedAt, storage.Version, storage.DetailedStatsJSON,
		storage.YearMonth, storage.WeekOfYear, storage.DayOfYear,
	)

	return err
}

func (repo *AggregatedMetricsRepository) findAggregatedMetrics(ctx context.Context, aggregationLevel, period, targetID string, startDate, endDate time.Time) ([]*analytics.AggregatedMetricsStorage, error) {
	query := `
		SELECT id, aggregation_level, aggregation_period, target_id, target_name,
			   period_start, period_end, total_prs, merged_prs, closed_prs,
			   avg_cycle_time_seconds, median_cycle_time_seconds, p95_cycle_time_seconds,
			   avg_review_time_seconds, median_review_time_seconds, p95_review_time_seconds,
			   avg_approval_time_seconds, median_approval_time_seconds, p95_approval_time_seconds,
			   avg_lines_changed, median_lines_changed, avg_files_changed, median_files_changed,
			   avg_review_comments, avg_review_rounds, first_pass_rate,
			   avg_complexity_score, median_complexity_score,
			   prs_per_day, lines_per_day, throughput,
			   cycle_time_trend, review_time_trend, quality_trend,
			   generated_at, updated_at, version, detailed_stats_json,
			   year_month, week_of_year, day_of_year
		FROM aggregated_metrics
		WHERE aggregation_level = $1 AND aggregation_period = $2
		  AND period_start >= $3 AND period_end <= $4
	`

	args := []interface{}{aggregationLevel, period, startDate, endDate}
	argIndex := 5

	if targetID != "" {
		query += fmt.Sprintf(" AND target_id = $%d", argIndex)
		args = append(args, targetID)
		argIndex++
	}

	query += " ORDER BY period_start DESC"

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregated metrics: %w", err)
	}
	defer rows.Close()

	var storageList []*analytics.AggregatedMetricsStorage
	for rows.Next() {
		var storage analytics.AggregatedMetricsStorage
		err := rows.Scan(
			&storage.ID, &storage.AggregationLevel, &storage.AggregationPeriod,
			&storage.TargetID, &storage.TargetName, &storage.PeriodStart, &storage.PeriodEnd,
			&storage.TotalPRs, &storage.MergedPRs, &storage.ClosedPRs,
			&storage.AvgCycleTimeSeconds, &storage.MedianCycleTimeSeconds, &storage.P95CycleTimeSeconds,
			&storage.AvgReviewTimeSeconds, &storage.MedianReviewTimeSeconds, &storage.P95ReviewTimeSeconds,
			&storage.AvgApprovalTimeSeconds, &storage.MedianApprovalTimeSeconds, &storage.P95ApprovalTimeSeconds,
			&storage.AvgLinesChanged, &storage.MedianLinesChanged, &storage.AvgFilesChanged, &storage.MedianFilesChanged,
			&storage.AvgReviewComments, &storage.AvgReviewRounds, &storage.FirstPassRate,
			&storage.AvgComplexityScore, &storage.MedianComplexityScore,
			&storage.PRsPerDay, &storage.LinesPerDay, &storage.Throughput,
			&storage.CycleTimeTrend, &storage.ReviewTimeTrend, &storage.QualityTrend,
			&storage.GeneratedAt, &storage.UpdatedAt, &storage.Version, &storage.DetailedStatsJSON,
			&storage.YearMonth, &storage.WeekOfYear, &storage.DayOfYear,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan aggregated metrics row: %w", err)
		}

		storageList = append(storageList, &storage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate aggregated metrics rows: %w", err)
	}

	return storageList, nil
}

func (repo *AggregatedMetricsRepository) deleteOldAggregatedDataByPeriod(ctx context.Context, period string, cutoffDate time.Time) (int64, error) {
	query := `DELETE FROM aggregated_metrics WHERE aggregation_period = $1 AND generated_at < $2`
	result, err := repo.db.ExecContext(ctx, query, period, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old aggregated data for period %s: %w", period, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// 変換メソッド群

func (repo *AggregatedMetricsRepository) convertTeamMetricsToStorage(metrics *analyticsApp.TeamMetrics) (*analytics.AggregatedMetricsStorage, error) {
	detailedStatsJSON, err := json.Marshal(map[string]interface{}{
		"cycleTimeStats":  metrics.CycleTimeStats,
		"reviewStats":     metrics.ReviewStats,
		"sizeStats":       metrics.SizeStats,
		"qualityStats":    metrics.QualityStats,
		"complexityStats": metrics.ComplexityStats,
		"trendAnalysis":   metrics.TrendAnalysis,
		"bottlenecks":     metrics.Bottlenecks,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal detailed stats: %w", err)
	}

	// 期間フィールドの生成
	yearMonth := metrics.DateRange.Start.Format("2006-01")
	year, week := metrics.DateRange.Start.ISOWeek()
	weekOfYear := fmt.Sprintf("%d-W%02d", year, week)
	dayOfYear := metrics.DateRange.Start.Format("2006-002")

	return &analytics.AggregatedMetricsStorage{
		ID:               fmt.Sprintf("team_%s_%d", string(metrics.Period), time.Now().Unix()),
		AggregationLevel: "team",
		AggregationPeriod: string(metrics.Period),
		TargetID:         "team",
		TargetName:       "Team",
		PeriodStart:      metrics.DateRange.Start,
		PeriodEnd:        metrics.DateRange.End,
		TotalPRs:         metrics.TotalPRs,
		MergedPRs:        metrics.TotalPRs, // 仮定: 集計対象は全てマージ済み
		ClosedPRs:        0,
		
		// サイクルタイム統計
		AvgCycleTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TotalCycleTime.Mean),
		MedianCycleTimeSeconds: repo.durationToSecondsPtr(metrics.CycleTimeStats.TotalCycleTime.Median),
		P95CycleTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TotalCycleTime.Percentiles.P95),
		
		// レビュー時間統計
		AvgReviewTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToFirstReview.Mean),
		MedianReviewTimeSeconds: repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToFirstReview.Median),
		P95ReviewTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToFirstReview.Percentiles.P95),
		
		// 承認時間統計
		AvgApprovalTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToApproval.Mean),
		MedianApprovalTimeSeconds: repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToApproval.Median),
		P95ApprovalTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToApproval.Percentiles.P95),
		
		// サイズ統計
		AvgLinesChanged:    metrics.SizeStats.LinesChanged.Mean,
		MedianLinesChanged: metrics.SizeStats.LinesChanged.Median,
		AvgFilesChanged:    metrics.SizeStats.FilesChanged.Mean,
		MedianFilesChanged: metrics.SizeStats.FilesChanged.Median,
		
		// 品質統計
		AvgReviewComments: metrics.ReviewStats.CommentCount.Mean,
		AvgReviewRounds:   metrics.ReviewStats.RoundCount.Mean,
		FirstPassRate:     metrics.ReviewStats.FirstReviewPassRate.Mean,
		
		// 複雑度統計
		AvgComplexityScore:    metrics.ComplexityStats.ComplexityScore.Mean,
		MedianComplexityScore: metrics.ComplexityStats.ComplexityScore.Median,
		
		// 生産性指標（チームでは計算しない）
		PRsPerDay:   float64(metrics.TotalPRs) / repo.calculateDays(metrics.DateRange),
		LinesPerDay: float64(metrics.SizeStats.LinesChanged.Sum) / repo.calculateDays(metrics.DateRange),
		Throughput:  float64(metrics.TotalPRs),
		
		// トレンド情報
		CycleTimeTrend:  metrics.TrendAnalysis.CycleTimeTrend.Trend,
		ReviewTimeTrend: metrics.TrendAnalysis.ReviewTimeTrend.Trend,
		QualityTrend:    metrics.TrendAnalysis.QualityTrend.Trend,
		
		// メタデータ
		GeneratedAt: metrics.GeneratedAt,
		UpdatedAt:   time.Now(),
		Version:     1,
		
		DetailedStatsJSON: string(detailedStatsJSON),
		YearMonth:         yearMonth,
		WeekOfYear:        weekOfYear,
		DayOfYear:         dayOfYear,
	}, nil
}

func (repo *AggregatedMetricsRepository) convertDeveloperMetricsToStorage(metrics *analyticsApp.DeveloperMetrics) (*analytics.AggregatedMetricsStorage, error) {
	detailedStatsJSON, err := json.Marshal(map[string]interface{}{
		"cycleTimeStats":  metrics.CycleTimeStats,
		"reviewStats":     metrics.ReviewStats,
		"sizeStats":       metrics.SizeStats,
		"qualityStats":    metrics.QualityStats,
		"complexityStats": metrics.ComplexityStats,
		"productivity":    metrics.Productivity,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal detailed stats: %w", err)
	}

	// 期間フィールドの生成
	yearMonth := metrics.DateRange.Start.Format("2006-01")
	year, week := metrics.DateRange.Start.ISOWeek()
	weekOfYear := fmt.Sprintf("%d-W%02d", year, week)
	dayOfYear := metrics.DateRange.Start.Format("2006-002")

	return &analytics.AggregatedMetricsStorage{
		ID:               fmt.Sprintf("dev_%s_%s_%d", metrics.Developer, string(metrics.Period), time.Now().Unix()),
		AggregationLevel: "developer",
		AggregationPeriod: string(metrics.Period),
		TargetID:         metrics.Developer,
		TargetName:       metrics.Developer,
		PeriodStart:      metrics.DateRange.Start,
		PeriodEnd:        metrics.DateRange.End,
		TotalPRs:         metrics.TotalPRs,
		MergedPRs:        metrics.TotalPRs,
		ClosedPRs:        0,
		
		// 各統計値の設定（チームメトリクスと同様）
		AvgCycleTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TotalCycleTime.Mean),
		MedianCycleTimeSeconds: repo.durationToSecondsPtr(metrics.CycleTimeStats.TotalCycleTime.Median),
		P95CycleTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TotalCycleTime.Percentiles.P95),
		
		AvgReviewTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToFirstReview.Mean),
		MedianReviewTimeSeconds: repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToFirstReview.Median),
		P95ReviewTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToFirstReview.Percentiles.P95),
		
		AvgApprovalTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToApproval.Mean),
		MedianApprovalTimeSeconds: repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToApproval.Median),
		P95ApprovalTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToApproval.Percentiles.P95),
		
		AvgLinesChanged:    metrics.SizeStats.LinesChanged.Mean,
		MedianLinesChanged: metrics.SizeStats.LinesChanged.Median,
		AvgFilesChanged:    metrics.SizeStats.FilesChanged.Mean,
		MedianFilesChanged: metrics.SizeStats.FilesChanged.Median,
		
		AvgReviewComments: metrics.ReviewStats.CommentCount.Mean,
		AvgReviewRounds:   metrics.ReviewStats.RoundCount.Mean,
		FirstPassRate:     metrics.ReviewStats.FirstReviewPassRate.Mean,
		
		AvgComplexityScore:    metrics.ComplexityStats.ComplexityScore.Mean,
		MedianComplexityScore: metrics.ComplexityStats.ComplexityScore.Median,
		
		// 生産性指標
		PRsPerDay:   metrics.Productivity.PRsPerDay,
		LinesPerDay: metrics.Productivity.LinesPerDay,
		Throughput:  metrics.Productivity.Throughput,
		
		// トレンド情報（開発者メトリクスには含まれないため空）
		CycleTimeTrend:  "stable",
		ReviewTimeTrend: "stable",
		QualityTrend:    "stable",
		
		// メタデータ
		GeneratedAt: metrics.GeneratedAt,
		UpdatedAt:   time.Now(),
		Version:     1,
		
		DetailedStatsJSON: string(detailedStatsJSON),
		YearMonth:         yearMonth,
		WeekOfYear:        weekOfYear,
		DayOfYear:         dayOfYear,
	}, nil
}

func (repo *AggregatedMetricsRepository) convertRepositoryMetricsToStorage(metrics *analyticsApp.RepositoryMetrics) (*analytics.AggregatedMetricsStorage, error) {
	detailedStatsJSON, err := json.Marshal(map[string]interface{}{
		"cycleTimeStats":  metrics.CycleTimeStats,
		"reviewStats":     metrics.ReviewStats,
		"sizeStats":       metrics.SizeStats,
		"qualityStats":    metrics.QualityStats,
		"complexityStats": metrics.ComplexityStats,
		"contributors":    metrics.Contributors,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal detailed stats: %w", err)
	}

	// 期間フィールドの生成
	yearMonth := metrics.DateRange.Start.Format("2006-01")
	year, week := metrics.DateRange.Start.ISOWeek()
	weekOfYear := fmt.Sprintf("%d-W%02d", year, week)
	dayOfYear := metrics.DateRange.Start.Format("2006-002")

	return &analytics.AggregatedMetricsStorage{
		ID:               fmt.Sprintf("repo_%s_%s_%d", metrics.Repository, string(metrics.Period), time.Now().Unix()),
		AggregationLevel: "repository",
		AggregationPeriod: string(metrics.Period),
		TargetID:         metrics.Repository,
		TargetName:       metrics.Repository,
		PeriodStart:      metrics.DateRange.Start,
		PeriodEnd:        metrics.DateRange.End,
		TotalPRs:         metrics.TotalPRs,
		MergedPRs:        metrics.TotalPRs,
		ClosedPRs:        0,
		
		// 各統計値の設定（チームメトリクスと同様）
		AvgCycleTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TotalCycleTime.Mean),
		MedianCycleTimeSeconds: repo.durationToSecondsPtr(metrics.CycleTimeStats.TotalCycleTime.Median),
		P95CycleTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TotalCycleTime.Percentiles.P95),
		
		AvgReviewTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToFirstReview.Mean),
		MedianReviewTimeSeconds: repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToFirstReview.Median),
		P95ReviewTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToFirstReview.Percentiles.P95),
		
		AvgApprovalTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToApproval.Mean),
		MedianApprovalTimeSeconds: repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToApproval.Median),
		P95ApprovalTimeSeconds:    repo.durationToSecondsPtr(metrics.CycleTimeStats.TimeToApproval.Percentiles.P95),
		
		AvgLinesChanged:    metrics.SizeStats.LinesChanged.Mean,
		MedianLinesChanged: metrics.SizeStats.LinesChanged.Median,
		AvgFilesChanged:    metrics.SizeStats.FilesChanged.Mean,
		MedianFilesChanged: metrics.SizeStats.FilesChanged.Median,
		
		AvgReviewComments: metrics.ReviewStats.CommentCount.Mean,
		AvgReviewRounds:   metrics.ReviewStats.RoundCount.Mean,
		FirstPassRate:     metrics.ReviewStats.FirstReviewPassRate.Mean,
		
		AvgComplexityScore:    metrics.ComplexityStats.ComplexityScore.Mean,
		MedianComplexityScore: metrics.ComplexityStats.ComplexityScore.Median,
		
		// 生産性指標
		PRsPerDay:   float64(metrics.TotalPRs) / repo.calculateDays(metrics.DateRange),
		LinesPerDay: float64(metrics.SizeStats.LinesChanged.Sum) / repo.calculateDays(metrics.DateRange),
		Throughput:  float64(metrics.TotalPRs),
		
		// トレンド情報（リポジトリメトリクスには含まれないため空）
		CycleTimeTrend:  "stable",
		ReviewTimeTrend: "stable",
		QualityTrend:    "stable",
		
		// メタデータ
		GeneratedAt: metrics.GeneratedAt,
		UpdatedAt:   time.Now(),
		Version:     1,
		
		DetailedStatsJSON: string(detailedStatsJSON),
		YearMonth:         yearMonth,
		WeekOfYear:        weekOfYear,
		DayOfYear:         dayOfYear,
	}, nil
}

func (repo *AggregatedMetricsRepository) convertStorageToTeamMetrics(storage *analytics.AggregatedMetricsStorage) (*analyticsApp.TeamMetrics, error) {
	// 詳細統計の復元
	var detailedStats map[string]interface{}
	if err := json.Unmarshal([]byte(storage.DetailedStatsJSON), &detailedStats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal detailed stats: %w", err)
	}

	return &analyticsApp.TeamMetrics{
		Period:      analyticsApp.AggregationPeriod(storage.AggregationPeriod),
		TotalPRs:    storage.TotalPRs,
		DateRange:   analyticsApp.DateRange{Start: storage.PeriodStart, End: storage.PeriodEnd},
		GeneratedAt: storage.GeneratedAt,
		// 詳細統計は必要に応じて復元
	}, nil
}

func (repo *AggregatedMetricsRepository) convertStorageToDeveloperMetrics(storage *analytics.AggregatedMetricsStorage) (*analyticsApp.DeveloperMetrics, error) {
	return &analyticsApp.DeveloperMetrics{
		Developer:   storage.TargetID,
		Period:      analyticsApp.AggregationPeriod(storage.AggregationPeriod),
		TotalPRs:    storage.TotalPRs,
		DateRange:   analyticsApp.DateRange{Start: storage.PeriodStart, End: storage.PeriodEnd},
		GeneratedAt: storage.GeneratedAt,
		Productivity: analyticsApp.ProductivityMetrics{
			PRsPerDay:   storage.PRsPerDay,
			LinesPerDay: storage.LinesPerDay,
			Throughput:  storage.Throughput,
		},
	}, nil
}

func (repo *AggregatedMetricsRepository) convertStorageToRepositoryMetrics(storage *analytics.AggregatedMetricsStorage) (*analyticsApp.RepositoryMetrics, error) {
	return &analyticsApp.RepositoryMetrics{
		Repository:  storage.TargetID,
		Period:      analyticsApp.AggregationPeriod(storage.AggregationPeriod),
		TotalPRs:    storage.TotalPRs,
		DateRange:   analyticsApp.DateRange{Start: storage.PeriodStart, End: storage.PeriodEnd},
		GeneratedAt: storage.GeneratedAt,
	}, nil
}

// ユーティリティメソッド

func (repo *AggregatedMetricsRepository) durationToSecondsPtr(d time.Duration) *int64 {
	if d == 0 {
		return nil
	}
	seconds := int64(d.Seconds())
	return &seconds
}

func (repo *AggregatedMetricsRepository) calculateDays(dateRange analyticsApp.DateRange) float64 {
	days := dateRange.End.Sub(dateRange.Start).Hours() / 24
	if days < 1 {
		return 1
	}
	return days
}