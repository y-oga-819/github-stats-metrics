# フェーズ2: Repository Interfaceパターンの導入

## 概要
Clean Architectureにおける依存関係逆転の原則を実現するため、Repository Interfaceパターンを導入します。これにより、Application層がInfrastructure層の具象実装に依存しないアーキテクチャを構築します。

## 現状の問題

### 依存関係の問題
```
Application層 → Infrastructure層（直接依存）
```

**ファイル**: `application/pull_request/get_pull_requests.go:6`
```go
githubApiClient "github-stats-metrics/infrastructure/github_api"
```

### 問題点
1. **依存関係逆転違反**: 上位層が下位層の具象実装に依存
2. **テスト困難**: 実際のGitHub APIを呼び出さないとテストできない
3. **実装の変更困難**: GitHub API以外のデータソースに変更する際の影響大
4. **Interface Segregation違反**: 必要以上の機能に依存

## 改善案

### Repository Interfaceの定義

#### `domain/pull_request/repository.go` （新規作成）
```go
package pull_request

import (
    "context"
)

// Repository は Pull Request データの永続化・取得を抽象化
type Repository interface {
    // GetPullRequests は指定条件でPull Requestsを取得
    GetPullRequests(ctx context.Context, req GetPullRequestsRequest) ([]PullRequest, error)
    
    // GetPullRequestByID は特定のIDでPull Requestを取得
    GetPullRequestByID(ctx context.Context, id string) (*PullRequest, error)
    
    // 将来の拡張性を考慮した追加メソッド
    // CountPullRequests(ctx context.Context, req GetPullRequestsRequest) (int, error)
    // GetPullRequestMetrics(ctx context.Context, req GetPullRequestsRequest) (*Metrics, error)
}

// Metrics はPull Requestの集計情報
type Metrics struct {
    TotalCount       int
    AverageReviewTime float64
    AverageApprovalTime float64
    AverageMergeTime float64
}
```

### Application層の改善

#### `application/pull_request/usecase.go` （新規作成）
```go
package pull_request

import (
    "context"
    "fmt"
    
    domain "github-stats-metrics/domain/pull_request"
)

// UseCase はPull Request関連のビジネスロジックを処理
type UseCase struct {
    repo domain.Repository
}

// NewUseCase はUseCaseのコンストラクタ
func NewUseCase(repo domain.Repository) *UseCase {
    return &UseCase{
        repo: repo,
    }
}

// GetPullRequests はPull Requestsを取得し、ビジネスルールを適用
func (uc *UseCase) GetPullRequests(ctx context.Context, req domain.GetPullRequestsRequest) ([]domain.PullRequest, error) {
    // リクエストバリデーション
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // リポジトリから取得
    pullRequests, err := uc.repo.GetPullRequests(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to get pull requests: %w", err)
    }
    
    // ビジネスルールの適用（例：epicブランチの除外）
    filtered := make([]domain.PullRequest, 0, len(pullRequests))
    for _, pr := range pullRequests {
        if !uc.shouldExcludePR(pr) {
            filtered = append(filtered, pr)
        }
    }
    
    return filtered, nil
}

// shouldExcludePR はPRを除外すべきかを判定（ビジネスルール）
func (uc *UseCase) shouldExcludePR(pr domain.PullRequest) bool {
    // epicブランチは除外
    if pr.HeadRefName != "" && len(pr.HeadRefName) > 5 && pr.HeadRefName[:5] == "epic/" {
        return true
    }
    
    // その他の除外ルールがあれば追加
    return false
}

// GetPullRequestMetrics はPull Requestの統計情報を計算
func (uc *UseCase) GetPullRequestMetrics(ctx context.Context, req domain.GetPullRequestsRequest) (*domain.Metrics, error) {
    pullRequests, err := uc.GetPullRequests(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return uc.calculateMetrics(pullRequests), nil
}

// calculateMetrics は統計情報を計算
func (uc *UseCase) calculateMetrics(pullRequests []domain.PullRequest) *domain.Metrics {
    if len(pullRequests) == 0 {
        return &domain.Metrics{}
    }
    
    var totalReviewTime, totalApprovalTime, totalMergeTime float64
    reviewCount, approvalCount, mergeCount := 0, 0, 0
    
    for _, pr := range pullRequests {
        if reviewTime := pr.ReviewTime(); reviewTime != nil {
            totalReviewTime += reviewTime.Seconds()
            reviewCount++
        }
        
        if approvalTime := pr.ApprovalTime(); approvalTime != nil {
            totalApprovalTime += approvalTime.Seconds()
            approvalCount++
        }
        
        if mergeTime := pr.MergeTime(); mergeTime != nil {
            totalMergeTime += mergeTime.Seconds()
            mergeCount++
        }
    }
    
    metrics := &domain.Metrics{
        TotalCount: len(pullRequests),
    }
    
    if reviewCount > 0 {
        metrics.AverageReviewTime = totalReviewTime / float64(reviewCount)
    }
    if approvalCount > 0 {
        metrics.AverageApprovalTime = totalApprovalTime / float64(approvalCount)
    }
    if mergeCount > 0 {
        metrics.AverageMergeTime = totalMergeTime / float64(mergeCount)
    }
    
    return metrics
}
```

