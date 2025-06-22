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
	memoryRepository "github-stats-metrics/infrastructure/memory"
)

func StartWebServer() error {
	fmt.Println("Start Web Server!")
	fmt.Println("ex) http://localhost:8080/api/pull_requests")
	
	// 依存関係の注入（Clean Architecture パターン）
	// Infrastructure層 → Application層 → Presentation層の順で組み立て
	
	// Pull Request関連の依存関係
	prRepository := githubRepository.NewRepository()
	prUseCase := pullRequestUseCase.NewUseCase(prRepository)
	prHandler := pullRequestHandler.NewHandler(prUseCase)
	
	// Todo関連の依存関係
	todoRepository := memoryRepository.NewTodoRepository()
	todoUseCaseInstance := todoUseCase.NewUseCase(todoRepository)
	todoHandlerInstance := todoHandler.NewHandler(todoUseCaseInstance)
	
	r := mux.NewRouter().StrictSlash(true)

	// URL に呼び出したい関数を登録
	r.HandleFunc("/api/todos", todoHandlerInstance.GetTodos).Methods("GET")
	r.HandleFunc("/api/pull_requests", prHandler.GetPullRequests).Methods("GET")

	// CORSミドルウェアを適用
	handler := corsMiddleware(r)

	// ポートを指定してサーバーを起動
	return http.ListenAndServe(":8080", handler)
}
