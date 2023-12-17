package pull_request

import (
	"encoding/json"
	"net/http"

	domain "github-stats-metrics/domain/pull_request"
)

func Success(w http.ResponseWriter, pullRequests []domain.PullRequest) {
	// フロントエンドとバックエンドのポートが違うので許可しておく
	// （すべてを許可する設定にしているので、本番ではより制限を厳しくしておくように）
	w.Header().Set("Access-Control-Allow-Origin", "*")

	responseBody, err := json.Marshal(pullRequests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(responseBody)
}
