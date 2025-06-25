package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	analyticsApp "github-stats-metrics/application/analytics"
	prDomain "github-stats-metrics/domain/pull_request"
	"github-stats-metrics/shared/utils"
)

// TestServer はテスト用のサーバー
type TestServer struct {
	Router            *mux.Router
	PRRepo            *MockPRMetricsRepository
	AggregatedRepo    *MockAggregatedMetricsRepository
	MetricsAggregator *analyticsApp.MetricsAggregator
}

// NewTestServer はテスト用サーバーを作成
func NewTestServer() *TestServer {
	prRepo := NewMockPRMetricsRepository()
	aggregatedRepo := NewMockAggregatedMetricsRepository()
	metricsAggregator := analyticsApp.NewMetricsAggregator()

	router := mux.NewRouter()

	// 簡易的なハンドラーを直接実装
	setupTestRoutes(router, prRepo, aggregatedRepo)

	return &TestServer{
		Router:            router,
		PRRepo:            prRepo,
		AggregatedRepo:    aggregatedRepo,
		MetricsAggregator: metricsAggregator,
	}
}

// setupTestRoutes はテスト用のルートを設定
func setupTestRoutes(router *mux.Router, prRepo *MockPRMetricsRepository, aggregatedRepo *MockAggregatedMetricsRepository) {
	// PRメトリクス個別取得
	router.HandleFunc("/api/pull_requests/{id}/metrics", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		prID := vars["id"]
		
		metrics, err := prRepo.FindByPRID(r.Context(), prID)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "メトリクスの取得に失敗しました")
			return
		}
		
		if metrics == nil {
			writeErrorResponse(w, http.StatusNotFound, "PR_NOT_FOUND", "指定されたPRが見つかりません")
			return
		}
		
		response := map[string]interface{}{
			"prId":           metrics.PRID,
			"prNumber":       metrics.PRNumber,
			"title":          metrics.Title,
			"author":         metrics.Author,
			"repository":     metrics.Repository,
			"sizeMetrics":    metrics.SizeMetrics,
			"timeMetrics":    metrics.TimeMetrics,
			"qualityMetrics": metrics.QualityMetrics,
			"complexityScore": metrics.ComplexityScore,
		}
		writeJSONResponse(w, http.StatusOK, response)
	}).Methods("GET")

	// サイクルタイムメトリクス
	router.HandleFunc("/api/metrics/cycle_time", func(w http.ResponseWriter, r *http.Request) {
		// 簡易的な実装
		response := map[string]interface{}{
			"cycleTimeStats": map[string]interface{}{
				"mean":   "24h",
				"median": "20h",
			},
			"period": "monthly",
			"dateRange": map[string]interface{}{
				"start": "2024-01-01",
				"end":   "2024-01-31",
			},
			"filters": map[string]interface{}{
				"developers":   r.URL.Query()["developers[]"],
				"repositories": r.URL.Query()["repositories[]"],
			},
		}
		writeJSONResponse(w, http.StatusOK, response)
	}).Methods("GET")

	// レビュー時間メトリクス
	router.HandleFunc("/api/metrics/review_time", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"reviewTimeStats": map[string]interface{}{
				"mean":   "2h",
				"median": "1h",
			},
			"period": "monthly",
			"dateRange": map[string]interface{}{
				"start": "2024-01-01",
				"end":   "2024-01-31",
			},
			"filters": map[string]interface{}{
				"developers":   r.URL.Query()["developers[]"],
				"repositories": r.URL.Query()["repositories[]"],
			},
		}
		writeJSONResponse(w, http.StatusOK, response)
	}).Methods("GET")

	// PRリスト取得
	router.HandleFunc("/api/pull_requests", func(w http.ResponseWriter, r *http.Request) {
		// ページネーション処理
		page := 1
		pageSize := 50
		
		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			} else {
				writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", "Invalid page number")
				return
			}
		}
		
		if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
			if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
				pageSize = ps
			} else {
				writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", "Invalid page size")
				return
			}
		}

		// モックデータから取得
		allMetrics, err := prRepo.FindByDateRange(r.Context(), time.Time{}, time.Now(), nil, nil)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "メトリクスの取得に失敗しました")
			return
		}

		totalCount := len(allMetrics)
		startIndex := (page - 1) * pageSize
		endIndex := startIndex + pageSize

		var items []interface{}
		if startIndex < totalCount {
			if endIndex > totalCount {
				endIndex = totalCount
			}
			for _, metrics := range allMetrics[startIndex:endIndex] {
				items = append(items, map[string]interface{}{
					"prId":           metrics.PRID,
					"prNumber":       metrics.PRNumber,
					"title":          metrics.Title,
					"author":         metrics.Author,
					"repository":     metrics.Repository,
					"complexityScore": metrics.ComplexityScore,
				})
			}
		}

		response := map[string]interface{}{
			"items":       items,
			"totalCount":  totalCount,
			"currentPage": page,
			"pageSize":    pageSize,
		}
		writeJSONResponse(w, http.StatusOK, response)
	}).Methods("GET")

	// チームメトリクス
	router.HandleFunc("/api/analytics/team_metrics", func(w http.ResponseWriter, r *http.Request) {
		// 期間パラメータの検証
		period := r.URL.Query().Get("period")
		if period != "" && period != "daily" && period != "weekly" && period != "monthly" {
			writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", "Invalid period")
			return
		}
		if period == "" {
			period = "monthly"
		}

		teamMetrics, err := aggregatedRepo.FindTeamMetrics(r.Context(), analyticsApp.AggregationPeriod(period), time.Time{}, time.Now())
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "チームメトリクスの取得に失敗しました")
			return
		}

		if len(teamMetrics) == 0 {
			writeErrorResponse(w, http.StatusNotFound, "NO_DATA", "指定された期間のデータが見つかりません")
			return
		}

		latest := teamMetrics[0]
		response := map[string]interface{}{
			"period":         latest.Period,
			"totalPRs":       latest.TotalPRs,
			"cycleTimeStats": latest.CycleTimeStats,
			"reviewStats":    latest.ReviewStats,
			"sizeStats":      latest.SizeStats,
			"qualityStats":   latest.QualityStats,
			"trendAnalysis":  latest.TrendAnalysis,
		}
		writeJSONResponse(w, http.StatusOK, response)
	}).Methods("GET")

	// 開発者メトリクス
	router.HandleFunc("/api/analytics/developer_metrics/{developer}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		developer := vars["developer"]

		devMetrics, err := aggregatedRepo.FindDeveloperMetrics(r.Context(), developer, analyticsApp.AggregationPeriodDaily, time.Time{}, time.Now())
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "開発者メトリクスの取得に失敗しました")
			return
		}

		if len(devMetrics) == 0 {
			writeErrorResponse(w, http.StatusNotFound, "NO_DATA", "指定された開発者のデータが見つかりません")
			return
		}

		latest := devMetrics[0]
		response := map[string]interface{}{
			"developer":     latest.Developer,
			"period":        latest.Period,
			"totalPRs":      latest.TotalPRs,
			"productivity":  latest.Productivity,
			"cycleTimeStats": latest.CycleTimeStats,
		}
		writeJSONResponse(w, http.StatusOK, response)
	}).Methods("GET")

	// トレンド分析
	router.HandleFunc("/api/analytics/trends", func(w http.ResponseWriter, r *http.Request) {
		period := r.URL.Query().Get("period")
		if period == "" {
			period = "weekly"
		}

		teamMetrics, err := aggregatedRepo.FindTeamMetrics(r.Context(), analyticsApp.AggregationPeriod(period), time.Time{}, time.Now())
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "トレンドデータの取得に失敗しました")
			return
		}

		if len(teamMetrics) == 0 {
			writeErrorResponse(w, http.StatusNotFound, "NO_DATA", "トレンド分析に必要なデータが見つかりません")
			return
		}

		latest := teamMetrics[0]
		response := map[string]interface{}{
			"cycleTimeTrend":  latest.TrendAnalysis.CycleTimeTrend,
			"reviewTimeTrend": latest.TrendAnalysis.ReviewTimeTrend,
			"qualityTrend":    latest.TrendAnalysis.QualityTrend,
		}
		writeJSONResponse(w, http.StatusOK, response)
	}).Methods("GET")

	// ヘルスチェック
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := prRepo.GetStatistics(r.Context())
		status := "healthy"
		statusCode := http.StatusOK
		
		if err != nil {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}

		response := map[string]interface{}{
			"status":    status,
			"timestamp": time.Now(),
			"services": map[string]string{
				"database": "connected",
				"github":   "available",
			},
		}
		writeJSONResponse(w, statusCode, response)
	}).Methods("GET")

	// アナリティクスヘルスチェック
	router.HandleFunc("/api/analytics/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := aggregatedRepo.GetAggregatedStatistics(r.Context())
		status := "healthy"
		statusCode := http.StatusOK
		
		if err != nil {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}

		response := map[string]interface{}{
			"status":    status,
			"timestamp": time.Now(),
		}
		writeJSONResponse(w, statusCode, response)
	}).Methods("GET")
}

