package pull_request

import (
	githubApiClient "github-stats-metrics/infrastructure/github_api"
	presenter "github-stats-metrics/presentation/pull_request"
	"net/http"
)

func GetPullRequests(w http.ResponseWriter, r *http.Request) {

	pullRequests := githubApiClient.Fetch()

	presenter.Success(w, pullRequests)
}
