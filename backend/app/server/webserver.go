package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	pullRequestUseCase "github-stats-metrics/application/pull_request"
	pullRequestHandler "github-stats-metrics/presentation/pull_request"
	githubRepository "github-stats-metrics/infrastructure/github_api"
	todoUseCase "github-stats-metrics/application/todo"
)

func StartWebServer() error {
	fmt.Println("Start Web Server!")
	fmt.Println("ex) http://localhost:8080/api/pull_requests")
	
	// 依存関係の注入（Clean Architecture パターン）
	// Infrastructure層 → Application層 → Presentation層の順で組み立て
	repository := githubRepository.NewRepository()
	useCase := pullRequestUseCase.NewUseCase(repository)
	handler := pullRequestHandler.NewHandler(useCase)
	
	r := mux.NewRouter().StrictSlash(true)

	// URL に呼び出したい関数を登録
	r.HandleFunc("/api/todos", todoUseCase.GetTodos).Methods("GET")
	r.HandleFunc("/api/pull_requests", handler.GetPullRequests).Methods("GET")

	// ポートを指定してサーバーを起動
	return http.ListenAndServe(":8080", r)
}
