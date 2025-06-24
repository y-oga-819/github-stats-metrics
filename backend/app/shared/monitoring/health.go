package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github-stats-metrics/shared/config"
)

// HealthChecker はアプリケーションの健全性チェックを実行
type HealthChecker struct {
	config   *config.Config
	startTime time.Time
	checks   map[string]HealthCheck
}

// HealthCheck は個別の健全性チェック
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) HealthStatus
}

// HealthStatus は健全性チェックの結果
type HealthStatus struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// OverallHealthResponse は全体的な健全性レスポンス
type OverallHealthResponse struct {
	Status    string                    `json:"status"`
	Timestamp time.Time                 `json:"timestamp"`
	Service   string                    `json:"service"`
	Version   string                    `json:"version"`
	Uptime    string                    `json:"uptime"`
	System    SystemInfo                `json:"system"`
	Checks    map[string]HealthStatus   `json:"checks"`
}

// SystemInfo はシステム情報
type SystemInfo struct {
	GoVersion     string `json:"go_version"`
	NumGoroutines int    `json:"num_goroutines"`
	NumCPU        int    `json:"num_cpu"`
	MemoryMB      uint64 `json:"memory_mb"`
}

// NewHealthChecker は新しいヘルスチェッカーを作成
func NewHealthChecker(cfg *config.Config) *HealthChecker {
	hc := &HealthChecker{
		config:    cfg,
		startTime: time.Now(),
		checks:    make(map[string]HealthCheck),
	}

	// デフォルトのヘルスチェックを追加
	hc.AddCheck(&ConfigHealthCheck{cfg})
	hc.AddCheck(&SystemHealthCheck{})

	return hc
}

// AddCheck はヘルスチェックを追加
func (hc *HealthChecker) AddCheck(check HealthCheck) {
	hc.checks[check.Name()] = check
}

// CheckHealth は全体的な健全性をチェック
func (hc *HealthChecker) CheckHealth(ctx context.Context) OverallHealthResponse {
	checks := make(map[string]HealthStatus)
	overallHealthy := true

	// 各チェックを実行
	for name, check := range hc.checks {
		status := check.Check(ctx)
		checks[name] = status
		if status.Status != "healthy" {
			overallHealthy = false
		}
	}

	// システム情報を取得
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	systemInfo := SystemInfo{
		GoVersion:     runtime.Version(),
		NumGoroutines: runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
		MemoryMB:      memStats.Alloc / 1024 / 1024,
	}

	// 全体ステータスを決定
	status := "healthy"
	if !overallHealthy {
		status = "unhealthy"
	}

	return OverallHealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Service:   "github-stats-metrics",
		Version:   "1.0.0",
		Uptime:    time.Since(hc.startTime).String(),
		System:    systemInfo,
		Checks:    checks,
	}
}

// Handler はHTTPハンドラーを返す
func (hc *HealthChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		health := hc.CheckHealth(ctx)

		w.Header().Set("Content-Type", "application/json")
		
		// ステータスコードを設定
		statusCode := http.StatusOK
		if health.Status != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}
		w.WriteHeader(statusCode)

		if err := json.NewEncoder(w).Encode(health); err != nil {
			http.Error(w, "Failed to encode health response", http.StatusInternalServerError)
		}
	}
}

// ConfigHealthCheck は設定の健全性をチェック
type ConfigHealthCheck struct {
	config *config.Config
}

func (c *ConfigHealthCheck) Name() string {
	return "config"
}

func (c *ConfigHealthCheck) Check(ctx context.Context) HealthStatus {
	// GitHub設定をチェック
	if c.config.GitHub.Token == "" {
		return HealthStatus{
			Status:  "unhealthy",
			Message: "GitHub token not configured",
		}
	}

	if len(c.config.GitHub.Repositories) == 0 {
		return HealthStatus{
			Status:  "unhealthy",
			Message: "No target repositories configured",
		}
	}

	return HealthStatus{
		Status: "healthy",
		Details: map[string]interface{}{
			"repositories_count": len(c.config.GitHub.Repositories),
			"log_level":         c.config.Logging.Level,
		},
	}
}

// SystemHealthCheck はシステムリソースの健全性をチェック
type SystemHealthCheck struct{}

func (s *SystemHealthCheck) Name() string {
	return "system"
}

func (s *SystemHealthCheck) Check(ctx context.Context) HealthStatus {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// メモリ使用量をチェック（100MB以上で警告）
	memoryMB := memStats.Alloc / 1024 / 1024
	goroutines := runtime.NumGoroutine()

	status := "healthy"
	message := ""

	if memoryMB > 100 {
		status = "warning"
		message = fmt.Sprintf("High memory usage: %dMB", memoryMB)
	}

	if goroutines > 100 {
		if status == "healthy" {
			status = "warning"
			message = fmt.Sprintf("High goroutine count: %d", goroutines)
		} else {
			message += fmt.Sprintf(", High goroutine count: %d", goroutines)
		}
	}

	return HealthStatus{
		Status:  status,
		Message: message,
		Details: map[string]interface{}{
			"memory_mb":      memoryMB,
			"goroutines":     goroutines,
			"num_cpu":        runtime.NumCPU(),
			"go_version":     runtime.Version(),
		},
	}
}