### Infrastructure層の実装

#### `infrastructure/github_api/repository.go` （新規作成）
```go
package github_api

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    
    domain "github-stats-metrics/domain/pull_request"
    "github.com/shurcooL/githubv4"
    "golang.org/x/oauth2"
)

// repository はGitHub API v4を使用したRepository interfaceの実装
type repository struct {
    client *githubv4.Client
}

// NewRepository はGitHub API repositoryのコンストラクタ
func NewRepository() domain.Repository {
    return &repository{
        client: createClient(),
    }
}

// GetPullRequests はGitHub APIからPull Requestsを取得
func (r *repository) GetPullRequests(ctx context.Context, req domain.GetPullRequestsRequest) ([]domain.PullRequest, error) {
    query := graphqlQuery{}
    variables := map[string]interface{}{
        "searchType": githubv4.SearchTypeIssue,
        "cursor":     (*githubv4.String)(nil),
        "query":      githubv4.String(r.createQuery(req)),
    }
    
    var allPRs []domain.PullRequest
    prCount := 0
    
    for {
        // GraphQL API呼び出し
        if err := r.client.Query(ctx, &query, variables); err != nil {
            return nil, fmt.Errorf("github api query failed: %w", err)
        }
        
        // レスポンスをDomainモデルに変換
        for _, node := range query.Search.Nodes {
            domainPR := r.convertToDomain(node.Pr)
            allPRs = append(allPRs, domainPR)
        }
        
        prCount += len(query.Search.Nodes)
        
        // レート制限チェック
        if query.RateLimit.Remaining == 0 {
            log.Printf("GitHub API rate limit reached. Reset at: %v", query.RateLimit.ResetAt)
            break
        }
        
        // ページネーション処理
        if !query.Search.PageInfo.HasNextPage {
            break
        }
        
        variables["cursor"] = githubv4.NewString(query.Search.PageInfo.EndCursor)
    }
    
    log.Printf("Retrieved %d pull requests from GitHub API", prCount)
    return allPRs, nil
}

// GetPullRequestByID は特定IDのPull Requestを取得
func (r *repository) GetPullRequestByID(ctx context.Context, id string) (*domain.PullRequest, error) {
    // 実装は今後必要に応じて追加
    return nil, fmt.Errorf("not implemented")
}

// createQuery はGitHub GraphQL検索クエリを生成
func (r *repository) createQuery(req domain.GetPullRequestsRequest) string {
    query := fmt.Sprintf("merged:%s..%s ", req.StartDate, req.EndDate)
    
    // リポジトリ設定
    repositories := strings.Split(os.Getenv("GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES"), ",")
    query += "repo:" + strings.Join(repositories, " repo:") + " "
    
    // 開発者設定
    query += "author:" + strings.Join(req.Developers, " author:")
    
    return query
}

// convertToDomain はGitHub APIレスポンスをDomainモデルに変換
func (r *repository) convertToDomain(apiPR githubv4PullRequest) domain.PullRequest {
    pr := domain.PullRequest{
        ID:          string(apiPR.Id),
        Number:      int(apiPR.Number),
        Title:       string(apiPR.Title),
        BaseRefName: string(apiPR.BaseRefName),
        HeadRefName: string(apiPR.HeadRefName),
        Author: domain.Author{
            Login:     string(apiPR.Author.Login),
            AvatarURL: string(apiPR.Author.AvatarURL),
        },
        Repository: domain.Repository{
            Name: string(apiPR.Repository.Name),
        },
        URL:       string(apiPR.URL),
        Additions: int(apiPR.Additions),
        Deletions: int(apiPR.Deletions),
        CreatedAt: apiPR.CreatedAt.Time,
    }
    
    // オプション値の変換
    if !apiPR.MergedAt.Time.IsZero() {
        pr.MergedAt = &apiPR.MergedAt.Time
    }
    
    if len(apiPR.FirstReviewed.Nodes) > 0 && !apiPR.FirstReviewed.Nodes[0].CreatedAt.Time.IsZero() {
        pr.FirstReviewed = &apiPR.FirstReviewed.Nodes[0].CreatedAt.Time
    }
    
    if len(apiPR.LastApprovedAt.Nodes) > 0 && !apiPR.LastApprovedAt.Nodes[0].CreatedAt.Time.IsZero() {
        pr.LastApproved = &apiPR.LastApprovedAt.Nodes[0].CreatedAt.Time
    }
    
    return pr
}

// createClient はGitHub APIクライアントを作成
func createClient() *githubv4.Client {
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        log.Fatal("GITHUB_TOKEN environment variable is required")
    }
    
    src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    httpClient := oauth2.NewClient(context.Background(), src)
    return githubv4.NewClient(httpClient)
}

// GitHub API用の内部構造体（外部に公開しない）
type githubv4PullRequest struct {
    Id          githubv4.String
    Number      githubv4.Int
    Title       githubv4.String
    BaseRefName githubv4.String
    HeadRefName githubv4.String
    Author      struct {
        Login     githubv4.String
        AvatarURL githubv4.URI `graphql:"avatarUrl(size:72)"`
    }
    Repository struct {
        Name githubv4.String
    }
    URL           githubv4.URI
    Additions     githubv4.Int
    Deletions     githubv4.Int
    CreatedAt     githubv4.DateTime
    FirstReviewed struct {
        Nodes []struct {
            CreatedAt githubv4.DateTime
        }
    } `graphql:"FirstReviewed: reviews(first: 1)"`
    LastApprovedAt struct {
        Nodes []struct {
            CreatedAt githubv4.DateTime
        }
    } `graphql:"LastApprovedAt: reviews(last: 1, states: APPROVED)"`
    MergedAt githubv4.DateTime
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
```

