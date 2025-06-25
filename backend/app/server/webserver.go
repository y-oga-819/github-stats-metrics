package server

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"

	analyticsApp "github-stats-metrics/application/analytics"
	pullRequestUseCase "github-stats-metrics/application/pull_request"
	pullRequestHandler "github-stats-metrics/presentation/pull_request"
	githubRepository "github-stats-metrics/infrastructure/github_api"
	"github-stats-metrics/infrastructure/repository"
	todoUseCase "github-stats-metrics/application/todo"
	todoHandler "github-stats-metrics/presentation/todo"
	memoryRepository "github-stats-metrics/infrastructure/memory"
	"github-stats-metrics/shared/config"
	"github-stats-metrics/shared/logging"
	"github-stats-metrics/shared/metrics"
	"github-stats-metrics/shared/middleware"
	"github-stats-metrics/shared/monitoring"
)

func StartWebServer(cfg *config.Config, logger *logging.StructuredLogger, metricsCollector *metrics.MetricsCollector, healthChecker *monitoring.HealthChecker, db *sql.DB) error {
	ctx := context.Background()
	
	logger.Info(ctx, "Initializing web server", map[string]interface{}{
		"listen_address": cfg.GetListenAddress(),
	})
	
	// 依存関係の注入（Clean Architecture パターン）
	// Infrastructure層 → Application層 → Presentation層の順で組み立て
	
	// Pull Request関連の依存関係
	prRepository := githubRepository.NewRepository(cfg)
	prUseCase := pullRequestUseCase.NewUseCase(prRepository)
	prHandler := pullRequestHandler.NewHandler(prUseCase)
	
	// PRメトリクス関連の依存関係
	prMetricsRepo := repository.NewPRMetricsRepository(db)
	metricsAggregator := analyticsApp.NewMetricsAggregator()
	prMetricsHandler := pullRequestHandler.NewPRMetricsHandler(prMetricsRepo, metricsAggregator)
	
	// Todo関連の依存関係
	todoRepository := memoryRepository.NewTodoRepository()
	todoUseCaseInstance := todoUseCase.NewUseCase(todoRepository)
	todoHandlerInstance := todoHandler.NewHandler(todoUseCaseInstance)
	
	r := mux.NewRouter().StrictSlash(true)

	// 監視エンドポイント
	r.Handle("/metrics", metricsCollector.GetHandler()).Methods("GET")
	r.HandleFunc("/health", healthChecker.Handler()).Methods("GET")

	// API エンドポイント
	r.HandleFunc("/api/todos", todoHandlerInstance.GetTodos).Methods("GET")
	r.HandleFunc("/api/pull_requests", prHandler.GetPullRequests).Methods("GET")
	
	// PRメトリクス API ルートの登録
	prMetricsHandler.RegisterRoutes(r)

	// ミドルウェアの適用
	handler := corsMiddleware(r, cfg)
	handler = middleware.MetricsMiddleware(metricsCollector)(handler)

	logger.Info(ctx, "Web server starting", map[string]interface{}{
		"endpoints": []string{
			"/api/todos",
			"/api/pull_requests", 
			"/api/pull_requests/{id}/metrics",
			"/api/metrics/cycle_time",
			"/api/metrics/review_time",
			"/api/developers/{developer}/metrics",
			"/api/repositories/{repository}/metrics",
			"/health",
			"/metrics",
		},
	})

	// 設定からポートを取得してサーバーを起動
	return http.ListenAndServe(cfg.GetListenAddress(), handler)
}
