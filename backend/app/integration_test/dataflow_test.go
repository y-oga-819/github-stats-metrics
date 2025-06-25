package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	analyticsApp "github-stats-metrics/application/analytics"
	prDomain "github-stats-metrics/domain/pull_request"
	"github-stats-metrics/shared/utils"
)

func TestCompleteDataFlow(t *testing.T) {
	server := NewTestServer()

	t.Run("PR_to_Analytics_Data_Flow", func(t *testing.T) {
		// 1. PRメトリクスデータの準備
		prMetrics := createRealisticPRDataset()
		server.PRRepo.SetDateRangeMetrics(prMetrics)

		// 2. 個別PRメトリクスの取得テスト
		req := httptest.NewRequest("GET", "/api/pull_requests/feature-pr-1/metrics", nil)
		rec := httptest.NewRecorder()
		server.PRRepo.SetPRMetrics("feature-pr-1", prMetrics[0])

		server.Router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var prResponse map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &prResponse)
		require.NoError(t, err)

		// 3. サイクルタイムメトリクスの計算テスト
		req = httptest.NewRequest("GET", "/api/metrics/cycle_time?startdate=2024-01-01&enddate=2024-02-01", nil)
		rec = httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var cycleTimeResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &cycleTimeResponse)
		require.NoError(t, err)

		// 4. レビュー時間メトリクスの計算テスト
		req = httptest.NewRequest("GET", "/api/metrics/review_time?startdate=2024-01-01&enddate=2024-02-01", nil)
		rec = httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var reviewTimeResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &reviewTimeResponse)
		require.NoError(t, err)

		// 5. 集計データからアナリティクスの計算テスト
		teamMetrics := createDerivedTeamMetrics(prMetrics)
		server.AggregatedRepo.SetTeamMetrics([]*analyticsApp.TeamMetrics{teamMetrics})

		req = httptest.NewRequest("GET", "/api/analytics/team_metrics?period=monthly", nil)
		rec = httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var teamResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &teamResponse)
		require.NoError(t, err)

		// 6. データの一貫性検証
		validateDataConsistency(t, prResponse, cycleTimeResponse, reviewTimeResponse, teamResponse)
	})
}

