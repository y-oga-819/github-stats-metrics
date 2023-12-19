package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	pullRequestUseCase "github-stats-metrics/application/pull_request"
	todoUseCase "github-stats-metrics/application/todo"
)

func StartWebServer() error {
	fmt.Println("Start Web Server!")
	fmt.Println("ex) http://localhost:8080/api/pull_requests")
	r := mux.NewRouter().StrictSlash(true)

	// URL に呼び出したい関数を登録
	r.HandleFunc("/api/todos", todoUseCase.GetTodos).Methods("GET")
	r.HandleFunc("/api/pull_requests", pullRequestUseCase.GetPullRequests).Methods("GET")

	// ポートを指定してサーバーを起動
	return http.ListenAndServe(":8080", r)
}
