# Clean Architecture改善計画

## 概要
現在のバックエンドは「レイヤー化アーキテクチャ」の状態であり、真のClean Architectureの原則に従っていない問題がある。本ドキュメントでは、Clean Architectureの原則に基づいた改善案を段階的に実装するための計画を記載する。

## 現状の問題点

### 依存関係の方向性
```
現状: Presentation → Application → Infrastructure
理想: Presentation → Application → Domain ← Infrastructure
```

### レイヤー構成の問題
- Domain層が外部ライブラリに依存している
- Repository Interfaceが存在しない  
- UseCase層がHTTPハンドラーの責任を持っている
- 依存関係逆転の原則に違反している

## 改善フェーズ

### フェーズ1: Domain層の純粋化 🔴高優先度
**対象**: `domain/pull_request/pull_request.go`

**現状の問題**:
```go
import "github.com/shurcooL/githubv4"

type PullRequest struct {
    Id          githubv4.String  // 外部ライブラリ依存
    Number      githubv4.Int
    // ...
}
```

**改善案**:
```go
type PullRequest struct {
    ID          string
    Number      int
    Title       string
    Author      Author
    Repository  Repository
    CreatedAt   time.Time
    MergedAt    *time.Time
}

type Author struct {
    Login     string
    AvatarURL string
}
```

### フェーズ2: Repository Interfaceの導入 🔴高優先度

**追加ファイル**: `domain/pull_request/repository.go`
```go
package pull_request

import "context"

type Repository interface {
    GetPullRequests(ctx context.Context, req GetPullRequestsRequest) ([]PullRequest, error)
}
```

### フェーズ3: UseCase層の責任分離 🔴高優先度

**現状**: 
- `application/pull_request/get_pull_requests.go`がHTTPハンドラーとして動作

**改善案**:
```go
// application/pull_request/usecase.go
type UseCase struct {
    repo Repository
}

func NewUseCase(repo Repository) *UseCase {
    return &UseCase{repo: repo}
}

func (uc *UseCase) GetPullRequests(ctx context.Context, req GetPullRequestsRequest) ([]PullRequest, error) {
    // バリデーション
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // リポジトリ呼び出し
    return uc.repo.GetPullRequests(ctx, req)
}
```

### フェーズ4: 依存関係の逆転実装 🔴高優先度

**Infrastructure層の実装**:
```go
// infrastructure/github_api/repository.go
type repository struct {
    client *githubv4.Client
}

func NewRepository() Repository {
    return &repository{
        client: createClient(),
    }
}

func (r *repository) GetPullRequests(ctx context.Context, req domain.GetPullRequestsRequest) ([]domain.PullRequest, error) {
    // GitHub API実装詳細
    // githubv4.PullRequest → domain.PullRequestの変換処理
}
```

**DI（依存性注入）の実装**:
```go
// cmd/main.go
func main() {
    // Infrastructure
    prRepo := github_api.NewRepository()
    
    // Application
    prUseCase := pull_request.NewUseCase(prRepo)
    
    // Presentation
    prHandler := pull_request.NewHandler(prUseCase)
    
    // Server setup
    server.StartWebServer(prHandler)
}
```

## 追加改善項目

### ドメインサービス層の追加 🟡中優先度
複雑なビジネスロジックが発生した場合の受け皿として：

```go
// domain/pull_request/service.go
type Service struct{}

func (s *Service) CalculateMetrics(prs []PullRequest) Metrics {
    // 複雑な計算ロジック
}
```

### 統一されたエラーハンドリング 🟡中優先度
```go
// domain/errors/errors.go
type DomainError struct {
    Code    string
    Message string
}

func (e DomainError) Error() string {
    return e.Message
}

var (
    ErrPullRequestNotFound = DomainError{Code: "PR001", Message: "Pull request not found"}
    ErrInvalidDateRange    = DomainError{Code: "PR002", Message: "Invalid date range"}
)
```

### 設定管理の分離 🟡中優先度
```go
// infrastructure/config/config.go
type Config struct {
    GitHubToken      string
    TargetRepos      []string
    Port             int
}

func Load() (*Config, error) {
    // 環境変数読み込み、バリデーション
}
```

## 実装順序

1. **Domain層の純粋化** - 外部依存を排除
2. **Repository Interface導入** - 抽象化層の追加
3. **UseCase分離** - HTTPハンドラーから分離
4. **依存関係逆転** - DI実装
5. **エラーハンドリング統一** - ドメインエラーの定義
6. **設定管理分離** - Infrastructure層への移動
7. **ドメインサービス追加** - 複雑なロジック対応

## 期待される効果

- **テスタビリティ向上**: モックによる単体テスト実装が容易
- **保守性向上**: 責任が明確に分離され、変更影響範囲が限定
- **拡張性向上**: 新しい要件への対応が柔軟
- **独立性確保**: 外部ライブラリの変更に強いアーキテクチャ

## 移行戦略

### 段階的移行アプローチ
1. 既存機能を維持しながら新しい構造を並行実装
2. 新旧両方のエンドポイントを一時的に提供
3. テスト完了後に旧実装を削除
4. 段階的にリファクタリングを進める

### リスク軽減策
- 各フェーズごとに十分なテストを実装
- 機能レベルでの動作確認を実施
- ロールバック可能な状態を維持