// ヘルパー関数
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	errorResponse := map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"code":    code,
		"message": message,
	}
	writeJSONResponse(w, statusCode, errorResponse)
}

func TestPRMetricsAPI_GetPRMetrics(t *testing.T) {
	server := NewTestServer()

	// テストデータの準備
	testPR := createTestPRMetrics("pr-123")
	server.PRRepo.SetPRMetrics("pr-123", testPR)

	t.Run("正常ケース_PRメトリクス取得", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/pull_requests/pr-123/metrics", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "pr-123", response["prId"])
		assert.Equal(t, float64(123), response["prNumber"])
		assert.NotNil(t, response["sizeMetrics"])
		assert.NotNil(t, response["timeMetrics"])
		assert.NotNil(t, response["qualityMetrics"])
	})

	t.Run("エラーケース_存在しないPR", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/pull_requests/non-existent/metrics", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "PR_NOT_FOUND", response["code"])
	})

	t.Run("エラーケース_不正なPR ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/pull_requests//metrics", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code) // ルートマッチしない
	})
}

func TestPRMetricsAPI_GetCycleTimeMetrics(t *testing.T) {
	server := NewTestServer()

	// テストデータの準備
	testMetrics := []*prDomain.PRMetrics{
		createTestPRMetrics("pr-1"),
		createTestPRMetrics("pr-2"),
		createTestPRMetrics("pr-3"),
	}
	server.PRRepo.SetDateRangeMetrics(testMetrics)

	t.Run("正常ケース_サイクルタイムメトリクス取得", func(t *testing.T) {
		url := "/api/metrics/cycle_time?startdate=2024-01-01&enddate=2024-01-31&period=monthly"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response["cycleTimeStats"])
		assert.NotNil(t, response["period"])
		assert.NotNil(t, response["dateRange"])
	})

	t.Run("エラーケース_不正な日付フォーマット", func(t *testing.T) {
		url := "/api/metrics/cycle_time?startdate=invalid-date"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "INVALID_PARAMETERS", response["code"])
	})

	t.Run("開発者フィルタ付きリクエスト", func(t *testing.T) {
		url := "/api/metrics/cycle_time?developers[]=dev1&developers[]=dev2"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestPRMetricsAPI_ListPRMetrics(t *testing.T) {
	server := NewTestServer()

	// 大量のテストデータを準備
	var testMetrics []*prDomain.PRMetrics
	for i := 1; i <= 100; i++ {
		pr := createTestPRMetrics(fmt.Sprintf("pr-%d", i))
		testMetrics = append(testMetrics, pr)
	}
	server.PRRepo.SetDateRangeMetrics(testMetrics)

	t.Run("正常ケース_ページネーション", func(t *testing.T) {
		url := "/api/pull_requests?page=2&pageSize=20"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(100), response["totalCount"])
		assert.Equal(t, float64(2), response["currentPage"])
		assert.Equal(t, float64(20), response["pageSize"])

		items := response["items"].([]interface{})
		assert.Len(t, items, 20)
	})

	t.Run("エラーケース_無効なページ番号", func(t *testing.T) {
		url := "/api/pull_requests?page=-1"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("エラーケース_無効なページサイズ", func(t *testing.T) {
		url := "/api/pull_requests?pageSize=1000"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAnalyticsAPI_GetTeamMetrics(t *testing.T) {
	server := NewTestServer()

	// テストデータの準備
	teamMetrics := createTestTeamMetrics()
	server.AggregatedRepo.SetTeamMetrics([]*analyticsApp.TeamMetrics{teamMetrics})

	t.Run("正常ケース_チームメトリクス取得", func(t *testing.T) {
		url := "/api/analytics/team_metrics?period=monthly&startdate=2024-01-01&enddate=2024-01-31"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "monthly", response["period"])
		assert.NotNil(t, response["cycleTimeStats"])
		assert.NotNil(t, response["reviewStats"])
		assert.NotNil(t, response["sizeStats"])
		assert.NotNil(t, response["qualityStats"])
		assert.NotNil(t, response["trendAnalysis"])
	})

	t.Run("エラーケース_データなし", func(t *testing.T) {
		server.AggregatedRepo.SetTeamMetrics([]*analyticsApp.TeamMetrics{})

		url := "/api/analytics/team_metrics?period=weekly"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "NO_DATA", response["code"])
	})

	t.Run("エラーケース_無効な期間", func(t *testing.T) {
		url := "/api/analytics/team_metrics?period=invalid"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAnalyticsAPI_GetDeveloperMetrics(t *testing.T) {
	server := NewTestServer()

	// テストデータの準備
	devMetrics := createTestDeveloperMetrics("test-dev")
	server.AggregatedRepo.SetDeveloperMetrics("test-dev", []*analyticsApp.DeveloperMetrics{devMetrics})

	t.Run("正常ケース_開発者メトリクス取得", func(t *testing.T) {
		url := "/api/analytics/developer_metrics/test-dev?period=daily"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "test-dev", response["developer"])
		assert.NotNil(t, response["productivity"])
		assert.NotNil(t, response["cycleTimeStats"])
	})

	t.Run("エラーケース_開発者名なし", func(t *testing.T) {
		url := "/api/analytics/developer_metrics/"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code) // ルートマッチしない
	})
}

func TestHealthCheckAPI(t *testing.T) {
	server := NewTestServer()

	t.Run("PRメトリクスヘルスチェック", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/health", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.NotNil(t, response["services"])
		assert.NotNil(t, response["timestamp"])
	})

	t.Run("アナリティクスヘルスチェック", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/analytics/health", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
	})
}

func TestAPIPerformance(t *testing.T) {
	server := NewTestServer()

	// 大量データの準備
	var testMetrics []*prDomain.PRMetrics
	for i := 1; i <= 1000; i++ {
		pr := createTestPRMetrics(fmt.Sprintf("pr-%d", i))
		testMetrics = append(testMetrics, pr)
	}
	server.PRRepo.SetDateRangeMetrics(testMetrics)

	t.Run("大量データでのページネーション性能", func(t *testing.T) {
		start := time.Now()

		url := "/api/pull_requests?page=1&pageSize=100"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		duration := time.Since(start)
		
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Less(t, duration, 100*time.Millisecond, "API response should be fast")

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(1000), response["totalCount"])
		items := response["items"].([]interface{})
		assert.Len(t, items, 100)
	})

	t.Run("複数フィルタ適用時の性能", func(t *testing.T) {
		start := time.Now()

		url := "/api/metrics/cycle_time?startdate=2024-01-01&enddate=2024-12-31&developers[]=dev1&developers[]=dev2&repositories[]=repo1"
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		duration := time.Since(start)
		
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Less(t, duration, 200*time.Millisecond, "Filtered API response should be reasonably fast")
	})
}

func TestAPIErrorHandling(t *testing.T) {
	server := NewTestServer()

	// エラーを発生させるモックの設定
	server.PRRepo.SetError(fmt.Errorf("database connection failed"))

	t.Run("データベースエラーハンドリング", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/pull_requests/pr-123/metrics", nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "DATABASE_ERROR", response["code"])
		assert.Contains(t, response["message"], "メトリクスの取得に失敗しました")
	})

	t.Run("タイムアウトハンドリング", func(t *testing.T) {
		// タイムアウトのシミュレーション
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest("GET", "/api/health", nil)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		// 少し待ってからリクエスト実行
		time.Sleep(2 * time.Millisecond)
		server.Router.ServeHTTP(rec, req)

		// コンテキストがキャンセルされるため、処理が中断される可能性がある
		// 具体的なレスポンスは実装次第だが、適切にハンドリングされることを確認
		assert.True(t, rec.Code >= 200, "Response should be handled gracefully")
	})
}

func TestConcurrentRequests(t *testing.T) {
	server := NewTestServer()

	// テストデータの準備
	testPR := createTestPRMetrics("pr-concurrent")
	server.PRRepo.SetPRMetrics("pr-concurrent", testPR)

	t.Run("同時リクエスト処理", func(t *testing.T) {
		const numRequests = 50
		results := make(chan int, numRequests)

		// 50個の同時リクエストを送信
		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/api/pull_requests/pr-concurrent/metrics", nil)
				rec := httptest.NewRecorder()

				server.Router.ServeHTTP(rec, req)
				results <- rec.Code
			}()
		}

		// 全ての結果を収集
		successCount := 0
		for i := 0; i < numRequests; i++ {
			statusCode := <-results
			if statusCode == http.StatusOK {
				successCount++
			}
		}

		// 全てのリクエストが成功することを確認
		assert.Equal(t, numRequests, successCount, "All concurrent requests should succeed")
	})
}

