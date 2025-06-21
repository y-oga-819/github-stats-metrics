# フェーズ1: Domain層の純粋化

## 概要
Clean Architectureの最も重要な原則の一つは、Domain層（ビジネスルール）が外部の詳細に依存しないことです。現在のDomain層は`githubv4`ライブラリに直接依存しており、これを解消する必要があります。

## 現状の問題

### ファイル: `domain/pull_request/pull_request.go`
```go
package pull_request

import "github.com/shurcooL/githubv4" // ← 外部ライブラリ依存

type PullRequest struct {
    Id          githubv4.String      // ← 外部型依存
    Number      githubv4.Int         // ← 外部型依存
    Title       githubv4.String      // ← 外部型依存
    BaseRefName githubv4.String      // ← 外部型依存
    HeadRefName githubv4.String      // ← 外部型依存
    Author      struct {
        Login     githubv4.String    // ← 外部型依存
        AvatarURL githubv4.URI       // ← 外部型依存
    }
    Repository struct {
        Name githubv4.String        // ← 外部型依存
    }
    URL           githubv4.URI       // ← 外部型依存
    Additions     githubv4.Int       // ← 外部型依存
    Deletions     githubv4.Int       // ← 外部型依存
    CreatedAt     githubv4.DateTime  // ← 外部型依存
    FirstReviewed struct {
        Nodes []struct {
            CreatedAt githubv4.DateTime // ← 外部型依存
        }
    } `graphql:"FirstReviewed: reviews(first: 1)"`
    LastApprovedAt struct {
        Nodes []struct {
            CreatedAt githubv4.DateTime // ← 外部型依存
        }
    } `graphql:"LastApprovedAt: reviews(last: 1, states: APPROVED)"`
    MergedAt githubv4.DateTime      // ← 外部型依存
}
```

### 問題点
1. **外部ライブラリ依存**: Domain層がGitHub APIライブラリに依存
2. **GraphQL構造の露出**: GraphQLのクエリ構造がドメインモデルに漏れている
3. **テスト困難**: 外部ライブラリの型を使用するためモックが困難
4. **変更影響範囲拡大**: GitHubクライアントライブラリ変更時にDomain層も影響

## 改善案

### 新しいDomain層の実装

#### `domain/pull_request/pull_request.go`
```go
package pull_request

import (
    "time"
)

type PullRequest struct {
    ID          string
    Number      int
    Title       string
    BaseRefName string
    HeadRefName string
    Author      Author
    Repository  Repository
    URL         string
    Additions   int
    Deletions   int
    CreatedAt   time.Time
    FirstReviewed *time.Time
    LastApproved  *time.Time
    MergedAt     *time.Time
}

type Author struct {
    Login     string
    AvatarURL string
}

type Repository struct {
    Name string
}

// ドメインロジック
func (pr PullRequest) IsReviewed() bool {
    return pr.FirstReviewed != nil
}

func (pr PullRequest) IsApproved() bool {
    return pr.LastApproved != nil
}

func (pr PullRequest) IsMerged() bool {
    return pr.MergedAt != nil
}

func (pr PullRequest) ReviewTime() *time.Duration {
    if pr.FirstReviewed == nil {
        return nil
    }
    duration := pr.FirstReviewed.Sub(pr.CreatedAt)
    return &duration
}

func (pr PullRequest) ApprovalTime() *time.Duration {
    if pr.FirstReviewed == nil || pr.LastApproved == nil {
        return nil
    }
    duration := pr.LastApproved.Sub(*pr.FirstReviewed)
    return &duration
}

func (pr PullRequest) MergeTime() *time.Duration {
    if pr.LastApproved == nil || pr.MergedAt == nil {
        return nil
    }
    duration := pr.MergedAt.Sub(*pr.LastApproved)
    return &duration
}
```

