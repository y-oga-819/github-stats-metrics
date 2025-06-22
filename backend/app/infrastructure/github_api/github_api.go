package github_api

import (
	"context"
	"errors"
	"fmt"
	prDomain "github-stats-metrics/domain/pull_request"
	"log"
	"os"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// repository はprDomain.Repositoryインターフェースの実装
type repository struct {
	client *githubv4.Client
}

// NewRepository はGitHub APIを使用するRepository実装を作成
func NewRepository() prDomain.Repository {
	client, err := createClient()
	if err != nil {
		log.Printf("Failed to create GitHub client: %v", err)
		// エラーを含むリポジトリを返す（実行時にエラーを返す）
		return &repository{client: nil}
	}
	return &repository{
		client: client,
	}
}

func createClient() (*githubv4.Client, error) {
	// GitHubトークンの存在確認
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("GITHUB_TOKEN environment variable is not set")
	}
	
	// 認証トークンを使ったクライアントを生成する
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return githubv4.NewClient(httpClient), nil
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

// GetPullRequests はGitHub APIからPull Requestsを取得
func (r *repository) GetPullRequests(ctx context.Context, req prDomain.GetPullRequestsRequest) ([]prDomain.PullRequest, error) {
	// クライアント初期化チェック
	if r.client == nil {
		return nil, errors.New("GitHub client is not initialized - check GITHUB_TOKEN environment variable")
	}
	
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	return r.fetchPullRequests(ctx, req)
}

// fetchPullRequests は実際のGitHub API呼び出しを実行
func (r *repository) fetchPullRequests(ctx context.Context, queryParametes prDomain.GetPullRequestsRequest) ([]prDomain.PullRequest, error) {
	// 既に初期化済みのクライアントを使用
	client := r.client

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
		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, r.handleGitHubAPIError(err)
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

	return array, nil
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

// GetPullRequestByID は特定IDのPull Requestを取得
func (r *repository) GetPullRequestByID(ctx context.Context, id string) (*prDomain.PullRequest, error) {
	// 実装は今後必要に応じて追加
	return nil, fmt.Errorf("not implemented yet")
}

// GetRepositories は対象リポジトリ一覧を取得
func (r *repository) GetRepositories(ctx context.Context) ([]string, error) {
	repositoriesStr := os.Getenv("GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES")
	if repositoriesStr == "" {
		return nil, fmt.Errorf("GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES environment variable is not set")
	}
	
	repositories := strings.Split(repositoriesStr, ",")
	for i, repo := range repositories {
		repositories[i] = strings.TrimSpace(repo)
	}
	
	return repositories, nil
}

// GetDevelopers は開発者一覧を取得
func (r *repository) GetDevelopers(ctx context.Context, repositories []string) ([]string, error) {
	// 実際の実装では、GitHub APIから開発者一覧を取得
	// 現在は設定ベースで実装
	return []string{"developer1", "developer2", "developer3"}, nil
}

// handleGitHubAPIError はGitHub APIエラーを適切に分類して返す
func (r *repository) handleGitHubAPIError(err error) error {
	errorMsg := err.Error()
	
	// 認証エラーの判定
	if strings.Contains(errorMsg, "401") || 
	   strings.Contains(errorMsg, "Bad credentials") ||
	   strings.Contains(errorMsg, "requires authentication") {
		log.Printf("GitHub authentication failed: %v", err)
		return fmt.Errorf("GitHub authentication failed - check GITHUB_TOKEN validity: %w", err)
	}
	
	// レート制限エラーの判定
	if strings.Contains(errorMsg, "rate limit") || 
	   strings.Contains(errorMsg, "403") ||
	   strings.Contains(errorMsg, "API rate limit exceeded") {
		log.Printf("GitHub rate limit exceeded: %v", err)
		return fmt.Errorf("GitHub rate limit exceeded - please wait and try again: %w", err)
	}
	
	// 権限エラーの判定
	if strings.Contains(errorMsg, "forbidden") || 
	   strings.Contains(errorMsg, "access denied") {
		log.Printf("GitHub access denied: %v", err)
		return fmt.Errorf("GitHub access denied - check repository permissions: %w", err)
	}
	
	// ネットワーク関連エラーの判定
	if strings.Contains(errorMsg, "connection") || 
	   strings.Contains(errorMsg, "timeout") ||
	   strings.Contains(errorMsg, "network") {
		log.Printf("GitHub network error: %v", err)
		return fmt.Errorf("GitHub network error - check internet connection: %w", err)
	}
	
	// その他のエラー
	log.Printf("GitHub API error: %v", err)
	return fmt.Errorf("GitHub API error: %w", err)
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
