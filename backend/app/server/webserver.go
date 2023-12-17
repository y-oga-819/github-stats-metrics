package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	pullRequestUseCase "github-stats-metrics/application/pull_request"
	todoUseCase "github-stats-metrics/application/todo"
)

// 頭文字を大文字にするとパッケージ外部から呼び出しできる
func StartWebServer() error {
	fmt.Println("Start Web Server!")
	r := mux.NewRouter().StrictSlash(true)

	// URL に呼び出したい関数を登録する
	r.HandleFunc("/api/todos", todoUseCase.GetTodos).Methods("GET")
	r.HandleFunc("/api/pull_requests", pullRequestUseCase.GetPullRequests).Methods("GET")

	// ポートを指定してサーバーを起動する
	return http.ListenAndServe(":8080", r)
}