func TestFilterAndAggregationFlow(t *testing.T) {
	server := NewTestServer()

	// 複数の開発者とリポジトリのデータを準備
	prMetrics := createMultiDeveloperDataset()
	server.PRRepo.SetDateRangeMetrics(prMetrics)

	t.Run("Developer_Filtering_Flow", func(t *testing.T) {
		// 特定開発者のデータをフィルタ
		req := httptest.NewRequest("GET", "/api/pull_requests?developers[]=alice&startdate=2024-01-01&enddate=2024-02-01", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// Aliceのデータのみが返されることを確認
		items := response["items"].([]interface{})
		for _, item := range items {
			prItem := item.(map[string]interface{})
			assert.Equal(t, "alice", prItem["author"], "Should only return Alice's PRs")
		}

		// 開発者別メトリクスの集計
		devMetrics := createDeveloperMetricsFromPRs(filterPRsByDeveloper(prMetrics, "alice"))
		server.AggregatedRepo.SetDeveloperMetrics("alice", []*analyticsApp.DeveloperMetrics{devMetrics})

		req = httptest.NewRequest("GET", "/api/analytics/developer_metrics/alice", nil)
		rec = httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var devResponse map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &devResponse)
		require.NoError(t, err)

		assert.Equal(t, "alice", devResponse["developer"])
		assert.NotNil(t, devResponse["productivity"])
	})

	t.Run("Repository_Filtering_Flow", func(t *testing.T) {
		// 特定リポジトリのデータをフィルタ
		req := httptest.NewRequest("GET", "/api/metrics/cycle_time?repositories[]=frontend-repo&startdate=2024-01-01&enddate=2024-02-01", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// フィルタされたデータの検証
		assert.NotNil(t, response["cycleTimeStats"])
		assert.Equal(t, "frontend-repo", response["filters"].(map[string]interface{})["repositories"].([]interface{})[0])
	})

	t.Run("Combined_Filtering_Flow", func(t *testing.T) {
		// 複数フィルタの組み合わせ
		req := httptest.NewRequest("GET", "/api/metrics/review_time?developers[]=alice&developers[]=bob&repositories[]=backend-repo&startdate=2024-01-01&enddate=2024-02-01", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// 複合フィルタの結果を検証
		assert.NotNil(t, response["reviewTimeStats"])
		
		filters := response["filters"].(map[string]interface{})
		developers := filters["developers"].([]interface{})
		assert.Contains(t, developers, "alice")
		assert.Contains(t, developers, "bob")
	})
}

func TestTrendAnalysisFlow(t *testing.T) {
	server := NewTestServer()

	// 時系列データの準備
	timeSeriesData := createTimeSeriesDataset()
	server.PRRepo.SetDateRangeMetrics(timeSeriesData)

	// トレンド分析データの準備
	teamMetrics := createTrendAnalysisData()
	server.AggregatedRepo.SetTeamMetrics([]*analyticsApp.TeamMetrics{teamMetrics})

	t.Run("Trend_Analysis_Flow", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/analytics/trends?period=weekly&startdate=2024-01-01&enddate=2024-03-01", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// トレンド分析結果の検証
		assert.NotNil(t, response["cycleTimeTrend"])
		assert.NotNil(t, response["reviewTimeTrend"])
		assert.NotNil(t, response["qualityTrend"])

		// トレンド方向の検証
		cycleTimeTrend := response["cycleTimeTrend"].(map[string]interface{})
		assert.Contains(t, []string{"increasing", "decreasing", "stable"}, cycleTimeTrend["trend"])
	})
}

func TestErrorPropagationFlow(t *testing.T) {
	server := NewTestServer()

	t.Run("Upstream_Error_Propagation", func(t *testing.T) {
		// リポジトリレイヤーでエラーを発生させる
		server.PRRepo.SetError(fmt.Errorf("database connection failed"))

		// エラーがAPIレイヤーまで適切に伝播することを確認
		endpoints := []string{
			"/api/pull_requests/pr-123/metrics",
			"/api/metrics/cycle_time",
			"/api/pull_requests",
		}

		for _, endpoint := range endpoints {
			req := httptest.NewRequest("GET", endpoint, nil)
			rec := httptest.NewRecorder()

			server.Router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusInternalServerError, rec.Code, "Endpoint %s should return 500", endpoint)

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "DATABASE_ERROR", response["code"])
			assert.Contains(t, response["message"], "メトリクスの取得に失敗しました")
		}

		// エラーをクリア
		server.PRRepo.SetError(nil)
	})

	t.Run("Analytics_Error_Propagation", func(t *testing.T) {
		// 集計リポジトリでエラーを発生させる
		server.AggregatedRepo.SetError(fmt.Errorf("aggregation service unavailable"))

		analyticsEndpoints := []string{
			"/api/analytics/team_metrics",
			"/api/analytics/developer_metrics/alice",
			"/api/analytics/trends",
		}

		for _, endpoint := range analyticsEndpoints {
			req := httptest.NewRequest("GET", endpoint, nil)
			rec := httptest.NewRecorder()

			server.Router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusInternalServerError, rec.Code, "Analytics endpoint %s should return 500", endpoint)

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "DATABASE_ERROR", response["code"])
		}

		// エラーをクリア
		server.AggregatedRepo.SetError(nil)
	})
}

func TestHealthCheckFlow(t *testing.T) {
	server := NewTestServer()

	t.Run("Healthy_System_Flow", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/health", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		
		services := response["services"].(map[string]interface{})
		assert.Equal(t, "connected", services["database"])
	})

	t.Run("Unhealthy_System_Flow", func(t *testing.T) {
		// システムを不健全にする
		server.PRRepo.SetError(fmt.Errorf("database down"))

		req := httptest.NewRequest("GET", "/api/health", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "unhealthy", response["status"])

		// エラーをクリア
		server.PRRepo.SetError(nil)
	})
}

