package github_api

import (
	"context"
	"fmt"
	developerDomain "github-stats-metrics/domain/developer"
	prDomain "github-stats-metrics/domain/pull_request"
	"log"
	"os"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GithubAPIv4の解説：https://zenn.dev/hsaki/articles/github-graphql
// Golangで GithubAPIv4を使うならこのライブラリを使う：https://github.com/shurcooL/githubv4
func Fetch() []prDomain.PullRequest {
	// 認証トークンを使ったクライアントを生成する
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	// クエリを構築
	var query struct {
		Viewer struct {
			CreatedPullRequests struct {
				Nodes    []PullRequest
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"createdPullRequests(first: 30, after: $cursor)"`
		}
	}

	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil),
	}

	// GithubAPIv4にアクセス
	for {
		// Execute the GraphQL query
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			log.Fatal(err)
		}

		// Print information about each PullRequest
		for _, pr := range query.Viewer.CreatedPullRequests.Nodes {
			fmt.Printf("Title: %s\n", pr.Title)
			fmt.Printf("URL: %s\n", pr.URL)
			fmt.Println("-----")
		}

		// Check if there are more pages
		if !query.Viewer.CreatedPullRequests.PageInfo.HasNextPage {
			break
		}

		// Set the cursor for the next page
		variables["cursor"] = githubv4.NewString(query.Viewer.CreatedPullRequests.PageInfo.EndCursor)
	}

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

// Define the structure to represent a PullRequest
type PullRequest struct {
	Title githubv4.String
	URL   githubv4.URI
}
