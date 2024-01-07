package github_api

import (
	"context"
	"fmt"
	prDomain "github-stats-metrics/domain/pull_request"
	"log"
	"os"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func createClient() *githubv4.Client {
	// 認証トークンを使ったクライアントを生成する
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return githubv4.NewClient(httpClient)
}

type graphqlQuery struct {
	Search struct {
		CodeCount githubv4.Int
		PageInfo  struct {
			HasNextPage githubv4.Boolean
			EndCursor   githubv4.String
		}
		Nodes []struct {
			Pr prDomain.PullRequest `graphql:"... on PullRequest"`
		}
	} `graphql:"search(type: $searchType, first: 100, after: $cursor, query: $query)"`
	RateLimit struct {
		Cost      githubv4.Int
		Limit     githubv4.Int
		Remaining githubv4.Int
		ResetAt   githubv4.DateTime
	}
}

// GithubAPIv4の解説：https://zenn.dev/hsaki/articles/github-graphql
// Golangで GithubAPIv4を使うならこのライブラリを使う：https://github.com/shurcooL/githubv4
func Fetch(queryParametes prDomain.GetPullRequestsRequest) []prDomain.PullRequest {
	// 認証を通したHTTP Clientを作成
	client := createClient()

	// クエリを構築
	query := graphqlQuery{}
	variables := map[string]interface{}{
		"searchType": githubv4.SearchTypeIssue,
		"cursor":     (*githubv4.String)(nil),
		"query":      githubv4.String(os.Getenv("GITHUB_GRAPHQL_SEARCH_QUERY")),
	}

	array := make([]prDomain.PullRequest, 0, 1)
	prCount := 0

	for {
		// GithubAPIv4にアクセス
		if err := client.Query(context.Background(), &query, variables); err != nil {
			log.Fatal(err)
		}

		// 検索結果を詰め替え
		for _, node := range query.Search.Nodes {
			array = append(array, node.Pr)
			// debugPrintf(node.Pr)
		}

		// 取得数をカウント
		prCount += len(query.Search.Nodes)

		// API LIMIT などのデータをデバッグ表示
		fmt.Print("\n------------------------------------------------------------\n")
		fmt.Printf("HasNextPage: %t\n", query.Search.PageInfo.HasNextPage)
		fmt.Printf("EndCursor: %s\n", query.Search.PageInfo.EndCursor)
		fmt.Printf("RateLimit: %+v\n", query.RateLimit)

		// データを全て取り切ったら終了
		if !query.Search.PageInfo.HasNextPage {
			fmt.Printf("取得したPR数: %d\n", prCount)
			break
		}

		// まだ取れるなら取得済みデータまでカーソルを移動する
		variables["cursor"] = githubv4.NewString(query.Search.PageInfo.EndCursor)
	}

	return array
}

// PullRequest型の構造体の中身をデバッグ表示する
func debugPrintf(pr prDomain.PullRequest) {
	// fmt.Printf("%+v\n", pr)
	fmt.Print("------------------------------------------------------------\n")
	fmt.Printf("%s\n", pr.Title)
	fmt.Printf("%s\n", pr.URL)
	fmt.Printf("CreatedAt: %s\n", pr.CreatedAt)
	fmt.Printf("FirstReviewed: %s\n", pr.FirstReviewed.Nodes[0].CreatedAt)
	fmt.Printf("LastApprovedAt: %s\n", pr.LastApprovedAt.Nodes[0].CreatedAt)
	fmt.Printf("MergedAt: %s\n", pr.MergedAt)

}
