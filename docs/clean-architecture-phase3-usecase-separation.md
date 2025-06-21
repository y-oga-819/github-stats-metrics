# フェーズ3: UseCase層の責任分離

## 概要
現在のApplication層は実質的にHTTPハンドラーとして機能しており、Clean ArchitectureのUseCase層としての責任を果たしていません。本フェーズでは、Web層（Presentation）とUseCase層（Application）を明確に分離し、適切な責任分担を実現します。

## 現状の問題

### ファイル: `application/pull_request/get_pull_requests.go`
```go
func GetPullRequests(w http.ResponseWriter, r *http.Request) {
    req := &prDomain.GetPullRequestsRequest{}

    if err := decoder.Decode(req, r.URL.Query()); err != nil {
        http.Error(w, fmt.Sprintf("bad request: %s", err.Error()), http.StatusBadRequest)
        return
    }

    pullRequests := githubApiClient.Fetch(*req)

    presenter.Success(w, pullRequests)
}
```

### 問題点
1. **責任混在**: HTTPリクエスト処理とビジネスロジックが同一関数内
2. **テスト困難**: HTTPハンドラーのテストが複雑
3. **再利用不可**: Web以外からのビジネスロジック呼び出しが不可
4. **依存関係違反**: Application層がHTTPの詳細に依存

## 改善案

### 新しいアーキテクチャ構成

```
Presentation層 (HTTP Handler) 
    ↓ 
Application層 (UseCase)
    ↓
Domain層 (Repository Interface)
    ↑
Infrastructure層 (Repository Implementation)
```

### UseCase層の実装

#### `application/pull_request/usecase.go`
```go
package pull_request

import (
    "context"
    "fmt"
    "log"

    domain "github-stats-metrics/domain/pull_request"
)

// UseCase はPull Request関連のビジネスロジックを統括
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
    // 入力バリデーション
    if err := req.Validate(); err != nil {
        return nil, NewValidationError("invalid request parameters", err)
    }

    // ログ出力（ビジネス要件）
    log.Printf("Fetching pull requests for developers: %v, period: %s to %s", 
        req.Developers, req.StartDate, req.EndDate)

    // リポジトリからデータ取得
    pullRequests, err := uc.repo.GetPullRequests(ctx, req)
    if err != nil {
        return nil, NewRepositoryError("failed to fetch pull requests", err)
    }

    // ビジネスルール適用
    filtered := uc.applyBusinessRules(pullRequests)

    log.Printf("Retrieved %d pull requests (filtered from %d)", len(filtered), len(pullRequests))
    return filtered, nil
}

// applyBusinessRules はビジネスルールを適用してPRをフィルタリング
func (uc *UseCase) applyBusinessRules(pullRequests []domain.PullRequest) []domain.PullRequest {
    var filtered []domain.PullRequest
    
    for _, pr := range pullRequests {
        if !uc.shouldExcludePR(pr) {
            filtered = append(filtered, pr)
        }
    }
    
    return filtered
}

// shouldExcludePR はPRを除外すべきかを判定するビジネスルール
func (uc *UseCase) shouldExcludePR(pr domain.PullRequest) bool {
    // ルール1: epicブランチは除外
    if isEpicBranch(pr.HeadRefName) {
        return true
    }
    
    // ルール2: ドラフトPRは除外（将来の拡張）
    // if pr.IsDraft {
    //     return true
    // }
    
    // ルール3: 特定のラベルが付いているPRは除外（将来の拡張）
    // if hasExcludeLabel(pr.Labels) {
    //     return true
    // }
    
    return false
}

// isEpicBranch はepicブランチかを判定
func isEpicBranch(branchName string) bool {
    return len(branchName) > 5 && branchName[:5] == "epic/"
}

// GetPullRequestsWithMetrics はPRとメトリクスを同時に取得
func (uc *UseCase) GetPullRequestsWithMetrics(ctx context.Context, req domain.GetPullRequestsRequest) (*PullRequestsWithMetrics, error) {
    pullRequests, err := uc.GetPullRequests(ctx, req)
    if err != nil {
        return nil, err
    }
    
    metrics := uc.calculateMetrics(pullRequests)
    
    return &PullRequestsWithMetrics{
        PullRequests: pullRequests,
        Metrics:      metrics,
    }, nil
}

// calculateMetrics はPRリストからメトリクスを計算
func (uc *UseCase) calculateMetrics(pullRequests []domain.PullRequest) domain.Metrics {
    if len(pullRequests) == 0 {
        return domain.Metrics{}
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
    
    metrics := domain.Metrics{
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

// PullRequestsWithMetrics はPRとメトリクスを含む複合型
type PullRequestsWithMetrics struct {
    PullRequests []domain.PullRequest `json:"pullRequests"`
    Metrics      domain.Metrics       `json:"metrics"`
}
```

