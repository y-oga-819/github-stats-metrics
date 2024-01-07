package pull_request

import (
	"fmt"
	prDomain "github-stats-metrics/domain/pull_request"
	githubApiClient "github-stats-metrics/infrastructure/github_api"
	presenter "github-stats-metrics/presentation/pull_request"
	"net/http"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

func GetPullRequests(w http.ResponseWriter, r *http.Request) {
	req := &prDomain.GetPullRequestsRequest{}

	if err := decoder.Decode(req, r.URL.Query()); err != nil {
		http.Error(w, fmt.Sprintf("bad request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	pullRequests := githubApiClient.Fetch(*req)

	presenter.Success(w, pullRequests)
}