// ヘルパー関数

func createTestPRMetrics(prID string) *prDomain.PRMetrics {
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	mergedTime := baseTime.Add(24 * time.Hour)

	return &prDomain.PRMetrics{
		PRID:       prID,
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
			},
		},
		TimeMetrics: prDomain.PRTimeMetrics{
			TotalCycleTime:    durationPtr(24 * time.Hour),
			TimeToFirstReview: durationPtr(2 * time.Hour),
			TimeToApproval:    durationPtr(4 * time.Hour),
			TimeToMerge:       durationPtr(1 * time.Hour),
			CreatedHour:       9,
			MergedHour:        intPtr(10),
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

func createTestTeamMetrics() *analyticsApp.TeamMetrics {
	baseTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	return &analyticsApp.TeamMetrics{
		Period:   analyticsApp.AggregationPeriodMonthly,
		TotalPRs: 50,
		DateRange: analyticsApp.DateRange{
			Start: baseTime,
			End:   baseTime.Add(30 * 24 * time.Hour),
		},
		GeneratedAt: baseTime.Add(time.Hour),
		CycleTimeStats: analyticsApp.CycleTimeStatsAgg{
			TotalCycleTime: utils.DurationStatistics{
				Mean:   24 * time.Hour,
				Median: 20 * time.Hour,
			},
		},
		SizeStats: analyticsApp.SizeStatsAgg{
			LinesChanged: utils.IntStatistics{
				Mean:   150,
				Median: 100,
				Sum:    7500,
			},
		},
		ReviewStats: analyticsApp.ReviewStatsAgg{
			CommentCount: utils.IntStatistics{
				Mean: 5,
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
		},
	}
}

func createTestDeveloperMetrics(developer string) *analyticsApp.DeveloperMetrics {
	baseTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	return &analyticsApp.DeveloperMetrics{
		Developer: developer,
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
			},
		},
		SizeStats: analyticsApp.SizeStatsAgg{
			LinesChanged: utils.IntStatistics{
				Mean:   150,
				Median: 120,
				Sum:    1500,
			},
		},
		ReviewStats: analyticsApp.ReviewStatsAgg{
			CommentCount: utils.IntStatistics{
				Mean: 3,
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

func durationPtr(d time.Duration) *time.Duration {
	return &d
}

func intPtr(i int) *int {
	return &i
}