// ヘルパー関数

func createRealisticPRDataset() []*prDomain.PRMetrics {
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	var metrics []*prDomain.PRMetrics

	prConfigs := []struct {
		id          string
		author      string
		repo        string
		complexity  float64
		reviewTime  time.Duration
		cycleTime   time.Duration
		linesChanged int
	}{
		{"feature-pr-1", "alice", "frontend-repo", 2.5, 2*time.Hour, 24*time.Hour, 150},
		{"bugfix-pr-2", "bob", "backend-repo", 1.8, 1*time.Hour, 8*time.Hour, 45},
		{"refactor-pr-3", "charlie", "frontend-repo", 4.2, 6*time.Hour, 48*time.Hour, 320},
		{"feature-pr-4", "alice", "backend-repo", 3.1, 3*time.Hour, 18*time.Hour, 200},
		{"hotfix-pr-5", "dave", "backend-repo", 1.2, 30*time.Minute, 4*time.Hour, 25},
	}

	for i, config := range prConfigs {
		pr := createTestPRMetrics(config.id)
		pr.Author = config.author
		pr.Repository = config.repo
		pr.ComplexityScore = config.complexity
		pr.SizeMetrics.LinesChanged = config.linesChanged
		
		// 時間を設定
		pr.CreatedAt = baseTime.Add(time.Duration(i) * 24 * time.Hour)
		pr.TimeMetrics.TimeToFirstReview = &config.reviewTime
		pr.TimeMetrics.TotalCycleTime = &config.cycleTime
		
		mergedTime := pr.CreatedAt.Add(config.cycleTime)
		pr.MergedAt = &mergedTime

		metrics = append(metrics, pr)
	}

	return metrics
}

func createMultiDeveloperDataset() []*prDomain.PRMetrics {
	baseTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	var metrics []*prDomain.PRMetrics

	developers := []string{"alice", "bob", "charlie", "dave"}
	repositories := []string{"frontend-repo", "backend-repo", "mobile-repo"}

	for i := 0; i < 20; i++ {
		pr := createTestPRMetrics(fmt.Sprintf("multi-pr-%d", i+1))
		pr.Author = developers[i%len(developers)]
		pr.Repository = repositories[i%len(repositories)]
		pr.CreatedAt = baseTime.Add(time.Duration(i) * 12 * time.Hour)
		
		// 開発者ごとに異なる特性を設定
		switch pr.Author {
		case "alice":
			pr.ComplexityScore = 2.0 + float64(i%3)*0.5
			pr.SizeMetrics.LinesChanged = 100 + i*10
		case "bob":
			pr.ComplexityScore = 1.5 + float64(i%2)*0.3
			pr.SizeMetrics.LinesChanged = 50 + i*5
		case "charlie":
			pr.ComplexityScore = 3.0 + float64(i%4)*0.8
			pr.SizeMetrics.LinesChanged = 200 + i*15
		case "dave":
			pr.ComplexityScore = 1.0 + float64(i%2)*0.2
			pr.SizeMetrics.LinesChanged = 30 + i*3
		}

		metrics = append(metrics, pr)
	}

	return metrics
}

func createTimeSeriesDataset() []*prDomain.PRMetrics {
	baseTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	var metrics []*prDomain.PRMetrics

	// 12週間のデータを作成（週ごとにトレンドを持つ）
	for week := 0; week < 12; week++ {
		for day := 0; day < 5; day++ { // 平日のみ
			for pr := 0; pr < 3; pr++ { // 1日3PR
				prMetrics := createTestPRMetrics(fmt.Sprintf("ts-pr-%d-%d-%d", week, day, pr))
				
				// 時間的トレンドを含む
				cycleTimeBase := 12 + week*2 // 週が進むにつれて増加
				reviewTimeBase := 2 + week/2  // 緩やかに増加
				
				prMetrics.CreatedAt = baseTime.Add(time.Duration(week*7+day) * 24 * time.Hour)
				prMetrics.TimeMetrics.TotalCycleTime = durationPtr(time.Duration(cycleTimeBase) * time.Hour)
				prMetrics.TimeMetrics.TimeToFirstReview = durationPtr(time.Duration(reviewTimeBase) * time.Hour)
				
				// 品質メトリクスにもトレンドを持たせる
				prMetrics.QualityMetrics.ReviewCommentCount = 3 + week/3
				prMetrics.ComplexityScore = 2.0 + float64(week)*0.1

				metrics = append(metrics, prMetrics)
			}
		}
	}

	return metrics
}

