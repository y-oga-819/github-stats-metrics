package github_api

import (
	developerDomain "github-stats-metrics/domain/developer"
	prDomain "github-stats-metrics/domain/pull_request"
)

// GithubAPIv4の解説：https://zenn.dev/hsaki/articles/github-graphql
// Golangで GithubAPIv4を使うならこのライブラリを使う：https://github.com/shurcooL/githubv4
func Fetch() []prDomain.PullRequest {
	array := make([]prDomain.PullRequest, 0, 1)

	count := 1
	for i := 0; i < count; i++ {
		array = append(array, prDomain.PullRequest{
			Id:     "prId",
			Title:  "テストPR名",
			Status: "Opened",
			Author: developerDomain.Developer{
				Id:         "y-oga-819",
				ScreenName: "いらないかも",
				ImageURL:   "https://avatars.githubusercontent.com/u/6323203?v=4",
			},
			Opened:        time.Now(),
			FirstReviewed: time.Now(),
			Approved:      time.Now(),
			Merged:        time.Now(),
		})
	}

	return array
}
