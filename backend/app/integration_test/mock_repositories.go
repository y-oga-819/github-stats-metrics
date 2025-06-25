package integration

import (
	"context"
	"fmt"
	"time"

	analyticsApp "github-stats-metrics/application/analytics"
	"github-stats-metrics/domain/analytics"
	prDomain "github-stats-metrics/domain/pull_request"
	"github-stats-metrics/infrastructure/repository"
)

// MockPRMetricsRepository はPRメトリクスリポジトリのモック
type MockPRMetricsRepository struct {
	prMetrics      map[string]*prDomain.PRMetrics
	dateRangeData  []*prDomain.PRMetrics
	error          error
	statistics     *repository.RepositoryStatistics
}

// NewMockPRMetricsRepository は新しいモックリポジトリを作成
func NewMockPRMetricsRepository() *MockPRMetricsRepository {
	return &MockPRMetricsRepository{
		prMetrics:     make(map[string]*prDomain.PRMetrics),
		dateRangeData: make([]*prDomain.PRMetrics, 0),
		statistics: &repository.RepositoryStatistics{
			TotalRecords:       100,
			UniqueDevelopers:   10,
			UniqueRepositories: 5,
			OldestRecord:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			NewestRecord:       time.Now(),
			AvgCycleTime:       floatPtr(86400.0), // 24 hours in seconds
			AvgComplexity:      floatPtr(2.5),
		},
	}
}

// SetPRMetrics はテスト用のPRメトリクスを設定
func (m *MockPRMetricsRepository) SetPRMetrics(prID string, metrics *prDomain.PRMetrics) {
	m.prMetrics[prID] = metrics
}

// SetDateRangeMetrics は日付範囲検索用のテストデータを設定
func (m *MockPRMetricsRepository) SetDateRangeMetrics(metrics []*prDomain.PRMetrics) {
	m.dateRangeData = metrics
}

// SetError は返すエラーを設定
func (m *MockPRMetricsRepository) SetError(err error) {
	m.error = err
}

// FindByPRID はPR IDによりPRメトリクスを取得
func (m *MockPRMetricsRepository) FindByPRID(ctx context.Context, prID string) (*prDomain.PRMetrics, error) {
	if m.error != nil {
		return nil, m.error
	}
	
	metrics, exists := m.prMetrics[prID]
	if !exists {
		return nil, nil
	}
	return metrics, nil
}

// FindByID はIDによりPRメトリクスを取得
func (m *MockPRMetricsRepository) FindByID(ctx context.Context, id string) (*prDomain.PRMetrics, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.prMetrics[id], nil
}

