package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector はアプリケーションメトリクスを収集
type MetricsCollector struct {
	// HTTP メトリクス
	httpRequestsTotal      *prometheus.CounterVec
	httpRequestDuration    *prometheus.HistogramVec
	httpRequestsInFlight   prometheus.Gauge

	// API メトリクス
	githubAPICallsTotal    *prometheus.CounterVec
	githubAPIRateLimit     prometheus.Gauge
	githubAPIRemaining     prometheus.Gauge

	// Business メトリクス
	pullRequestsProcessed  *prometheus.CounterVec
	pullRequestsRetrieved  prometheus.Histogram
	businessRuleFiltered   *prometheus.CounterVec

	// System メトリクス
	applicationStartTime   prometheus.Gauge
	applicationInfo        *prometheus.GaugeVec
}

// NewMetricsCollector は新しいメトリクスコレクターを作成
func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		httpRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),
		githubAPICallsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "github_api_calls_total",
				Help: "Total number of GitHub API calls",
			},
			[]string{"operation", "status"},
		),
		githubAPIRateLimit: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "github_api_rate_limit",
				Help: "GitHub API rate limit",
			},
		),
		githubAPIRemaining: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "github_api_remaining",
				Help: "GitHub API remaining rate limit",
			},
		),
		pullRequestsProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pull_requests_processed_total",
				Help: "Total number of pull requests processed",
			},
			[]string{"status"},
		),
		pullRequestsRetrieved: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "pull_requests_retrieved_count",
				Help:    "Number of pull requests retrieved per request",
				Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
			},
		),
		businessRuleFiltered: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "business_rule_filtered_total",
				Help: "Total number of items filtered by business rules",
			},
			[]string{"rule_type"},
		),
		applicationStartTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "application_start_time_seconds",
				Help: "Application start time in Unix timestamp",
			},
		),
		applicationInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "application_info",
				Help: "Application information",
			},
			[]string{"version", "build_date", "git_commit"},
		),
	}

	// アプリケーション開始時間を記録
	mc.applicationStartTime.SetToCurrentTime()

	return mc
}

// Register はPrometheusレジストリにメトリクスを登録
func (mc *MetricsCollector) Register() error {
	metrics := []prometheus.Collector{
		mc.httpRequestsTotal,
		mc.httpRequestDuration,
		mc.httpRequestsInFlight,
		mc.githubAPICallsTotal,
		mc.githubAPIRateLimit,
		mc.githubAPIRemaining,
		mc.pullRequestsProcessed,
		mc.pullRequestsRetrieved,
		mc.businessRuleFiltered,
		mc.applicationStartTime,
		mc.applicationInfo,
	}

	for _, metric := range metrics {
		if err := prometheus.Register(metric); err != nil {
			return fmt.Errorf("failed to register metric: %w", err)
		}
	}

	return nil
}

// HTTPミドルウェア用のメソッド

// RecordHTTPRequest はHTTPリクエストメトリクスを記録
func (mc *MetricsCollector) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	mc.httpRequestsTotal.WithLabelValues(method, endpoint, strconv.Itoa(statusCode)).Inc()
	mc.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// IncrementInFlightRequests はインフライトリクエスト数を増加
func (mc *MetricsCollector) IncrementInFlightRequests() {
	mc.httpRequestsInFlight.Inc()
}

// DecrementInFlightRequests はインフライトリクエスト数を減少
func (mc *MetricsCollector) DecrementInFlightRequests() {
	mc.httpRequestsInFlight.Dec()
}

// GitHub API用のメソッド

// RecordGitHubAPICall はGitHub APIコールを記録
func (mc *MetricsCollector) RecordGitHubAPICall(operation, status string) {
	mc.githubAPICallsTotal.WithLabelValues(operation, status).Inc()
}

// UpdateGitHubAPIRateLimit はGitHub APIレート制限を更新
func (mc *MetricsCollector) UpdateGitHubAPIRateLimit(limit, remaining int) {
	mc.githubAPIRateLimit.Set(float64(limit))
	mc.githubAPIRemaining.Set(float64(remaining))
}

// ビジネスメトリクス用のメソッド

// RecordPullRequestProcessed はプルリクエスト処理数を記録
func (mc *MetricsCollector) RecordPullRequestProcessed(status string) {
	mc.pullRequestsProcessed.WithLabelValues(status).Inc()
}

// RecordPullRequestsRetrieved は取得したプルリクエスト数を記録
func (mc *MetricsCollector) RecordPullRequestsRetrieved(count int) {
	mc.pullRequestsRetrieved.Observe(float64(count))
}

// RecordBusinessRuleFiltered はビジネスルールによるフィルタリングを記録
func (mc *MetricsCollector) RecordBusinessRuleFiltered(ruleType string) {
	mc.businessRuleFiltered.WithLabelValues(ruleType).Inc()
}

// SetApplicationInfo はアプリケーション情報を設定
func (mc *MetricsCollector) SetApplicationInfo(version, buildDate, gitCommit string) {
	mc.applicationInfo.WithLabelValues(version, buildDate, gitCommit).Set(1)
}

// GetHandler はPrometheusメトリクスハンドラーを返す
func (mc *MetricsCollector) GetHandler() http.Handler {
	return promhttp.Handler()
}