#### `application/pull_request/errors.go` （新規作成）
```go
package pull_request

import "fmt"

// UseCaseError はUseCase層で発生するエラーの基底型
type UseCaseError struct {
    Type    string
    Message string
    Cause   error
}

func (e UseCaseError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e UseCaseError) Unwrap() error {
    return e.Cause
}

// NewValidationError はバリデーションエラーを生成
func NewValidationError(message string, cause error) UseCaseError {
    return UseCaseError{
        Type:    "VALIDATION_ERROR",
        Message: message,
        Cause:   cause,
    }
}

// NewRepositoryError はリポジトリエラーを生成
func NewRepositoryError(message string, cause error) UseCaseError {
    return UseCaseError{
        Type:    "REPOSITORY_ERROR",
        Message: message,
        Cause:   cause,
    }
}

// NewBusinessRuleError はビジネスルールエラーを生成
func NewBusinessRuleError(message string, cause error) UseCaseError {
    return UseCaseError{
        Type:    "BUSINESS_RULE_ERROR",
        Message: message,
        Cause:   cause,
    }
}
```

### Presentation層（HTTPハンドラー）の実装

#### `presentation/pull_request/handler.go` （新規作成）
```go
package pull_request

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/gorilla/schema"
    
    usecase "github-stats-metrics/application/pull_request"
    domain "github-stats-metrics/domain/pull_request"
)

// Handler はPull Request APIのHTTPハンドラー
type Handler struct {
    useCase *usecase.UseCase
    decoder *schema.Decoder
}

// NewHandler はHandlerのコンストラクタ
func NewHandler(useCase *usecase.UseCase) *Handler {
    decoder := schema.NewDecoder()
    decoder.IgnoreUnknownKeys(true) // 不明なクエリパラメータを無視
    
    return &Handler{
        useCase: useCase,
        decoder: decoder,
    }
}

// GetPullRequests はPull RequestsのGETエンドポイント
func (h *Handler) GetPullRequests(w http.ResponseWriter, r *http.Request) {
    // リクエスト解析
    req, err := h.parseRequest(r)
    if err != nil {
        h.respondError(w, http.StatusBadRequest, "Invalid request parameters", err)
        return
    }
    
    // タイムアウト設定
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()
    
    // UseCase呼び出し
    pullRequests, err := h.useCase.GetPullRequests(ctx, *req)
    if err != nil {
        h.handleUseCaseError(w, err)
        return
    }
    
    // レスポンス返却
    h.respondSuccess(w, pullRequests)
}

// GetPullRequestsWithMetrics はPRとメトリクスを同時に返すエンドポイント
func (h *Handler) GetPullRequestsWithMetrics(w http.ResponseWriter, r *http.Request) {
    req, err := h.parseRequest(r)
    if err != nil {
        h.respondError(w, http.StatusBadRequest, "Invalid request parameters", err)
        return
    }
    
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()
    
    result, err := h.useCase.GetPullRequestsWithMetrics(ctx, *req)
    if err != nil {
        h.handleUseCaseError(w, err)
        return
    }
    
    h.respondSuccess(w, result)
}

// parseRequest はHTTPリクエストからDomainリクエストを生成
func (h *Handler) parseRequest(r *http.Request) (*domain.GetPullRequestsRequest, error) {
    req := &domain.GetPullRequestsRequest{}
    
    if err := h.decoder.Decode(req, r.URL.Query()); err != nil {
        return nil, fmt.Errorf("failed to decode query parameters: %w", err)
    }
    
    return req, nil
}

// handleUseCaseError はUseCase層のエラーを適切なHTTPレスポンスに変換
func (h *Handler) handleUseCaseError(w http.ResponseWriter, err error) {
    if useCaseErr, ok := err.(usecase.UseCaseError); ok {
        switch useCaseErr.Type {
        case "VALIDATION_ERROR":
            h.respondError(w, http.StatusBadRequest, "Validation failed", err)
        case "REPOSITORY_ERROR":
            h.respondError(w, http.StatusInternalServerError, "Data access failed", err)
        case "BUSINESS_RULE_ERROR":
            h.respondError(w, http.StatusUnprocessableEntity, "Business rule violation", err)
        default:
            h.respondError(w, http.StatusInternalServerError, "Internal server error", err)
        }
    } else {
        h.respondError(w, http.StatusInternalServerError, "Internal server error", err)
    }
}

// respondSuccess は成功レスポンスを返却
func (h *Handler) respondSuccess(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // 本番では適切に設定
    
    if err := json.NewEncoder(w).Encode(data); err != nil {
        h.respondError(w, http.StatusInternalServerError, "Failed to encode response", err)
    }
}

// respondError はエラーレスポンスを返却
func (h *Handler) respondError(w http.ResponseWriter, statusCode int, message string, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // 本番では適切に設定
    w.WriteHeader(statusCode)
    
    errorResponse := ErrorResponse{
        Error:   message,
        Details: err.Error(),
    }
    
    json.NewEncoder(w).Encode(errorResponse)
}

// ErrorResponse はエラーレスポンスの構造
type ErrorResponse struct {
    Error   string `json:"error"`
    Details string `json:"details,omitempty"`
}
```