#### `domain/pull_request/get_pull_requests_request.go` （更新）
```go
package pull_request

import (
    "fmt"
    "time"
)

type GetPullRequestsRequest struct {
    StartDate  string   `schema:"startdate,required"`
    EndDate    string   `schema:"enddate,required"`
    Developers []string `schema:"developers,required"`
}

// バリデーションロジック
func (req GetPullRequestsRequest) Validate() error {
    if req.StartDate == "" {
        return fmt.Errorf("start date is required")
    }
    if req.EndDate == "" {
        return fmt.Errorf("end date is required")
    }
    if len(req.Developers) == 0 {
        return fmt.Errorf("at least one developer is required")
    }
    
    startDate, err := time.Parse("2006-01-02", req.StartDate)
    if err != nil {
        return fmt.Errorf("invalid start date format: %w", err)
    }
    
    endDate, err := time.Parse("2006-01-02", req.EndDate)
    if err != nil {
        return fmt.Errorf("invalid end date format: %w", err)
    }
    
    if startDate.After(endDate) {
        return fmt.Errorf("start date must be before end date")
    }
    
    return nil
}

func (req GetPullRequestsRequest) GetStartDate() (time.Time, error) {
    return time.Parse("2006-01-02", req.StartDate)
}

func (req GetPullRequestsRequest) GetEndDate() (time.Time, error) {
    return time.Parse("2006-01-02", req.EndDate)
}
```

### 移行戦略

#### ステップ1: 新しいDomainモデルの並行実装
1. 現在の`pull_request.go`を`pull_request_legacy.go`にリネーム
2. 新しい`pull_request.go`を実装
3. 両方のモデルが共存する状態を作る

#### ステップ2: Infrastructure層での変換実装
```go
// infrastructure/github_api/converter.go
package github_api

import (
    "time"
    domain "github-stats-metrics/domain/pull_request"
    "github.com/shurcooL/githubv4"
)

// GitHub APIのレスポンスをDomainモデルに変換
func convertToDomain(apiPR githubv4PullRequest) domain.PullRequest {
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
        MergedAt:  timePtr(apiPR.MergedAt.Time),
    }
    
    // FirstReviewed の変換
    if len(apiPR.FirstReviewed.Nodes) > 0 {
        pr.FirstReviewed = timePtr(apiPR.FirstReviewed.Nodes[0].CreatedAt.Time)
    }
    
    // LastApproved の変換
    if len(apiPR.LastApprovedAt.Nodes) > 0 {
        pr.LastApproved = timePtr(apiPR.LastApprovedAt.Nodes[0].CreatedAt.Time)
    }
    
    return pr
}

func timePtr(t time.Time) *time.Time {
    if t.IsZero() {
        return nil
    }
    return &t
}

// GitHub API用の構造体（Infrastructure層でのみ使用）
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
```

#### ステップ3: 段階的移行
1. Infrastructure層で変換処理を実装
2. Application層とPresentation層で新しいDomainモデルを使用
3. テスト実装と動作確認
4. Legacy構造体の削除

## 実装チェックリスト

- [ ] 新しいDomainモデルの実装
- [ ] バリデーションロジックの追加
- [ ] ビジネスロジックメソッドの実装
- [ ] Infrastructure層での変換処理実装
- [ ] 単体テストの作成
- [ ] 統合テストでの動作確認
- [ ] Legacy構造体の削除

## 期待される効果

1. **テスタビリティ向上**: 標準的なGo型のためモック作成が容易
2. **保守性向上**: 外部ライブラリ変更の影響を受けない
3. **ビジネスロジック明確化**: ドメインロジックがメソッドとして表現
4. **型安全性向上**: null値の適切な処理（ポインタ型使用）

## リスク対策

1. **データ変換エラー**: 変換処理の十分なテスト実装
2. **パフォーマンス影響**: 変換処理によるオーバーヘッドの測定
3. **機能デグレード**: 既存機能との互換性確認
4. **移行期間の複雑性**: 段階的移行による影響範囲の限定