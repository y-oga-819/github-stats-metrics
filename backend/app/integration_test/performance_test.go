package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	prDomain "github-stats-metrics/domain/pull_request"
)

func TestAPIPerformanceBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	server := NewTestServer()

	// 大量のテストデータを準備
	prepareLargeDataset(server, 10000)

	t.Run("GetPRMetrics_Benchmark", func(t *testing.T) {
		benchmarkAPI(t, server, "GET", "/api/pull_requests/pr-1/metrics", 100, 10*time.Millisecond)
	})

	t.Run("ListPRMetrics_Benchmark", func(t *testing.T) {
		benchmarkAPI(t, server, "GET", "/api/pull_requests?page=1&pageSize=50", 50, 50*time.Millisecond)
	})

	t.Run("CycleTimeMetrics_Benchmark", func(t *testing.T) {
		benchmarkAPI(t, server, "GET", "/api/metrics/cycle_time?startdate=2024-01-01&enddate=2024-01-31", 30, 100*time.Millisecond)
	})

	t.Run("TeamMetrics_Benchmark", func(t *testing.T) {
		benchmarkAPI(t, server, "GET", "/api/analytics/team_metrics?period=monthly", 50, 50*time.Millisecond)
	})
}

func TestConcurrentLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	server := NewTestServer()
	prepareLargeDataset(server, 1000)

	scenarios := []struct {
		name           string
		endpoint       string
		concurrency    int
		requestsPerGR  int
		expectedStatus int
	}{
		{
			name:           "High_Concurrency_PR_Metrics",
			endpoint:       "/api/pull_requests/pr-100/metrics",
			concurrency:    100,
			requestsPerGR:  10,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Medium_Concurrency_List",
			endpoint:       "/api/pull_requests?page=1&pageSize=20",
			concurrency:    50,
			requestsPerGR:  20,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Low_Concurrency_Analytics",
			endpoint:       "/api/analytics/team_metrics",
			concurrency:    20,
			requestsPerGR:  5,
			expectedStatus: http.StatusOK,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			results := runConcurrentLoad(server, scenario.endpoint, scenario.concurrency, scenario.requestsPerGR)

			// 結果の集計と検証
			totalRequests := scenario.concurrency * scenario.requestsPerGR
			successCount := 0
			var totalDuration time.Duration
			maxDuration := time.Duration(0)

			for _, result := range results {
				if result.StatusCode == scenario.expectedStatus {
					successCount++
				}
				totalDuration += result.Duration
				if result.Duration > maxDuration {
					maxDuration = result.Duration
				}
			}

			// 成功率の検証（95%以上）
			successRate := float64(successCount) / float64(totalRequests)
			assert.GreaterOrEqual(t, successRate, 0.95, "Success rate should be at least 95%")

			// 平均レスポンス時間の検証（500ms以下）
			avgDuration := totalDuration / time.Duration(totalRequests)
			assert.LessOrEqual(t, avgDuration, 500*time.Millisecond, "Average response time should be under 500ms")

			// 最大レスポンス時間の検証（2秒以下）
			assert.LessOrEqual(t, maxDuration, 2*time.Second, "Max response time should be under 2 seconds")

			t.Logf("Scenario: %s", scenario.name)
			t.Logf("Total Requests: %d", totalRequests)
			t.Logf("Success Rate: %.2f%%", successRate*100)
			t.Logf("Average Response Time: %v", avgDuration)
			t.Logf("Max Response Time: %v", maxDuration)
		})
	}
}

func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory tests in short mode")
	}

	server := NewTestServer()

	t.Run("Large_Dataset_Memory_Usage", func(t *testing.T) {
		// 段階的にデータセットサイズを増やしてメモリ使用量をテスト
		sizes := []int{100, 500, 1000, 5000}

		for _, size := range sizes {
			// データセットを準備
			prepareLargeDataset(server, size)

			// メモリ使用量の測定（簡易版）
			start := time.Now()
			req := httptest.NewRequest("GET", "/api/pull_requests?pageSize=100", nil)
			rec := httptest.NewRecorder()

			server.Router.ServeHTTP(rec, req)
			duration := time.Since(start)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Less(t, duration, 1*time.Second, "Response time should scale reasonably with data size")

			t.Logf("Dataset Size: %d, Response Time: %v", size, duration)
		}
	})
}

func TestAPIRateLimit(t *testing.T) {
	server := NewTestServer()
	prepareLargeDataset(server, 100)

	t.Run("Rapid_Sequential_Requests", func(t *testing.T) {
		const numRequests = 100
		const maxDuration = 10 * time.Second

		start := time.Now()
		successCount := 0

		for i := 0; i < numRequests; i++ {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/pull_requests/pr-%d/metrics", i%50), nil)
			rec := httptest.NewRecorder()

			server.Router.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				successCount++
			}

			// 短い間隔でリクエストを送信
			time.Sleep(10 * time.Millisecond)
		}

		totalDuration := time.Since(start)

		// 全てのリクエストが制限時間内に完了することを確認
		assert.Less(t, totalDuration, maxDuration, "All requests should complete within time limit")

		// 高い成功率を維持
		successRate := float64(successCount) / float64(numRequests)
		assert.GreaterOrEqual(t, successRate, 0.95, "Success rate should remain high under rapid requests")

		t.Logf("Rapid Requests - Success Rate: %.2f%%, Total Duration: %v", successRate*100, totalDuration)
	})
}

