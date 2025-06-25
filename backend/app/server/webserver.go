package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	pullRequestUseCase "github-stats-metrics/application/pull_request"
	pullRequestHandler "github-stats-metrics/presentation/pull_request"
	githubRepository "github-stats-metrics/infrastructure/github_api"
	todoUseCase "github-stats-metrics/application/todo"
	todoHandler "github-stats-metrics/presentation/todo"
	healthHandler "github-stats-metrics/presentation/health"
	memoryRepository "github-stats-metrics/infrastructure/memory"
	"github-stats-metrics/shared/config"
)

func StartWebServer(cfg *config.Config) error {
	fmt.Println("Start Web Server!")
	fmt.Printf("ex) http://localhost:%d/api/pull_requests\n", cfg.Server.Port)
	
	// 依存関係の注入（Clean Architecture パターン）
	// Infrastructure層 → Application層 → Presentation層の順で組み立て
	
	// Pull Request関連の依存関係
	prRepository := githubRepository.NewRepository(cfg)
	prUseCase := pullRequestUseCase.NewUseCase(prRepository)
	prHandler := pullRequestHandler.NewHandler(prUseCase)
	
	// Todo関連の依存関係
	todoRepository := memoryRepository.NewTodoRepository()
	todoUseCaseInstance := todoUseCase.NewUseCase(todoRepository)
	todoHandlerInstance := todoHandler.NewHandler(todoUseCaseInstance)
	
	// Health関連の依存関係
	healthHandlerInstance := healthHandler.NewHandler()
	
	r := mux.NewRouter().StrictSlash(true)

	// APIエンドポイントの登録
	r.HandleFunc("/health", healthHandlerInstance.GetHealth).Methods("GET")
	r.HandleFunc("/api/todos", todoHandlerInstance.GetTodos).Methods("GET")
	r.HandleFunc("/api/pull_requests", prHandler.GetPullRequests).Methods("GET")

	// CORSミドルウェアを適用
	handler := corsMiddleware(r, cfg)

	// 設定からポートを取得してサーバーを起動
	return http.ListenAndServe(cfg.GetListenAddress(), handler)
}
