package middleware

import (
	"net/http"
	"strings"
	"time"

	"github-stats-metrics/shared/metrics"
)

// MetricsMiddleware はHTTPリクエストのメトリクスを収集するミドルウェア
func MetricsMiddleware(collector *metrics.MetricsCollector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// インフライトリクエストを増加
			collector.IncrementInFlightRequests()
			defer collector.DecrementInFlightRequests()

			// レスポンスライターをラップしてステータスコードを取得
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// リクエスト処理時間を測定
			start := time.Now()
			next.ServeHTTP(wrapped, r)
			duration := time.Since(start)

			// エンドポイントを正規化（パスパラメータを除去）
			endpoint := normalizeEndpoint(r.URL.Path)

			// メトリクスを記録
			collector.RecordHTTPRequest(r.Method, endpoint, wrapped.statusCode, duration)
		})
	}
}

// responseWriter はhttp.ResponseWriterをラップしてステータスコードを取得
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader はステータスコードを記録
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// normalizeEndpoint はエンドポイントパスを正規化
func normalizeEndpoint(path string) string {
	// メトリクスエンドポイントはそのまま
	if path == "/metrics" || path == "/health" {
		return path
	}

	// APIエンドポイントの正規化
	if strings.HasPrefix(path, "/api/") {
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			// /api/pull_requests, /api/todos などを正規化
			return "/api/" + parts[2]
		}
		return "/api/*"
	}

	// その他のパスは根本パスとして扱う
	if path == "/" || path == "" {
		return "/"
	}

	return "/other"
}