### テスト実装例

#### `application/pull_request/usecase_test.go` （新規作成）
```go
package pull_request

import (
    "context"
    "testing"
    "time"
    
    domain "github-stats-metrics/domain/pull_request"
)

// MockRepository はテスト用のRepository実装
type MockRepository struct {
    pullRequests []domain.PullRequest
    err          error
}

func (m *MockRepository) GetPullRequests(ctx context.Context, req domain.GetPullRequestsRequest) ([]domain.PullRequest, error) {
    return m.pullRequests, m.err
}

func (m *MockRepository) GetPullRequestByID(ctx context.Context, id string) (*domain.PullRequest, error) {
    return nil, nil
}

func TestUseCase_GetPullRequests(t *testing.T) {
    // テストデータ
    now := time.Now()
    mockPRs := []domain.PullRequest{
        {
            ID:          "1",
            HeadRefName: "feature/test",
            CreatedAt:   now,
        },
        {
            ID:          "2",
            HeadRefName: "epic/major-feature", // 除外対象
            CreatedAt:   now,
        },
    }
    
    mockRepo := &MockRepository{pullRequests: mockPRs}
    uc := NewUseCase(mockRepo)
    
    req := domain.GetPullRequestsRequest{
        StartDate:  "2023-01-01",
        EndDate:    "2023-01-31",
        Developers: []string{"user1"},
    }
    
    result, err := uc.GetPullRequests(context.Background(), req)
    
    // 検証
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    // epicブランチが除外されているか確認
    if len(result) != 1 {
        t.Fatalf("expected 1 PR, got %d", len(result))
    }
    
    if result[0].ID != "1" {
        t.Fatalf("expected PR ID 1, got %s", result[0].ID)
    }
}
```

## 実装チェックリスト

- [ ] Repository Interfaceの定義
- [ ] UseCase層の実装
- [ ] Infrastructure層でのInterface実装
- [ ] 依存性注入の設定
- [ ] Mock実装とテストの作成
- [ ] 既存コードからの移行
- [ ] 動作確認とテスト実行

## 期待される効果

1. **テスタビリティ向上**: Mock実装による独立したテスト実行
2. **実装の柔軟性**: 異なるデータソースへの切り替えが容易
3. **依存関係の明確化**: Interface を通じた疎結合
4. **開発効率向上**: 並行開発が可能（Interface合意後）

## 移行戦略

1. **Interface定義**: Repository Interfaceを先に定義
2. **Mock実装**: テスト用のMock Repository作成
3. **UseCase実装**: 新しいUseCase層の実装
4. **Infrastructure実装**: 既存ロジックをInterface実装に移行
5. **段階的置換**: Handler → UseCase → Repository の順で置換

## リスク対策

1. **パフォーマンス**: Interface呼び出しによるオーバーヘッド測定
2. **複雑性増加**: 適切なドキュメント作成と例外処理
3. **移行中の不整合**: 段階的移行による影響範囲の限定