### サーバー設定の更新

#### `server/webserver.go` （更新）
```go
package server

import (
    "fmt"
    "net/http"
    
    "github.com/gorilla/mux"
    
    pullRequestUseCase "github-stats-metrics/application/pull_request"
    pullRequestHandler "github-stats-metrics/presentation/pull_request"
    githubRepo "github-stats-metrics/infrastructure/github_api"
    todoUseCase "github-stats-metrics/application/todo"
)

// StartWebServer はWebサーバーを起動
func StartWebServer() error {
    fmt.Println("Start Web Server!")
    
    // 依存性注入の設定
    dependencies := setupDependencies()
    
    // ルーター設定
    r := mux.NewRouter().StrictSlash(true)
    setupRoutes(r, dependencies)
    
    fmt.Println("Server endpoints:")
    fmt.Println("  GET /api/pull_requests - Pull Requests取得")
    fmt.Println("  GET /api/pull_requests/metrics - Pull Requests + メトリクス取得")
    fmt.Println("  GET /api/todos - Todo取得")
    fmt.Println("  GET /health - ヘルスチェック")
    fmt.Println("  Server running on: http://localhost:8080")
    
    return http.ListenAndServe(":8080", r)
}

// Dependencies は依存関係を管理
type Dependencies struct {
    PullRequestHandler *pullRequestHandler.Handler
}

// setupDependencies は依存関係を初期化
func setupDependencies() *Dependencies {
    // Infrastructure層
    prRepo := githubRepo.NewRepository()
    
    // Application層
    prUseCase := pullRequestUseCase.NewUseCase(prRepo)
    
    // Presentation層
    prHandler := pullRequestHandler.NewHandler(prUseCase)
    
    return &Dependencies{
        PullRequestHandler: prHandler,
    }
}

// setupRoutes はルーティングを設定
func setupRoutes(r *mux.Router, deps *Dependencies) {
    // Pull Request関連
    r.HandleFunc("/api/pull_requests", deps.PullRequestHandler.GetPullRequests).Methods("GET")
    r.HandleFunc("/api/pull_requests/metrics", deps.PullRequestHandler.GetPullRequestsWithMetrics).Methods("GET")
    
    // 既存のTodo（後で同様にリファクタリング）
    r.HandleFunc("/api/todos", todoUseCase.GetTodos).Methods("GET")
    
    // ヘルスチェック
    r.HandleFunc("/health", healthCheck).Methods("GET")
}

// healthCheck はヘルスチェックエンドポイント
func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok","service":"github-stats-metrics"}`))
}
```

### テスト実装

#### `presentation/pull_request/handler_test.go` （新規作成）
```go
package pull_request

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    usecase "github-stats-metrics/application/pull_request"
    domain "github-stats-metrics/domain/pull_request"
)