func createDerivedTeamMetrics(prMetrics []*prDomain.PRMetrics) *analyticsApp.TeamMetrics {
	return &analyticsApp.TeamMetrics{
		Period:   analyticsApp.AggregationPeriodMonthly,
		TotalPRs: len(prMetrics),
		DateRange: analyticsApp.DateRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		GeneratedAt: time.Now(),
	}
}

func createDeveloperMetricsFromPRs(prMetrics []*prDomain.PRMetrics) *analyticsApp.DeveloperMetrics {
	if len(prMetrics) == 0 {
		return nil
	}

	return &analyticsApp.DeveloperMetrics{
		Developer: prMetrics[0].Author,
		Period:    analyticsApp.AggregationPeriodMonthly,
		TotalPRs:  len(prMetrics),
		DateRange: analyticsApp.DateRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		GeneratedAt: time.Now(),
		Productivity: analyticsApp.ProductivityMetrics{
			PRsPerDay:   float64(len(prMetrics)) / 30.0,
			LinesPerDay: 150.0, // 仮の値
			Throughput:  float64(len(prMetrics)),
		},
	}
}

func createTrendAnalysisData() *analyticsApp.TeamMetrics {
	baseTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	return &analyticsApp.TeamMetrics{
		Period:   analyticsApp.AggregationPeriodWeekly,
		TotalPRs: 35,
		DateRange: analyticsApp.DateRange{
			Start: baseTime,
			End:   baseTime.Add(7 * 24 * time.Hour),
		},
		GeneratedAt: baseTime.Add(time.Hour),
		TrendAnalysis: analyticsApp.TrendAnalysisResult{
			CycleTimeTrend: utils.TrendAnalysis{
				Trend:            "decreasing",
				Slope:            -0.5,
				CorrelationCoeff: -0.8,
				Confidence:       0.95,
			},
			ReviewTimeTrend: utils.TrendAnalysis{
				Trend:            "stable",
				Slope:            0.1,
				CorrelationCoeff: 0.2,
				Confidence:       0.65,
			},
			QualityTrend: utils.TrendAnalysis{
				Trend:            "increasing",
				Slope:            0.3,
				CorrelationCoeff: 0.7,
				Confidence:       0.85,
			},
		},
	}
}

func filterPRsByDeveloper(prMetrics []*prDomain.PRMetrics, developer string) []*prDomain.PRMetrics {
	var filtered []*prDomain.PRMetrics
	for _, pr := range prMetrics {
		if pr.Author == developer {
			filtered = append(filtered, pr)
		}
	}
	return filtered
}

func validateDataConsistency(t *testing.T, prResponse, cycleTimeResponse, reviewTimeResponse, teamResponse map[string]interface{}) {
	// 基本的な一貫性チェック
	assert.NotNil(t, prResponse["sizeMetrics"])
	assert.NotNil(t, cycleTimeResponse["cycleTimeStats"])
	assert.NotNil(t, reviewTimeResponse["reviewTimeStats"])
	assert.NotNil(t, teamResponse["cycleTimeStats"])

	// データの形式チェック
	assert.IsType(t, map[string]interface{}{}, prResponse["sizeMetrics"])
	assert.IsType(t, map[string]interface{}{}, cycleTimeResponse["cycleTimeStats"])
	
	// 時間的一貫性チェック（詳細な検証は実装に依存）
	assert.NotNil(t, teamResponse["dateRange"])
	dateRange := teamResponse["dateRange"].(map[string]interface{})
	assert.NotNil(t, dateRange["start"])
	assert.NotNil(t, dateRange["end"])
}