// FindByDateRange は日付範囲によりPRメトリクスを取得
func (m *MockPRMetricsRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time, developers []string, repositories []string) ([]*prDomain.PRMetrics, error) {
	if m.error != nil {
		return nil, m.error
	}

	var filtered []*prDomain.PRMetrics
	for _, metrics := range m.dateRangeData {
		// 日付範囲チェック
		if metrics.CreatedAt.Before(startDate) || metrics.CreatedAt.After(endDate) {
			continue
		}

		// 開発者フィルタ
		if len(developers) > 0 {
			found := false
			for _, dev := range developers {
				if metrics.Author == dev {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// リポジトリフィルタ
		if len(repositories) > 0 {
			found := false
			for _, repo := range repositories {
				if metrics.Repository == repo {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, metrics)
	}

	return filtered, nil
}

// FindByDeveloper は開発者によりPRメトリクスを取得
func (m *MockPRMetricsRepository) FindByDeveloper(ctx context.Context, developer string, startDate, endDate time.Time) ([]*prDomain.PRMetrics, error) {
	return m.FindByDateRange(ctx, startDate, endDate, []string{developer}, nil)
}

// FindByRepository はリポジトリによりPRメトリクスを取得
func (m *MockPRMetricsRepository) FindByRepository(ctx context.Context, repository string, startDate, endDate time.Time) ([]*prDomain.PRMetrics, error) {
	return m.FindByDateRange(ctx, startDate, endDate, nil, []string{repository})
}

// GetStatistics はリポジトリの統計情報を取得
func (m *MockPRMetricsRepository) GetStatistics(ctx context.Context) (*repository.RepositoryStatistics, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.statistics, nil
}

// 実装しないメソッド（テストでは使用しない）
func (m *MockPRMetricsRepository) Save(ctx context.Context, metrics *prDomain.PRMetrics) error {
	return fmt.Errorf("not implemented")
}

func (m *MockPRMetricsRepository) SaveBatch(ctx context.Context, metricsList []*prDomain.PRMetrics) error {
	return fmt.Errorf("not implemented")
}

func (m *MockPRMetricsRepository) Update(ctx context.Context, metrics *prDomain.PRMetrics) error {
	return fmt.Errorf("not implemented")
}

func (m *MockPRMetricsRepository) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockPRMetricsRepository) DeleteOldData(ctx context.Context, retentionDays int) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

// MockAggregatedMetricsRepository は集計メトリクスリポジトリのモック
type MockAggregatedMetricsRepository struct {
	teamMetrics       []*analyticsApp.TeamMetrics
	developerMetrics  map[string][]*analyticsApp.DeveloperMetrics
	repositoryMetrics map[string][]*analyticsApp.RepositoryMetrics
	allDevMetrics     map[string]*analyticsApp.DeveloperMetrics
	allRepoMetrics    map[string]*analyticsApp.RepositoryMetrics
	error             error
	statistics        *repository.AggregatedRepositoryStatistics
}

// NewMockAggregatedMetricsRepository は新しいモック集計リポジトリを作成
func NewMockAggregatedMetricsRepository() *MockAggregatedMetricsRepository {
	return &MockAggregatedMetricsRepository{
		teamMetrics:       make([]*analyticsApp.TeamMetrics, 0),
		developerMetrics:  make(map[string][]*analyticsApp.DeveloperMetrics),
		repositoryMetrics: make(map[string][]*analyticsApp.RepositoryMetrics),
		allDevMetrics:     make(map[string]*analyticsApp.DeveloperMetrics),
		allRepoMetrics:    make(map[string]*analyticsApp.RepositoryMetrics),
		statistics: &repository.AggregatedRepositoryStatistics{
			LevelStats: map[string]*repository.AggregationLevelStats{
				"team": {
					PeriodStats: map[string]*repository.PeriodStats{
						"monthly": {
							RecordCount:   12,
							UniqueTargets: 1,
							OldestPeriod:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							NewestPeriod:  time.Now(),
						},
					},
				},
			},
		},
	}
}

// SetTeamMetrics はテスト用のチームメトリクスを設定
func (m *MockAggregatedMetricsRepository) SetTeamMetrics(metrics []*analyticsApp.TeamMetrics) {
	m.teamMetrics = metrics
}

// SetDeveloperMetrics はテスト用の開発者メトリクスを設定
func (m *MockAggregatedMetricsRepository) SetDeveloperMetrics(developer string, metrics []*analyticsApp.DeveloperMetrics) {
	m.developerMetrics[developer] = metrics
	if len(metrics) > 0 {
		m.allDevMetrics[developer] = metrics[0]
	}
}

// SetRepositoryMetrics はテスト用のリポジトリメトリクスを設定
func (m *MockAggregatedMetricsRepository) SetRepositoryMetrics(repository string, metrics []*analyticsApp.RepositoryMetrics) {
	m.repositoryMetrics[repository] = metrics
	if len(metrics) > 0 {
		m.allRepoMetrics[repository] = metrics[0]
	}
}

// SetError は返すエラーを設定
func (m *MockAggregatedMetricsRepository) SetError(err error) {
	m.error = err
}

// FindTeamMetrics はチームメトリクスを取得
func (m *MockAggregatedMetricsRepository) FindTeamMetrics(ctx context.Context, period analyticsApp.AggregationPeriod, startDate, endDate time.Time) ([]*analyticsApp.TeamMetrics, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.teamMetrics, nil
}

// FindDeveloperMetrics は開発者メトリクスを取得
func (m *MockAggregatedMetricsRepository) FindDeveloperMetrics(ctx context.Context, developer string, period analyticsApp.AggregationPeriod, startDate, endDate time.Time) ([]*analyticsApp.DeveloperMetrics, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.developerMetrics[developer], nil
}

// FindRepositoryMetrics はリポジトリメトリクスを取得
func (m *MockAggregatedMetricsRepository) FindRepositoryMetrics(ctx context.Context, repository string, period analyticsApp.AggregationPeriod, startDate, endDate time.Time) ([]*analyticsApp.RepositoryMetrics, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.repositoryMetrics[repository], nil
}

// FindAllDeveloperMetrics は全開発者のメトリクスを取得
func (m *MockAggregatedMetricsRepository) FindAllDeveloperMetrics(ctx context.Context, period analyticsApp.AggregationPeriod, startDate, endDate time.Time) (map[string]*analyticsApp.DeveloperMetrics, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.allDevMetrics, nil
}

// FindAllRepositoryMetrics は全リポジトリのメトリクスを取得
func (m *MockAggregatedMetricsRepository) FindAllRepositoryMetrics(ctx context.Context, period analyticsApp.AggregationPeriod, startDate, endDate time.Time) (map[string]*analyticsApp.RepositoryMetrics, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.allRepoMetrics, nil
}

// FindLatestMetrics は最新の集計メトリクスを取得
func (m *MockAggregatedMetricsRepository) FindLatestMetrics(ctx context.Context, aggregationLevel, targetID string, period analyticsApp.AggregationPeriod) (*analytics.AggregatedMetricsStorage, error) {
	if m.error != nil {
		return nil, m.error
	}
	return nil, nil // 必要に応じて実装
}

// GetAggregatedStatistics は集計データの統計情報を取得
func (m *MockAggregatedMetricsRepository) GetAggregatedStatistics(ctx context.Context) (*repository.AggregatedRepositoryStatistics, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.statistics, nil
}

// 実装しないメソッド（テストでは使用しない）
func (m *MockAggregatedMetricsRepository) SaveTeamMetrics(ctx context.Context, metrics *analyticsApp.TeamMetrics) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAggregatedMetricsRepository) SaveDeveloperMetrics(ctx context.Context, metrics *analyticsApp.DeveloperMetrics) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAggregatedMetricsRepository) SaveRepositoryMetrics(ctx context.Context, metrics *analyticsApp.RepositoryMetrics) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAggregatedMetricsRepository) DeleteOldAggregatedData(ctx context.Context, retentionPolicy analytics.DataRetentionPolicy) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

// ヘルパー関数
func floatPtr(f float64) *float64 {
	return &f
}