package github_api

import (
	"context"
	"fmt"
	prDomain "github-stats-metrics/domain/pull_request"
	"log"
	"os"
	"strings"

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
			Pr githubv4PullRequest `graphql:"... on PullRequest"`
		}
	} `graphql:"search(type: $searchType, first: 100, after: $cursor, query: $query)"`
	RateLimit struct {
		Cost      githubv4.Int
		Limit     githubv4.Int
		Remaining githubv4.Int
		ResetAt   githubv4.DateTime
	}
}

// Fetch はGitHub APIからPull Requestsを取得してDomainモデルに変換
func Fetch(queryParametes prDomain.GetPullRequestsRequest) []prDomain.PullRequest {
	// 認証を通したHTTP Clientを作成
	client := createClient()

	// クエリを構築
	query := graphqlQuery{}
	variables := map[string]interface{}{
		"searchType": githubv4.SearchTypeIssue,
		"cursor":     (*githubv4.String)(nil),
		"query":      githubv4.String(createQuery(queryParametes.StartDate, queryParametes.EndDate, queryParametes.Developers)),
	}

	array := make([]prDomain.PullRequest, 0, 1)
	prCount := 0

	for {
		// GithubAPIv4にアクセス
		if err := client.Query(context.Background(), &query, variables); err != nil {
			log.Printf("GitHub API query failed: %v", err)
			return nil
		}

		// 検索結果をDomainモデルに変換
		for _, node := range query.Search.Nodes {
			domainPR := convertToDomain(node.Pr)
			array = append(array, domainPR)
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

// GitHub API v4 にリクエストするクエリの検索条件文字列を生成する
func createQuery(startDate string, endDate string, developers []string) string {
	// 期間
	query := fmt.Sprintf("merged:%s..%s ", startDate, endDate)

	// リポジトリ
	repositories := strings.Split(os.Getenv("GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES"), ",")
	query += "repo:" + strings.Join(repositories, " repo:") + " "

	// 開発者
	query += "author:" + strings.Join(developers, " author:")

	fmt.Println(query)
	return query
}

// debugPrintf はPullRequest型の構造体の中身をデバッグ表示する
func debugPrintf(pr prDomain.PullRequest) {
	fmt.Print("------------------------------------------------------------\n")
	fmt.Printf("Title: %s\n", pr.Title)
	fmt.Printf("URL: %s\n", pr.URL)
	fmt.Printf("CreatedAt: %s\n", pr.CreatedAt)
	if pr.FirstReviewed != nil {
		fmt.Printf("FirstReviewed: %s\n", *pr.FirstReviewed)
	}
	if pr.LastApproved != nil {
		fmt.Printf("LastApproved: %s\n", *pr.LastApproved)
	}
	if pr.MergedAt != nil {
		fmt.Printf("MergedAt: %s\n", *pr.MergedAt)
	}
}
