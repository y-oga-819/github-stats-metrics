package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	prDomain "github-stats-metrics/domain/pull_request"
)

func TestPRMetricsRepository_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPRMetricsRepository(db)
	
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	metrics := createTestPRMetricsForRepo(baseTime)

	// PRメトリクス保存のSQL期待値
	mock.ExpectExec(`INSERT INTO pr_metrics`).
		WithArgs(
			sqlmock.AnyArg(), metrics.PRID, metrics.PRNumber, metrics.Title, metrics.Author,
			metrics.Repository, metrics.CreatedAt, metrics.MergedAt, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), metrics.QualityMetrics.ReviewCommentCount,
			metrics.QualityMetrics.ReviewRoundCount, metrics.QualityMetrics.ReviewerCount,
			metrics.QualityMetrics.FirstReviewPassRate, sqlmock.AnyArg(),
			metrics.ComplexityScore, metrics.SizeCategory, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// ファイル変更削除のSQL期待値
	mock.ExpectExec(`DELETE FROM file_changes WHERE pr_metrics_id`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// ファイル変更挿入のSQL期待値
	for range metrics.SizeMetrics.FileChanges {
		mock.ExpectExec(`INSERT INTO file_changes`).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	err = repo.Save(context.Background(), metrics)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPRMetricsRepository_SaveBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPRMetricsRepository(db)
	
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	metricsList := []*prDomain.PRMetrics{
		createTestPRMetricsForRepo(baseTime),
		createTestPRMetricsForRepo(baseTime.Add(time.Hour)),
	}

	// トランザクション開始
	mock.ExpectBegin()

	// 各メトリクスに対するSQL期待値
	for i := 0; i < len(metricsList); i++ {
		// PRメトリクス挿入
		mock.ExpectExec(`INSERT INTO pr_metrics`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// ファイル変更挿入（各メトリクスに2ファイルずつあると仮定）
		for j := 0; j < 2; j++ {
			mock.ExpectExec(`INSERT INTO file_changes`).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 
					sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
					sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}
	}

	// トランザクションコミット
	mock.ExpectCommit()

	err = repo.SaveBatch(context.Background(), metricsList)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPRMetricsRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPRMetricsRepository(db)
	
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	expectedMetrics := createTestPRMetricsForRepo(baseTime)
	
	// ストレージモデルを準備
	storage, err := repo.convertToStorage(expectedMetrics)
	require.NoError(t, err)

	// クエリ期待値の設定
	rows := sqlmock.NewRows([]string{
		"id", "pr_id", "pr_number", "title", "author", "repository", "created_at", "merged_at", "collected_at",
		"size_metrics_json", "total_cycle_time_seconds", "time_to_first_review_seconds",
		"time_to_approval_seconds", "time_to_merge_seconds", "time_metrics_json",
		"review_comment_count", "review_round_count", "reviewer_count", "first_review_pass_rate",
		"quality_metrics_json", "complexity_score", "size_category",
		"year_month", "week_of_year", "day_of_year",
	}).AddRow(
		storage.ID, storage.PRID, storage.PRNumber, storage.Title, storage.Author,
		storage.Repository, storage.CreatedAt, storage.MergedAt, storage.CollectedAt,
		storage.SizeMetricsJSON, storage.TotalCycleTimeSeconds, storage.TimeToFirstReviewSeconds,
		storage.TimeToApprovalSeconds, storage.TimeToMergeSeconds, storage.TimeMetricsJSON,
		storage.ReviewCommentCount, storage.ReviewRoundCount, storage.ReviewerCount,
		storage.FirstReviewPassRate, storage.QualityMetricsJSON, storage.ComplexityScore,
		storage.SizeCategory, storage.YearMonth, storage.WeekOfYear, storage.DayOfYear,
	)

	mock.ExpectQuery(`SELECT .+ FROM pr_metrics WHERE id`).
		WithArgs("test-id").
		WillReturnRows(rows)

	result, err := repo.FindByID(context.Background(), "test-id")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMetrics.PRID, result.PRID)
	assert.Equal(t, expectedMetrics.PRNumber, result.PRNumber)
	assert.Equal(t, expectedMetrics.Title, result.Title)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPRMetricsRepository_FindByPRID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPRMetricsRepository(db)
	
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	expectedMetrics := createTestPRMetricsForRepo(baseTime)
	
	// ストレージモデルを準備
	storage, err := repo.convertToStorage(expectedMetrics)
	require.NoError(t, err)

	// クエリ期待値の設定
	rows := sqlmock.NewRows([]string{
		"id", "pr_id", "pr_number", "title", "author", "repository", "created_at", "merged_at", "collected_at",
		"size_metrics_json", "total_cycle_time_seconds", "time_to_first_review_seconds",
		"time_to_approval_seconds", "time_to_merge_seconds", "time_metrics_json",
		"review_comment_count", "review_round_count", "reviewer_count", "first_review_pass_rate",
		"quality_metrics_json", "complexity_score", "size_category",
		"year_month", "week_of_year", "day_of_year",
	}).AddRow(
		storage.ID, storage.PRID, storage.PRNumber, storage.Title, storage.Author,
		storage.Repository, storage.CreatedAt, storage.MergedAt, storage.CollectedAt,
		storage.SizeMetricsJSON, storage.TotalCycleTimeSeconds, storage.TimeToFirstReviewSeconds,
		storage.TimeToApprovalSeconds, storage.TimeToMergeSeconds, storage.TimeMetricsJSON,
		storage.ReviewCommentCount, storage.ReviewRoundCount, storage.ReviewerCount,
		storage.FirstReviewPassRate, storage.QualityMetricsJSON, storage.ComplexityScore,
		storage.SizeCategory, storage.YearMonth, storage.WeekOfYear, storage.DayOfYear,
	)

	mock.ExpectQuery(`SELECT .+ FROM pr_metrics WHERE pr_id`).
		WithArgs("pr-123").
		WillReturnRows(rows)

	result, err := repo.FindByPRID(context.Background(), "pr-123")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMetrics.PRID, result.PRID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPRMetricsRepository_FindByDateRange(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPRMetricsRepository(db)
	
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	developers := []string{"dev1", "dev2"}
	repositories := []string{"repo1"}

	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	expectedMetrics := createTestPRMetricsForRepo(baseTime)
	storage, err := repo.convertToStorage(expectedMetrics)
	require.NoError(t, err)

	// クエリ期待値の設定
	rows := sqlmock.NewRows([]string{
		"id", "pr_id", "pr_number", "title", "author", "repository", "created_at", "merged_at", "collected_at",
		"size_metrics_json", "total_cycle_time_seconds", "time_to_first_review_seconds",
		"time_to_approval_seconds", "time_to_merge_seconds", "time_metrics_json",
		"review_comment_count", "review_round_count", "reviewer_count", "first_review_pass_rate",
		"quality_metrics_json", "complexity_score", "size_category",
		"year_month", "week_of_year", "day_of_year",
	}).AddRow(
		storage.ID, storage.PRID, storage.PRNumber, storage.Title, storage.Author,
		storage.Repository, storage.CreatedAt, storage.MergedAt, storage.CollectedAt,
		storage.SizeMetricsJSON, storage.TotalCycleTimeSeconds, storage.TimeToFirstReviewSeconds,
		storage.TimeToApprovalSeconds, storage.TimeToMergeSeconds, storage.TimeMetricsJSON,
		storage.ReviewCommentCount, storage.ReviewRoundCount, storage.ReviewerCount,
		storage.FirstReviewPassRate, storage.QualityMetricsJSON, storage.ComplexityScore,
		storage.SizeCategory, storage.YearMonth, storage.WeekOfYear, storage.DayOfYear,
	)

	mock.ExpectQuery(`SELECT .+ FROM pr_metrics WHERE created_at >= .+ AND created_at <= .+ AND author = ANY.+ AND repository = ANY.+ ORDER BY created_at DESC`).
		WithArgs(startDate, endDate, developers, repositories).
		WillReturnRows(rows)

	result, err := repo.FindByDateRange(context.Background(), startDate, endDate, developers, repositories)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, expectedMetrics.PRID, result[0].PRID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPRMetricsRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPRMetricsRepository(db)
	
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	metrics := createTestPRMetricsForRepo(baseTime)

	mock.ExpectExec(`UPDATE pr_metrics SET`).
		WithArgs(
			sqlmock.AnyArg(), metrics.PRNumber, metrics.Title, metrics.Author,
			metrics.Repository, metrics.CreatedAt, metrics.MergedAt, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), metrics.QualityMetrics.ReviewCommentCount,
			metrics.QualityMetrics.ReviewRoundCount, metrics.QualityMetrics.ReviewerCount,
			metrics.QualityMetrics.FirstReviewPassRate, sqlmock.AnyArg(),
			metrics.ComplexityScore, metrics.SizeCategory, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(context.Background(), metrics)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPRMetricsRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPRMetricsRepository(db)

	// トランザクション開始
	mock.ExpectBegin()

	// 関連データの削除
	mock.ExpectExec(`DELETE FROM file_changes WHERE pr_metrics_id`).
		WithArgs("test-id").
		WillReturnResult(sqlmock.NewResult(0, 2))

	mock.ExpectExec(`DELETE FROM review_events WHERE pr_metrics_id`).
		WithArgs("test-id").
		WillReturnResult(sqlmock.NewResult(0, 5))

	// メインデータの削除
	mock.ExpectExec(`DELETE FROM pr_metrics WHERE id`).
		WithArgs("test-id").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// トランザクションコミット
	mock.ExpectCommit()

	err = repo.Delete(context.Background(), "test-id")
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPRMetricsRepository_DeleteOldData(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPRMetricsRepository(db)

	retentionDays := 30
	
	mock.ExpectExec(`DELETE FROM pr_metrics WHERE collected_at <`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 10))

	deletedCount, err := repo.DeleteOldData(context.Background(), retentionDays)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), deletedCount)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPRMetricsRepository_GetStatistics(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPRMetricsRepository(db)

	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	rows := sqlmock.NewRows([]string{
		"total_records", "unique_developers", "unique_repositories",
		"oldest_record", "newest_record", "avg_cycle_time", "avg_complexity",
	}).AddRow(
		100, 10, 5, baseTime, baseTime.Add(30*24*time.Hour), 86400.0, 2.5,
	)

	mock.ExpectQuery(`SELECT .+ FROM pr_metrics`).
		WillReturnRows(rows)

	stats, err := repo.GetStatistics(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(100), stats.TotalRecords)
	assert.Equal(t, int64(10), stats.UniqueDevelopers)
	assert.Equal(t, int64(5), stats.UniqueRepositories)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPRMetricsRepository_ConversionMethods(t *testing.T) {
	repo := &PRMetricsRepository{}
	
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	originalMetrics := createTestPRMetricsForRepo(baseTime)

	t.Run("convertToStorage", func(t *testing.T) {
		storage, err := repo.convertToStorage(originalMetrics)
		assert.NoError(t, err)
		assert.NotNil(t, storage)
		
		// 基本フィールドの確認
		assert.Equal(t, originalMetrics.PRID, storage.PRID)
		assert.Equal(t, originalMetrics.PRNumber, storage.PRNumber)
		assert.Equal(t, originalMetrics.Title, storage.Title)
		assert.Equal(t, originalMetrics.Author, storage.Author)
		assert.Equal(t, originalMetrics.Repository, storage.Repository)
		
		// JSON化されたフィールドの確認
		assert.NotEmpty(t, storage.SizeMetricsJSON)
		assert.NotEmpty(t, storage.TimeMetricsJSON)
		assert.NotEmpty(t, storage.QualityMetricsJSON)
		
		// 時間フィールドの確認（秒単位に変換されているか）
		assert.NotNil(t, storage.TotalCycleTimeSeconds)
		assert.NotNil(t, storage.TimeToFirstReviewSeconds)
		
		// インデックス用フィールドの確認
		assert.Equal(t, "2024-01", storage.YearMonth)
		assert.NotEmpty(t, storage.WeekOfYear)
		assert.NotEmpty(t, storage.DayOfYear)
	})

	t.Run("convertFromStorage", func(t *testing.T) {
		// まず、ストレージモデルに変換
		storage, err := repo.convertToStorage(originalMetrics)
		require.NoError(t, err)
		
		// ストレージモデルからドメインモデルに戻す
		convertedMetrics, err := repo.convertFromStorage(storage)
		assert.NoError(t, err)
		assert.NotNil(t, convertedMetrics)
		
		// 基本フィールドの確認
		assert.Equal(t, originalMetrics.PRID, convertedMetrics.PRID)
		assert.Equal(t, originalMetrics.PRNumber, convertedMetrics.PRNumber)
		assert.Equal(t, originalMetrics.Title, convertedMetrics.Title)
		assert.Equal(t, originalMetrics.Author, convertedMetrics.Author)
		assert.Equal(t, originalMetrics.Repository, convertedMetrics.Repository)
		
		// 時間フィールドの確認
		assert.Equal(t, originalMetrics.CreatedAt.Unix(), convertedMetrics.CreatedAt.Unix())
		
		// サイズメトリクスの確認
		assert.Equal(t, originalMetrics.SizeMetrics.LinesAdded, convertedMetrics.SizeMetrics.LinesAdded)
		assert.Equal(t, originalMetrics.SizeMetrics.LinesDeleted, convertedMetrics.SizeMetrics.LinesDeleted)
		assert.Equal(t, originalMetrics.SizeMetrics.FilesChanged, convertedMetrics.SizeMetrics.FilesChanged)
		
		// 品質メトリクスの確認
		assert.Equal(t, originalMetrics.QualityMetrics.ReviewCommentCount, convertedMetrics.QualityMetrics.ReviewCommentCount)
		assert.Equal(t, originalMetrics.QualityMetrics.ReviewRoundCount, convertedMetrics.QualityMetrics.ReviewRoundCount)
		
		// 複雑度とサイズカテゴリの確認
		assert.Equal(t, originalMetrics.ComplexityScore, convertedMetrics.ComplexityScore)
		assert.Equal(t, originalMetrics.SizeCategory, convertedMetrics.SizeCategory)
	})
}

func TestPRMetricsRepository_EdgeCases(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPRMetricsRepository(db)

	t.Run("FindByID_NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM pr_metrics WHERE id`).
			WithArgs("non-existent-id").
			WillReturnError(sql.ErrNoRows)

		result, err := repo.FindByID(context.Background(), "non-existent-id")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("FindByPRID_NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM pr_metrics WHERE pr_id`).
			WithArgs("non-existent-pr-id").
			WillReturnError(sql.ErrNoRows)

		result, err := repo.FindByPRID(context.Background(), "non-existent-pr-id")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("FindByDateRange_EmptyResult", func(t *testing.T) {
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		rows := sqlmock.NewRows([]string{
			"id", "pr_id", "pr_number", "title", "author", "repository", "created_at", "merged_at", "collected_at",
			"size_metrics_json", "total_cycle_time_seconds", "time_to_first_review_seconds",
			"time_to_approval_seconds", "time_to_merge_seconds", "time_metrics_json",
			"review_comment_count", "review_round_count", "reviewer_count", "first_review_pass_rate",
			"quality_metrics_json", "complexity_score", "size_category",
			"year_month", "week_of_year", "day_of_year",
		})

		mock.ExpectQuery(`SELECT .+ FROM pr_metrics WHERE created_at >= .+ AND created_at <= .+ ORDER BY created_at DESC`).
			WithArgs(startDate, endDate).
			WillReturnRows(rows)

		result, err := repo.FindByDateRange(context.Background(), startDate, endDate, nil, nil)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("SaveBatch_EmptyList", func(t *testing.T) {
		err := repo.SaveBatch(context.Background(), []*prDomain.PRMetrics{})
		assert.NoError(t, err)
	})

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestPRMetricsRepository_TransactionFailures(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPRMetricsRepository(db)

	t.Run("SaveBatch_TransactionBeginFailure", func(t *testing.T) {
		baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
		metricsList := []*prDomain.PRMetrics{
			createTestPRMetricsForRepo(baseTime),
		}

		mock.ExpectBegin().WillReturnError(fmt.Errorf("begin transaction failed"))

		err := repo.SaveBatch(context.Background(), metricsList)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to begin transaction")
	})

	t.Run("Delete_TransactionBeginFailure", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(fmt.Errorf("begin transaction failed"))

		err := repo.Delete(context.Background(), "test-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to begin transaction")
	})

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// ヘルパー関数

func createTestPRMetricsForRepo(baseTime time.Time) *prDomain.PRMetrics {
	mergedTime := baseTime.Add(24 * time.Hour)
	
	return &prDomain.PRMetrics{
		PRID:       "pr-123",
		PRNumber:   123,
		Title:      "Test PR",
		Author:     "test-user",
		Repository: "test-repo",
		CreatedAt:  baseTime,
		MergedAt:   &mergedTime,
		SizeMetrics: prDomain.PRSizeMetrics{
			LinesAdded:   100,
			LinesDeleted: 50,
			LinesChanged: 150,
			FilesChanged: 5,
			FileTypeBreakdown: map[string]int{
				".go": 3,
				".js": 2,
			},
			DirectoryCount: 2,
			FileChanges: []prDomain.FileChangeMetrics{
				{
					FileName:     "main.go",
					FileType:     ".go",
					LinesAdded:   50,
					LinesDeleted: 10,
					IsNewFile:    false,
					IsDeleted:    false,
					IsRenamed:    false,
				},
				{
					FileName:     "app.js",
					FileType:     ".js",
					LinesAdded:   30,
					LinesDeleted: 20,
					IsNewFile:    false,
					IsDeleted:    false,
					IsRenamed:    false,
				},
			},
		},
		TimeMetrics: prDomain.PRTimeMetrics{
			TotalCycleTime:    durationPtrRepo(24 * time.Hour),
			TimeToFirstReview: durationPtrRepo(2 * time.Hour),
			TimeToApproval:    durationPtrRepo(4 * time.Hour),
			TimeToMerge:       durationPtrRepo(1 * time.Hour),
			CreatedHour:       9,
			MergedHour:        intPtrRepo(10),
		},
		QualityMetrics: prDomain.PRQualityMetrics{
			ReviewCommentCount:    5,
			ReviewRoundCount:      2,
			ReviewerCount:         3,
			ReviewersInvolved:     []string{"reviewer1", "reviewer2", "reviewer3"},
			CommitCount:           8,
			FixupCommitCount:      1,
			ForceUpdateCount:      0,
			FirstReviewPassRate:   0.8,
			AverageCommentPerFile: 1.0,
			ApprovalsReceived:     2,
			ApproversInvolved:     []string{"approver1", "approver2"},
		},
		ComplexityScore: 2.5,
		SizeCategory:    prDomain.PRSizeMedium,
	}
}

func durationPtrRepo(d time.Duration) *time.Duration {
	return &d
}

func intPtrRepo(i int) *int {
	return &i
}