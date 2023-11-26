package todo

import (
	"encoding/json"
	"net/http"

	todoDomain "github-stats-metrics/domain/todo"
)

func Success(w http.ResponseWriter, todos []todoDomain.Todo) {
	// フロントエンドとバックエンドのポートが違うので許可しておく
	// （すべてを許可する設定にしているので、本番ではより制限を厳しくしておくように）
	w.Header().Set("Access-Control-Allow-Origin", "*")

	responseBody, err := json.Marshal(todos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(responseBody)
}