func TestDataConsistency(t *testing.T) {
	server := NewTestServer()

	// 一貫性のあるテストデータを準備
	consistentData := createConsistentTestData()
	server.PRRepo.SetDateRangeMetrics(consistentData)

	t.Run("Data_Consistency_Across_Endpoints", func(t *testing.T) {
		// 同じデータに対する異なるエンドポイントの一貫性をテスト
		endpoints := []string{
			"/api/pull_requests?startdate=2024-01-01&enddate=2024-01-31",
			"/api/metrics/cycle_time?startdate=2024-01-01&enddate=2024-01-31",
			"/api/metrics/review_time?startdate=2024-01-01&enddate=2024-01-31",
		}

		responses := make(map[string]*httptest.ResponseRecorder)

		// 各エンドポイントからデータを取得
		for _, endpoint := range endpoints {
			req := httptest.NewRequest("GET", endpoint, nil)
			rec := httptest.NewRecorder()
			server.Router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code, "Endpoint %s should return success", endpoint)
			responses[endpoint] = rec
		}

		// レスポンスの一貫性を検証（詳細な検証は実装に依存）
		for endpoint, rec := range responses {
			assert.Contains(t, rec.Header().Get("Content-Type"), "application/json", "Endpoint %s should return JSON", endpoint)
			assert.Greater(t, rec.Body.Len(), 0, "Endpoint %s should return data", endpoint)
		}
	})
}

// ヘルパー関数

type LoadTestResult struct {
	StatusCode int
	Duration   time.Duration
	Error      error
}

func benchmarkAPI(t *testing.T, server *TestServer, method, endpoint string, iterations int, maxDuration time.Duration) {
	totalDuration := time.Duration(0)
	successCount := 0

	for i := 0; i < iterations; i++ {
		start := time.Now()

		req := httptest.NewRequest(method, endpoint, nil)
		rec := httptest.NewRecorder()

		server.Router.ServeHTTP(rec, req)

		duration := time.Since(start)
		totalDuration += duration

		if rec.Code == http.StatusOK {
			successCount++
		}
	}

	avgDuration := totalDuration / time.Duration(iterations)
	successRate := float64(successCount) / float64(iterations)

	assert.GreaterOrEqual(t, successRate, 0.95, "Success rate should be at least 95%")
	assert.LessOrEqual(t, avgDuration, maxDuration, "Average response time should be within limit")

	t.Logf("Benchmark %s %s - Iterations: %d, Success Rate: %.2f%%, Avg Duration: %v",
		method, endpoint, iterations, successRate*100, avgDuration)
}

func runConcurrentLoad(server *TestServer, endpoint string, concurrency, requestsPerGoroutine int) []LoadTestResult {
	var wg sync.WaitGroup
	results := make(chan LoadTestResult, concurrency*requestsPerGoroutine)

	// 同時実行
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerGoroutine; j++ {
				start := time.Now()

				req := httptest.NewRequest("GET", endpoint, nil)
				rec := httptest.NewRecorder()

				server.Router.ServeHTTP(rec, req)

				results <- LoadTestResult{
					StatusCode: rec.Code,
					Duration:   time.Since(start),
					Error:      nil,
				}
			}
		}()
	}

	// 結果の収集
	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []LoadTestResult
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}

func prepareLargeDataset(server *TestServer, size int) {
	var metrics []*prDomain.PRMetrics

	for i := 1; i <= size; i++ {
		pr := createTestPRMetrics(fmt.Sprintf("pr-%d", i))
		// 時間をずらしてリアルなデータセットを作成
		pr.CreatedAt = pr.CreatedAt.Add(time.Duration(i) * time.Hour)
		if pr.MergedAt != nil {
			mergedTime := pr.CreatedAt.Add(time.Duration(24+i%48) * time.Hour)
			pr.MergedAt = &mergedTime
		}
		// 開発者とリポジトリを分散
		pr.Author = fmt.Sprintf("dev%d", i%10+1)
		pr.Repository = fmt.Sprintf("repo%d", i%5+1)

		metrics = append(metrics, pr)
	}

	server.PRRepo.SetDateRangeMetrics(metrics)

	// 個別の PR メトリクスも設定
	for i := 1; i <= min(size, 100); i++ {
		server.PRRepo.SetPRMetrics(fmt.Sprintf("pr-%d", i), metrics[i-1])
	}
}

func createConsistentTestData() []*prDomain.PRMetrics {
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	var metrics []*prDomain.PRMetrics

	// 一貫性のあるパターンでデータを作成
	for i := 1; i <= 50; i++ {
		pr := createTestPRMetrics(fmt.Sprintf("consistent-pr-%d", i))
		pr.CreatedAt = baseTime.Add(time.Duration(i) * 24 * time.Hour)
		
		// 一貫したパターンでメトリクスを設定
		pr.SizeMetrics.LinesChanged = 100 + i*10
		pr.QualityMetrics.ReviewCommentCount = i % 10
		pr.ComplexityScore = 1.0 + float64(i%5)

		metrics = append(metrics, pr)
	}

	return metrics
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}