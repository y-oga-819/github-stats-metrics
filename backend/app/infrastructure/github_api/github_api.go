package github_api

import (
	"context"
	"errors"
	"fmt"
	prDomain "github-stats-metrics/domain/pull_request"
	"github-stats-metrics/shared/config"
	"github-stats-metrics/shared/logger"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// repository はprDomain.Repositoryインターフェースの実装
type repository struct {
	client *githubv4.Client
	config *config.Config
	logger *logger.LevelLogger
}

// NewRepository はGitHub APIを使用するRepository実装を作成
func NewRepository(cfg *config.Config) prDomain.Repository {
	client, err := createClient(cfg)
	levelLogger := logger.NewLevelLogger()
	
	if err != nil {
		levelLogger.Error("Failed to create GitHub client", "error", err)
		// エラーを含むリポジトリを返す（実行時にエラーを返す）
		return &repository{client: nil, config: cfg, logger: levelLogger}
	}
	
	levelLogger.Info("GitHub API client initialized successfully")
	return &repository{
		client: client,
		config: cfg,
		logger: levelLogger,
	}
}

func createClient(cfg *config.Config) (*githubv4.Client, error) {
	// 設定からGitHubトークンを取得
	token := cfg.GitHub.Token
	if token == "" {
		return nil, errors.New("GitHub token is not configured")
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
		"query":      githubv4.String(r.createQuery(queryParametes.StartDate, queryParametes.EndDate, queryParametes.Developers)),
	}

	array := make([]prDomain.PullRequest, 0, 1)
	prCount := 0

	retryCount := 0
	maxRetries := 3

	for {
		// レート制限チェックと待機
		if err := r.checkRateLimit(ctx, &query); err != nil {
			return nil, fmt.Errorf("rate limit check failed: %w", err)
		}

		// GithubAPIv4にアクセス（指数バックオフ付きリトライ）
		err := r.queryWithRetry(ctx, client, &query, variables, &retryCount, maxRetries)
		if err != nil {
			return nil, r.handleGitHubAPIError(err)
		}

		// 検索結果をDomainモデルに変換
		for _, node := range query.Search.Nodes {
			domainPR := convertToDomain(node.Pr)
			array = append(array, domainPR)
		}

		// 取得数をカウント
		prCount += len(query.Search.Nodes)

		// API情報をデバッグレベルで表示
		r.logger.Debug("GitHub API pagination info",
			"hasNextPage", query.Search.PageInfo.HasNextPage,
			"endCursor", query.Search.PageInfo.EndCursor,
			"prCount", prCount)

		// レート制限情報をログに記録
		r.logRateLimitInfo(query.RateLimit)

		// API LIMIT などのデータをデバッグ表示（デバッグモード時のみ）
		if r.config.IsDebugMode() {
			fmt.Print("\n------------------------------------------------------------\n")
			fmt.Printf("HasNextPage: %t\n", query.Search.PageInfo.HasNextPage)
			fmt.Printf("EndCursor: %s\n", query.Search.PageInfo.EndCursor)
			fmt.Printf("RateLimit: %+v\n", query.RateLimit)
		}

		// データを全て取り切ったら終了
		if !query.Search.PageInfo.HasNextPage {
			r.logger.Info("GitHub API data fetch completed", "totalPullRequests", prCount)
			if r.config.IsDebugMode() {
				fmt.Printf("取得したPR数: %d\n", prCount)
			}
			break
		}

		// 次のページへ
		variables["cursor"] = githubv4.NewString(query.Search.PageInfo.EndCursor)
		
		// ページ間の適切な間隔を設ける
		time.Sleep(100 * time.Millisecond)
	}

	return array, nil
}

// GitHub API v4 にリクエストするクエリの検索条件文字列を生成する
func (r *repository) createQuery(startDate string, endDate string, developers []string) string {
	// 期間
	query := fmt.Sprintf("merged:%s..%s ", startDate, endDate)

	// リポジトリ（設定から取得）
	repositories := r.config.GetCleanRepositories()
	query += "repo:" + strings.Join(repositories, " repo:") + " "

	// 開発者
	query += "author:" + strings.Join(developers, " author:")

	// デバッグレベルでクエリを出力
	r.logger.Debug("GitHub GraphQL query generated", "query", query)
	
	if r.config.IsDebugMode() {
		fmt.Println("GitHub query:", query)
	}
	return query
}

// checkRateLimit はAPI呼び出し前にレート制限をチェック
func (r *repository) checkRateLimit(ctx context.Context, query *graphqlQuery) error {
	// 事前にレート制限情報を取得
	tempQuery := struct {
		RateLimit struct {
			Cost      githubv4.Int
			Limit     githubv4.Int
			Remaining githubv4.Int
			ResetAt   githubv4.DateTime
		}
	}{}
	
	if err := r.client.Query(ctx, &tempQuery, nil); err != nil {
		r.logger.Warn("Failed to check rate limit", "error", err)
		return nil // レート制限チェックに失敗してもAPIコールは続行
	}
	
	remaining := int(tempQuery.RateLimit.Remaining)
	resetAt := tempQuery.RateLimit.ResetAt.Time
	
	// レート制限が少ない場合は警告
	if remaining < 100 {
		r.logger.Warn("GitHub API rate limit is low", 
			"remaining", remaining, 
			"resetAt", resetAt.Format(time.RFC3339))
		
		// レート制限が極端に少ない場合は待機
		if remaining < 10 {
			waitDuration := time.Until(resetAt)
			if waitDuration > 0 && waitDuration < time.Hour {
				r.logger.Error("Rate limit critically low, waiting until reset",
					"waitDuration", waitDuration,
					"resetAt", resetAt.Format(time.RFC3339))
				
				select {
				case <-time.After(waitDuration):
					r.logger.Info("Rate limit reset, continuing API calls")
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
	
	return nil
}

// queryWithRetry は指数バックオフ付きでクエリを実行
func (r *repository) queryWithRetry(ctx context.Context, client *githubv4.Client, query *graphqlQuery, variables map[string]interface{}, retryCount *int, maxRetries int) error {
	for *retryCount <= maxRetries {
		err := client.Query(ctx, query, variables)
		if err == nil {
			*retryCount = 0 // 成功時はリトライカウントをリセット
			return nil
		}
		
		// リトライ可能なエラーかチェック
		if !r.isRetryableError(err) {
			return err
		}
		
		*retryCount++
		if *retryCount > maxRetries {
			return fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, err)
		}
		
		// 指数バックオフで待機
		backoffDuration := time.Duration(math.Pow(2, float64(*retryCount-1))) * time.Second
		r.logger.Warn("API call failed, retrying with backoff",
			"attempt", *retryCount,
			"maxRetries", maxRetries,
			"backoffDuration", backoffDuration,
			"error", err)
		
		select {
		case <-time.After(backoffDuration):
			// 待機完了、リトライ
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	return fmt.Errorf("unexpected error in retry loop")
}

// isRetryableError はエラーがリトライ可能かを判定
func (r *repository) isRetryableError(err error) bool {
	errStr := strings.ToLower(err.Error())
	
	retryableErrors := []string{
		"rate limit",
		"timeout",
		"connection reset",
		"temporary failure",
		"server error",
		"503",
		"502",
		"500",
	}
	
	for _, retryable := range retryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}
	
	return false
}

// logRateLimitInfo はレート制限情報をログに記録
func (r *repository) logRateLimitInfo(rateLimit struct {
	Cost      githubv4.Int
	Limit     githubv4.Int
	Remaining githubv4.Int
	ResetAt   githubv4.DateTime
}) {
	remaining := int(rateLimit.Remaining)
	limit := int(rateLimit.Limit)
	cost := int(rateLimit.Cost)
	resetAt := rateLimit.ResetAt.Time
	
	// レート制限の使用率を計算
	usagePercent := float64(limit-remaining) / float64(limit) * 100
	
	r.logger.Debug("GitHub API Rate Limit status",
		"usagePercent", fmt.Sprintf("%.1f%%", usagePercent),
		"used", limit-remaining,
		"limit", limit,
		"cost", cost,
		"resetAt", resetAt.Format(time.RFC3339))
	
	// 警告レベルの判定
	if remaining < 500 {
		r.logger.Warn("GitHub API rate limit is getting low", 
			"remaining", remaining,
			"resetAt", resetAt.Format(time.RFC3339))
	}
}

// GetPullRequestByID は特定IDのPull Requestを取得
func (r *repository) GetPullRequestByID(ctx context.Context, id string) (*prDomain.PullRequest, error) {
	if r.client == nil {
		return nil, errors.New("GitHub client is not initialized")
	}
	
	query := ReviewMetricsQuery{}
	variables := map[string]interface{}{
		"prId": githubv4.ID(id),
	}
	
	if err := r.client.Query(ctx, &query, variables); err != nil {
		return nil, r.handleGitHubAPIError(err)
	}
	
	domainPR := convertExtendedToDomain(query.Node.PullRequest)
	return &domainPR, nil
}

// GetPullRequestWithMetrics は詳細メトリクス付きでPRを取得
func (r *repository) GetPullRequestWithMetrics(ctx context.Context, id string) (*prDomain.PRMetrics, error) {
	if r.client == nil {
		return nil, errors.New("GitHub client is not initialized")
	}
	
	query := ReviewMetricsQuery{}
	variables := map[string]interface{}{
		"prId": githubv4.ID(id),
	}
	
	if err := r.client.Query(ctx, &query, variables); err != nil {
		return nil, r.handleGitHubAPIError(err)
	}
	
	prMetrics := convertToPRMetrics(query.Node.PullRequest)
	return prMetrics, nil
}

// GetFileDetails は特定PRのファイル詳細を取得
func (r *repository) GetFileDetails(ctx context.Context, prId string) ([]prDomain.FileChangeMetrics, error) {
	if r.client == nil {
		return nil, errors.New("GitHub client is not initialized")
	}
	
	var allFiles []prDomain.FileChangeMetrics
	cursor := (*githubv4.String)(nil)
	
	for {
		query := FileDetailsQuery{}
		variables := map[string]interface{}{
			"prId":   githubv4.ID(prId),
			"cursor": cursor,
		}
		
		if err := r.client.Query(ctx, &query, variables); err != nil {
			return nil, r.handleGitHubAPIError(err)
		}
		
		// ファイル情報を変換
		for _, file := range query.Node.PullRequest.Files.Nodes {
			fileMetrics := prDomain.FileChangeMetrics{
				FileName:     string(file.Path),
				FileType:     getFileExtension(string(file.Path)),
				LinesAdded:   int(file.Additions),
				LinesDeleted: int(file.Deletions),
				IsNewFile:    string(file.ChangeType) == "ADDED",
				IsDeleted:    string(file.ChangeType) == "DELETED",
				IsRenamed:    string(file.ChangeType) == "RENAMED",
			}
			allFiles = append(allFiles, fileMetrics)
		}
		
		if !query.Node.PullRequest.Files.PageInfo.HasNextPage {
			break
		}
		
		cursor = githubv4.NewString(query.Node.PullRequest.Files.PageInfo.EndCursor)
	}
	
	return allFiles, nil
}

// GetReviewTimeline は特定PRのレビュータイムラインを取得
func (r *repository) GetReviewTimeline(ctx context.Context, prId string) ([]prDomain.ReviewEvent, error) {
	if r.client == nil {
		return nil, errors.New("GitHub client is not initialized")
	}
	
	var allEvents []prDomain.ReviewEvent
	cursor := (*githubv4.String)(nil)
	
	for {
		query := ReviewTimelineQuery{}
		variables := map[string]interface{}{
			"prId":   githubv4.ID(prId),
			"cursor": cursor,
		}
		
		if err := r.client.Query(ctx, &query, variables); err != nil {
			return nil, r.handleGitHubAPIError(err)
		}
		
		// タイムラインイベントを変換
		for _, item := range query.Node.PullRequest.TimelineItems.Nodes {
			if !item.ReviewRequestedEvent.CreatedAt.Time.IsZero() {
				event := prDomain.ReviewEvent{
					Type:      prDomain.ReviewEventTypeRequested,
					CreatedAt: item.ReviewRequestedEvent.CreatedAt.Time,
					Actor:     string(item.ReviewRequestedEvent.Actor.Login),
					Reviewer:  string(item.ReviewRequestedEvent.RequestedReviewer.User.Login),
				}
				allEvents = append(allEvents, event)
			}
			
			if !item.PullRequestReview.CreatedAt.Time.IsZero() {
				event := prDomain.ReviewEvent{
					Type:      convertReviewState(item.PullRequestReview.State),
					CreatedAt: item.PullRequestReview.CreatedAt.Time,
					Actor:     string(item.PullRequestReview.Author.Login),
					Reviewer:  string(item.PullRequestReview.Author.Login),
				}
				allEvents = append(allEvents, event)
			}
			
			if !item.ReadyForReviewEvent.CreatedAt.Time.IsZero() {
				event := prDomain.ReviewEvent{
					Type:      prDomain.ReviewEventTypeReadyForReview,
					CreatedAt: item.ReadyForReviewEvent.CreatedAt.Time,
					Actor:     string(item.ReadyForReviewEvent.Actor.Login),
				}
				allEvents = append(allEvents, event)
			}
			
			if !item.MergedEvent.CreatedAt.Time.IsZero() {
				event := prDomain.ReviewEvent{
					Type:      prDomain.ReviewEventTypeMerged,
					CreatedAt: item.MergedEvent.CreatedAt.Time,
					Actor:     string(item.MergedEvent.Actor.Login),
				}
				allEvents = append(allEvents, event)
			}
		}
		
		if !query.Node.PullRequest.TimelineItems.PageInfo.HasNextPage {
			break
		}
		
		cursor = githubv4.NewString(query.Node.PullRequest.TimelineItems.PageInfo.EndCursor)
	}
	
	return allEvents, nil
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