// MockUseCase はテスト用のUseCase実装
type MockUseCase struct {
    pullRequests []domain.PullRequest
    err          error
}

func (m *MockUseCase) GetPullRequests(ctx context.Context, req domain.GetPullRequestsRequest) ([]domain.PullRequest, error) {
    return m.pullRequests, m.err
}

func (m *MockUseCase) GetPullRequestsWithMetrics(ctx context.Context, req domain.GetPullRequestsRequest) (*usecase.PullRequestsWithMetrics, error) {
    if m.err != nil {
        return nil, m.err
    }
    return &usecase.PullRequestsWithMetrics{
        PullRequests: m.pullRequests,
        Metrics:      domain.Metrics{TotalCount: len(m.pullRequests)},
    }, nil
}

func TestHandler_GetPullRequests(t *testing.T) {
    // テストデータ
    now := time.Now()
    mockPRs := []domain.PullRequest{
        {
            ID:        "1",
            Title:     "Test PR",
            CreatedAt: now,
        },
    }
    
    // Mock UseCase
    mockUseCase := &MockUseCase{pullRequests: mockPRs}
    handler := NewHandler(mockUseCase)
    
    // HTTPリクエスト作成
    req := httptest.NewRequest("GET", "/api/pull_requests?startdate=2023-01-01&enddate=2023-01-31&developers=user1", nil)
    w := httptest.NewRecorder()
    
    // ハンドラー実行
    handler.GetPullRequests(w, req)
    
    // レスポンス検証
    if w.Code != http.StatusOK {
        t.Fatalf("expected status 200, got %d", w.Code)
    }
    
    var response []domain.PullRequest
    if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }
    
    if len(response) != 1 {
        t.Fatalf("expected 1 PR, got %d", len(response))
    }
    
    if response[0].ID != "1" {
        t.Fatalf("expected PR ID 1, got %s", response[0].ID)
    }
}
```

## 実装チェックリスト

- [ ] UseCase層の実装（ビジネスロジック分離）
- [ ] エラーハンドリングの統一
- [ ] Handler層の実装（HTTP詳細分離）
- [ ] 依存性注入の設定
- [ ] ルーティングの更新
- [ ] テストの実装
- [ ] ヘルスチェックエンドポイント追加
- [ ] 既存エンドポイントとの互換性確認

## 期待される効果

1. **責任分離**: Web層とビジネスロジック層の明確な分離
2. **テスタビリティ**: 各層を独立してテスト可能
3. **再利用性**: UseCaseをCLI、バッチ処理等でも利用可能
4. **保守性**: 変更影響範囲の限定化
5. **エラーハンドリング**: 統一されたエラー処理とレスポンス

## 移行戦略

1. **UseCase実装**: 既存ロジックをUseCase層に移行
2. **Handler実装**: HTTPハンドラーを新規作成
3. **段階的置換**: 既存エンドポイントと並行稼働
4. **テスト実装**: 各層のテスト作成
5. **旧実装削除**: 動作確認後に旧ハンドラー削除

## リスク対策

1. **パフォーマンス**: レイヤー分離によるオーバーヘッド測定
2. **複雑性**: 適切なドキュメント作成
3. **移行期間**: 段階的移行による影響範囲の限定
4. **API互換性**: 既存クライアント（フロントエンド）との互換性維持