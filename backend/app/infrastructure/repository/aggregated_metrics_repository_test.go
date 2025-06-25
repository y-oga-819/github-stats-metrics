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

	analyticsApp "github-stats-metrics/application/analytics"
	"github-stats-metrics/domain/analytics"
	"github-stats-metrics/shared/utils"
)

func TestAggregatedMetricsRepository_SaveTeamMetrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAggregatedMetricsRepository(db)
	
	baseTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	teamMetrics := createTestTeamMetrics(baseTime)

	mock.ExpectExec(`INSERT INTO aggregated_metrics`).
		WithArgs(
			sqlmock.AnyArg(), "team", string(teamMetrics.Period), "team", "Team",
			teamMetrics.DateRange.Start, teamMetrics.DateRange.End,
			teamMetrics.TotalPRs, teamMetrics.TotalPRs, 0,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // cycle time stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // review time stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // approval time stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // size stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // quality stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), // complexity stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // productivity stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // trend stats
			teamMetrics.GeneratedAt, sqlmock.AnyArg(), 1, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // date fields
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SaveTeamMetrics(context.Background(), teamMetrics)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAggregatedMetricsRepository_SaveDeveloperMetrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAggregatedMetricsRepository(db)
	
	baseTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	devMetrics := createTestDeveloperMetrics(baseTime)

	mock.ExpectExec(`INSERT INTO aggregated_metrics`).
		WithArgs(
			sqlmock.AnyArg(), "developer", string(devMetrics.Period), devMetrics.Developer, devMetrics.Developer,
			devMetrics.DateRange.Start, devMetrics.DateRange.End,
			devMetrics.TotalPRs, devMetrics.TotalPRs, 0,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // cycle time stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // review time stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // approval time stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // size stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // quality stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), // complexity stats
			devMetrics.Productivity.PRsPerDay, devMetrics.Productivity.LinesPerDay, devMetrics.Productivity.Throughput,
			"stable", "stable", "stable", // trend stats (固定値)
			devMetrics.GeneratedAt, sqlmock.AnyArg(), 1, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // date fields
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SaveDeveloperMetrics(context.Background(), devMetrics)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAggregatedMetricsRepository_SaveRepositoryMetrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAggregatedMetricsRepository(db)
	
	baseTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	repoMetrics := createTestRepositoryMetrics(baseTime)

	mock.ExpectExec(`INSERT INTO aggregated_metrics`).
		WithArgs(
			sqlmock.AnyArg(), "repository", string(repoMetrics.Period), repoMetrics.Repository, repoMetrics.Repository,
			repoMetrics.DateRange.Start, repoMetrics.DateRange.End,
			repoMetrics.TotalPRs, repoMetrics.TotalPRs, 0,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // cycle time stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // review time stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // approval time stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // size stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // quality stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), // complexity stats
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // productivity stats
			"stable", "stable", "stable", // trend stats (固定値)
			repoMetrics.GeneratedAt, sqlmock.AnyArg(), 1, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), // date fields
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SaveRepositoryMetrics(context.Background(), repoMetrics)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAggregatedMetricsRepository_FindTeamMetrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAggregatedMetricsRepository(db)
	
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	period := analyticsApp.AggregationPeriodDaily

	// テストデータの準備
	storage := createTestAggregatedMetricsStorage("team", string(period), "team", startDate, endDate)

	rows := createAggregatedMetricsRows().AddRow(
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

	mock.ExpectQuery(`SELECT .+ FROM aggregated_metrics WHERE aggregation_level = .+ AND aggregation_period = .+ AND period_start >= .+ AND period_end <= .+ ORDER BY period_start DESC`).
		WithArgs("team", string(period), startDate, endDate).
		WillReturnRows(rows)

	result, err := repo.FindTeamMetrics(context.Background(), period, startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, period, result[0].Period)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAggregatedMetricsRepository_FindDeveloperMetrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAggregatedMetricsRepository(db)
	
	developer := "test-dev"
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	period := analyticsApp.AggregationPeriodDaily

	storage := createTestAggregatedMetricsStorage("developer", string(period), developer, startDate, endDate)

	rows := createAggregatedMetricsRows().AddRow(
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

	mock.ExpectQuery(`SELECT .+ FROM aggregated_metrics WHERE aggregation_level = .+ AND aggregation_period = .+ AND period_start >= .+ AND period_end <= .+ AND target_id = .+ ORDER BY period_start DESC`).
		WithArgs("developer", string(period), startDate, endDate, developer).
		WillReturnRows(rows)

	result, err := repo.FindDeveloperMetrics(context.Background(), developer, period, startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, developer, result[0].Developer)
	assert.Equal(t, period, result[0].Period)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAggregatedMetricsRepository_FindRepositoryMetrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAggregatedMetricsRepository(db)
	
	repository := "test-repo"
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	period := analyticsApp.AggregationPeriodDaily

	storage := createTestAggregatedMetricsStorage("repository", string(period), repository, startDate, endDate)

	rows := createAggregatedMetricsRows().AddRow(
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

	mock.ExpectQuery(`SELECT .+ FROM aggregated_metrics WHERE aggregation_level = .+ AND aggregation_period = .+ AND period_start >= .+ AND period_end <= .+ AND target_id = .+ ORDER BY period_start DESC`).
		WithArgs("repository", string(period), startDate, endDate, repository).
		WillReturnRows(rows)

	result, err := repo.FindRepositoryMetrics(context.Background(), repository, period, startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, repository, result[0].Repository)
	assert.Equal(t, period, result[0].Period)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAggregatedMetricsRepository_FindLatestMetrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAggregatedMetricsRepository(db)
	
	aggregationLevel := "team"
	targetID := "team"
	period := analyticsApp.AggregationPeriodDaily
	baseTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	storage := createTestAggregatedMetricsStorage(aggregationLevel, string(period), targetID, baseTime, baseTime.Add(24*time.Hour))

	rows := createAggregatedMetricsRows().AddRow(
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

	mock.ExpectQuery(`SELECT .+ FROM aggregated_metrics WHERE aggregation_level = .+ AND aggregation_period = .+ AND target_id = .+ ORDER BY generated_at DESC LIMIT 1`).
		WithArgs(aggregationLevel, string(period), targetID).
		WillReturnRows(rows)

	result, err := repo.FindLatestMetrics(context.Background(), aggregationLevel, targetID, period)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, aggregationLevel, result.AggregationLevel)
	assert.Equal(t, targetID, result.TargetID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAggregatedMetricsRepository_DeleteOldAggregatedData(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAggregatedMetricsRepository(db)
	
	retentionPolicy := analytics.DataRetentionPolicy{
		DailyAggregationRetentionDays:   30,
		WeeklyAggregationRetentionDays:  90,
		MonthlyAggregationRetentionDays: 365,
	}

	// 日次データ削除
	mock.ExpectExec(`DELETE FROM aggregated_metrics WHERE aggregation_period = .+ AND generated_at <`).
		WithArgs("daily", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 10))

	// 週次データ削除
	mock.ExpectExec(`DELETE FROM aggregated_metrics WHERE aggregation_period = .+ AND generated_at <`).
		WithArgs("weekly", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 5))

	// 月次データ削除
	mock.ExpectExec(`DELETE FROM aggregated_metrics WHERE aggregation_period = .+ AND generated_at <`).
		WithArgs("monthly", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))

	deletedCount, err := repo.DeleteOldAggregatedData(context.Background(), retentionPolicy)
	assert.NoError(t, err)
	assert.Equal(t, int64(17), deletedCount) // 10 + 5 + 2

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAggregatedMetricsRepository_GetAggregatedStatistics(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAggregatedMetricsRepository(db)

	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"aggregation_level", "aggregation_period", "record_count",
		"oldest_period", "newest_period", "unique_targets",
	}).
		AddRow("team", "daily", 30, baseTime, endTime, 1).
		AddRow("developer", "daily", 150, baseTime, endTime, 5).
		AddRow("repository", "weekly", 12, baseTime, endTime, 3)

	mock.ExpectQuery(`SELECT .+ FROM aggregated_metrics GROUP BY aggregation_level, aggregation_period ORDER BY aggregation_level, aggregation_period`).
		WillReturnRows(rows)

	stats, err := repo.GetAggregatedStatistics(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Contains(t, stats.LevelStats, "team")
	assert.Contains(t, stats.LevelStats, "developer")
	assert.Contains(t, stats.LevelStats, "repository")

	// チーム統計の確認
	teamStats := stats.LevelStats["team"]
	assert.Contains(t, teamStats.PeriodStats, "daily")
	assert.Equal(t, int64(30), teamStats.PeriodStats["daily"].RecordCount)
	assert.Equal(t, int64(1), teamStats.PeriodStats["daily"].UniqueTargets)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAggregatedMetricsRepository_EdgeCases(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAggregatedMetricsRepository(db)

	t.Run("FindLatestMetrics_NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM aggregated_metrics WHERE aggregation_level = .+ AND aggregation_period = .+ ORDER BY generated_at DESC LIMIT 1`).
			WithArgs("team", "daily").
			WillReturnError(sql.ErrNoRows)

		result, err := repo.FindLatestMetrics(context.Background(), "team", "", analyticsApp.AggregationPeriodDaily)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("FindTeamMetrics_EmptyResult", func(t *testing.T) {
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		rows := createAggregatedMetricsRows()

		mock.ExpectQuery(`SELECT .+ FROM aggregated_metrics WHERE aggregation_level = .+ AND aggregation_period = .+ AND period_start >= .+ AND period_end <= .+ ORDER BY period_start DESC`).
			WithArgs("team", "daily", startDate, endDate).
			WillReturnRows(rows)

		result, err := repo.FindTeamMetrics(context.Background(), analyticsApp.AggregationPeriodDaily, startDate, endDate)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("DeleteOldData_NoRetentionPolicy", func(t *testing.T) {
		retentionPolicy := analytics.DataRetentionPolicy{
			DailyAggregationRetentionDays:   0,
			WeeklyAggregationRetentionDays:  0,
			MonthlyAggregationRetentionDays: 0,
		}

		deletedCount, err := repo.DeleteOldAggregatedData(context.Background(), retentionPolicy)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), deletedCount)
	})

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAggregatedMetricsRepository_ConversionMethods(t *testing.T) {
	repo := &AggregatedMetricsRepository{}
	
	baseTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("convertTeamMetricsToStorage", func(t *testing.T) {
		teamMetrics := createTestTeamMetrics(baseTime)
		
		storage, err := repo.convertTeamMetricsToStorage(teamMetrics)
		assert.NoError(t, err)
		assert.NotNil(t, storage)
		
		assert.Equal(t, "team", storage.AggregationLevel)
		assert.Equal(t, string(teamMetrics.Period), storage.AggregationPeriod)
		assert.Equal(t, "team", storage.TargetID)
		assert.Equal(t, teamMetrics.TotalPRs, storage.TotalPRs)
		assert.NotEmpty(t, storage.DetailedStatsJSON)
	})

	t.Run("convertDeveloperMetricsToStorage", func(t *testing.T) {
		devMetrics := createTestDeveloperMetrics(baseTime)
		
		storage, err := repo.convertDeveloperMetricsToStorage(devMetrics)
		assert.NoError(t, err)
		assert.NotNil(t, storage)
		
		assert.Equal(t, "developer", storage.AggregationLevel)
		assert.Equal(t, devMetrics.Developer, storage.TargetID)
		assert.Equal(t, devMetrics.Productivity.PRsPerDay, storage.PRsPerDay)
		assert.Equal(t, devMetrics.Productivity.LinesPerDay, storage.LinesPerDay)
		assert.Equal(t, devMetrics.Productivity.Throughput, storage.Throughput)
	})

	t.Run("convertRepositoryMetricsToStorage", func(t *testing.T) {
		repoMetrics := createTestRepositoryMetrics(baseTime)
		
		storage, err := repo.convertRepositoryMetricsToStorage(repoMetrics)
		assert.NoError(t, err)
		assert.NotNil(t, storage)
		
		assert.Equal(t, "repository", storage.AggregationLevel)
		assert.Equal(t, repoMetrics.Repository, storage.TargetID)
		assert.Equal(t, repoMetrics.TotalPRs, storage.TotalPRs)
	})

	t.Run("convertStorageToTeamMetrics", func(t *testing.T) {
		storage := createTestAggregatedMetricsStorage("team", "daily", "team", baseTime, baseTime.Add(24*time.Hour))
		
		teamMetrics, err := repo.convertStorageToTeamMetrics(storage)
		assert.NoError(t, err)
		assert.NotNil(t, teamMetrics)
		
		assert.Equal(t, analyticsApp.AggregationPeriodDaily, teamMetrics.Period)
		assert.Equal(t, storage.TotalPRs, teamMetrics.TotalPRs)
	})
}

// ヘルパー関数

func createTestTeamMetrics(baseTime time.Time) *analyticsApp.TeamMetrics {
	return &analyticsApp.TeamMetrics{
		Period:   analyticsApp.AggregationPeriodDaily,
		TotalPRs: 50,
		DateRange: analyticsApp.DateRange{
			Start: baseTime,
			End:   baseTime.Add(24 * time.Hour),
		},
		GeneratedAt: baseTime.Add(time.Hour),
		CycleTimeStats: analyticsApp.CycleTimeStatsAgg{
			TotalCycleTime: utils.DurationStatistics{
				Mean:   24 * time.Hour,
				Median: 20 * time.Hour,
				Percentiles: utils.DurationPercentiles{
					P95: 48 * time.Hour,
				},
			},
			TimeToFirstReview: utils.DurationStatistics{
				Mean:   2 * time.Hour,
				Median: 1 * time.Hour,
				Percentiles: utils.DurationPercentiles{
					P95: 8 * time.Hour,
				},
			},
			TimeToApproval: utils.DurationStatistics{
				Mean:   4 * time.Hour,
				Median: 3 * time.Hour,
				Percentiles: utils.DurationPercentiles{
					P95: 12 * time.Hour,
				},
			},
		},
		SizeStats: analyticsApp.SizeStatsAgg{
			LinesChanged: utils.IntStatistics{
				Mean:   150,
				Median: 100,
				Sum:    7500,
			},
			FilesChanged: utils.IntStatistics{
				Mean:   5,
				Median: 3,
				Sum:    250,
			},
		},
		ReviewStats: analyticsApp.ReviewStatsAgg{
			CommentCount: utils.IntStatistics{
				Mean: 5,
			},
			RoundCount: utils.IntStatistics{
				Mean: 2,
			},
			FirstReviewPassRate: utils.FloatStatistics{
				Mean: 0.8,
			},
		},
		ComplexityStats: analyticsApp.ComplexityStatsAgg{
			ComplexityScore: utils.FloatStatistics{
				Mean:   2.5,
				Median: 2.0,
			},
		},
		TrendAnalysis: analyticsApp.TrendAnalysisResult{
			CycleTimeTrend: utils.TrendAnalysis{
				Trend: "decreasing",
			},
			ReviewTimeTrend: utils.TrendAnalysis{
				Trend: "stable",
			},
			QualityTrend: utils.TrendAnalysis{
				Trend: "improving",
			},
		},
	}
}

func createTestDeveloperMetrics(baseTime time.Time) *analyticsApp.DeveloperMetrics {
	return &analyticsApp.DeveloperMetrics{
		Developer: "test-dev",
		Period:    analyticsApp.AggregationPeriodDaily,
		TotalPRs:  10,
		DateRange: analyticsApp.DateRange{
			Start: baseTime,
			End:   baseTime.Add(24 * time.Hour),
		},
		GeneratedAt: baseTime.Add(time.Hour),
		Productivity: analyticsApp.ProductivityMetrics{
			PRsPerDay:   2.0,
			LinesPerDay: 300.0,
			Throughput:  10.0,
		},
		CycleTimeStats: analyticsApp.CycleTimeStatsAgg{
			TotalCycleTime: utils.DurationStatistics{
				Mean:   18 * time.Hour,
				Median: 16 * time.Hour,
				Percentiles: utils.DurationPercentiles{
					P95: 36 * time.Hour,
				},
			},
			TimeToFirstReview: utils.DurationStatistics{
				Mean:   1 * time.Hour,
				Median: 30 * time.Minute,
				Percentiles: utils.DurationPercentiles{
					P95: 4 * time.Hour,
				},
			},
			TimeToApproval: utils.DurationStatistics{
				Mean:   3 * time.Hour,
				Median: 2 * time.Hour,
				Percentiles: utils.DurationPercentiles{
					P95: 8 * time.Hour,
				},
			},
		},
		SizeStats: analyticsApp.SizeStatsAgg{
			LinesChanged: utils.IntStatistics{
				Mean:   150,
				Median: 120,
				Sum:    1500,
			},
			FilesChanged: utils.IntStatistics{
				Mean:   4,
				Median: 3,
				Sum:    40,
			},
		},
		ReviewStats: analyticsApp.ReviewStatsAgg{
			CommentCount: utils.IntStatistics{
				Mean: 3,
			},
			RoundCount: utils.IntStatistics{
				Mean: 1,
			},
			FirstReviewPassRate: utils.FloatStatistics{
				Mean: 0.9,
			},
		},
		ComplexityStats: analyticsApp.ComplexityStatsAgg{
			ComplexityScore: utils.FloatStatistics{
				Mean:   2.0,
				Median: 1.8,
			},
		},
	}
}

func createTestRepositoryMetrics(baseTime time.Time) *analyticsApp.RepositoryMetrics {
	return &analyticsApp.RepositoryMetrics{
		Repository: "test-repo",
		Period:     analyticsApp.AggregationPeriodDaily,
		TotalPRs:   25,
		DateRange: analyticsApp.DateRange{
			Start: baseTime,
			End:   baseTime.Add(24 * time.Hour),
		},
		GeneratedAt: baseTime.Add(time.Hour),
		CycleTimeStats: analyticsApp.CycleTimeStatsAgg{
			TotalCycleTime: utils.DurationStatistics{
				Mean:   22 * time.Hour,
				Median: 18 * time.Hour,
				Percentiles: utils.DurationPercentiles{
					P95: 40 * time.Hour,
				},
			},
			TimeToFirstReview: utils.DurationStatistics{
				Mean:   3 * time.Hour,
				Median: 2 * time.Hour,
				Percentiles: utils.DurationPercentiles{
					P95: 10 * time.Hour,
				},
			},
			TimeToApproval: utils.DurationStatistics{
				Mean:   5 * time.Hour,
				Median: 4 * time.Hour,
				Percentiles: utils.DurationPercentiles{
					P95: 15 * time.Hour,
				},
			},
		},
		SizeStats: analyticsApp.SizeStatsAgg{
			LinesChanged: utils.IntStatistics{
				Mean:   200,
				Median: 150,
				Sum:    5000,
			},
			FilesChanged: utils.IntStatistics{
				Mean:   6,
				Median: 4,
				Sum:    150,
			},
		},
		ReviewStats: analyticsApp.ReviewStatsAgg{
			CommentCount: utils.IntStatistics{
				Mean: 7,
			},
			RoundCount: utils.IntStatistics{
				Mean: 2,
			},
			FirstReviewPassRate: utils.FloatStatistics{
				Mean: 0.7,
			},
		},
		ComplexityStats: analyticsApp.ComplexityStatsAgg{
			ComplexityScore: utils.FloatStatistics{
				Mean:   3.0,
				Median: 2.5,
			},
		},
	}
}

func createTestAggregatedMetricsStorage(level, period, targetID string, startDate, endDate time.Time) *analytics.AggregatedMetricsStorage {
	yearMonth := startDate.Format("2006-01")
	year, week := startDate.ISOWeek()
	weekOfYear := fmt.Sprintf("%d-W%02d", year, week)
	dayOfYear := startDate.Format("2006-002")

	avgCycleTime := int64(86400) // 24 hours in seconds
	avgReviewTime := int64(7200) // 2 hours in seconds
	avgApprovalTime := int64(14400) // 4 hours in seconds

	return &analytics.AggregatedMetricsStorage{
		ID:                fmt.Sprintf("%s_%s_%s_%d", level, targetID, period, startDate.Unix()),
		AggregationLevel:  level,
		AggregationPeriod: period,
		TargetID:          targetID,
		TargetName:        targetID,
		PeriodStart:       startDate,
		PeriodEnd:         endDate,
		TotalPRs:          50,
		MergedPRs:         50,
		ClosedPRs:         0,
		
		AvgCycleTimeSeconds:    &avgCycleTime,
		MedianCycleTimeSeconds: &avgCycleTime,
		P95CycleTimeSeconds:    &avgCycleTime,
		
		AvgReviewTimeSeconds:    &avgReviewTime,
		MedianReviewTimeSeconds: &avgReviewTime,
		P95ReviewTimeSeconds:    &avgReviewTime,
		
		AvgApprovalTimeSeconds:    &avgApprovalTime,
		MedianApprovalTimeSeconds: &avgApprovalTime,
		P95ApprovalTimeSeconds:    &avgApprovalTime,
		
		AvgLinesChanged:    150.0,
		MedianLinesChanged: 100.0,
		AvgFilesChanged:    5.0,
		MedianFilesChanged: 3.0,
		
		AvgReviewComments: 5.0,
		AvgReviewRounds:   2.0,
		FirstPassRate:     0.8,
		
		AvgComplexityScore:    2.5,
		MedianComplexityScore: 2.0,
		
		PRsPerDay:   2.0,
		LinesPerDay: 300.0,
		Throughput:  50.0,
		
		CycleTimeTrend:  "stable",
		ReviewTimeTrend: "improving",
		QualityTrend:    "stable",
		
		GeneratedAt: startDate.Add(time.Hour),
		UpdatedAt:   startDate.Add(time.Hour),
		Version:     1,
		
		DetailedStatsJSON: `{"test": "data"}`,
		YearMonth:         yearMonth,
		WeekOfYear:        weekOfYear,
		DayOfYear:         dayOfYear,
	}
}

func createAggregatedMetricsRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "aggregation_level", "aggregation_period", "target_id", "target_name",
		"period_start", "period_end", "total_prs", "merged_prs", "closed_prs",
		"avg_cycle_time_seconds", "median_cycle_time_seconds", "p95_cycle_time_seconds",
		"avg_review_time_seconds", "median_review_time_seconds", "p95_review_time_seconds",
		"avg_approval_time_seconds", "median_approval_time_seconds", "p95_approval_time_seconds",
		"avg_lines_changed", "median_lines_changed", "avg_files_changed", "median_files_changed",
		"avg_review_comments", "avg_review_rounds", "first_pass_rate",
		"avg_complexity_score", "median_complexity_score",
		"prs_per_day", "lines_per_day", "throughput",
		"cycle_time_trend", "review_time_trend", "quality_trend",
		"generated_at", "updated_at", "version", "detailed_stats_json",
		"year_month", "week_of_year", "day_of_year",
	})
}