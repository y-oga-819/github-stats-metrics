# フェーズ4: 依存関係の逆転実装

## 概要
Clean Architectureの核心である「依存関係逆転の原則」を完全に実装します。現在の外向きの依存関係を内向きに変更し、上位層が下位層の抽象に依存する構造を確立します。

## 現状の依存関係

### 問題のある依存関係の方向
```
Presentation → Application → Infrastructure
     ↓              ↓              ↓
  (HTTP詳細)    (ビジネス)      (GitHub API)
```

**具体的な問題**:
- `application/pull_request/get_pull_requests.go:6`
  ```go
  githubApiClient "github-stats-metrics/infrastructure/github_api"
  ```
- Application層がInfrastructure層の具象実装に直接依存

### 理想的な依存関係の方向
```
Presentation → Application ← Infrastructure
     ↓              ↓              ↑
  (HTTP詳細)    (ビジネス)      (GitHub API)
                    ↓
                 (抽象)
```

## 依存関係逆転の実装

### 1. 抽象の定義（Domain層）

#### `domain/pull_request/repository.go`（完全版）
```go
package pull_request

import (
    "context"
    "time"
)

// Repository はPull Requestデータアクセスの抽象化
type Repository interface {
    // GetPullRequests は条件に基づいてPull Requestsを取得
    GetPullRequests(ctx context.Context, req GetPullRequestsRequest) ([]PullRequest, error)
    
    // GetPullRequestByID は特定IDのPull Requestを取得
    GetPullRequestByID(ctx context.Context, id string) (*PullRequest, error)
    
    // GetRepositories は対象リポジトリ一覧を取得
    GetRepositories(ctx context.Context) ([]string, error)
    
    // GetDevelopers は開発者一覧を取得
    GetDevelopers(ctx context.Context, repositories []string) ([]string, error)
}

// ConfigRepository は設定データアクセスの抽象化
type ConfigRepository interface {
    // GetGitHubToken はGitHubトークンを取得
    GetGitHubToken() (string, error)
    
    // GetTargetRepositories は対象リポジトリを取得
    GetTargetRepositories() ([]string, error)
    
    // GetServerPort はサーバーポートを取得
    GetServerPort() (int, error)
}

// Logger はログ出力の抽象化
type Logger interface {
    Info(message string, fields ...interface{})
    Error(message string, err error, fields ...interface{})
    Debug(message string, fields ...interface{})
    Warn(message string, fields ...interface{})
}

// MetricsRepository はメトリクス保存の抽象化（将来の拡張用）
type MetricsRepository interface {
    // SaveMetrics はメトリクスを永続化
    SaveMetrics(ctx context.Context, metrics Metrics) error
    
    // GetHistoricalMetrics は過去のメトリクスを取得
    GetHistoricalMetrics(ctx context.Context, period Period) ([]Metrics, error)
}

// Period はメトリクス期間を表現
type Period struct {
    StartDate time.Time
    EndDate   time.Time
}
```

### 2. Application層の完全分離

#### `application/pull_request/usecase.go`（依存関係逆転版）
```go
package pull_request

import (
    "context"
    "fmt"
    
    domain "github-stats-metrics/domain/pull_request"
)

// UseCase はPull Request関連のビジネスロジックを統括
type UseCase struct {
    prRepo     domain.Repository
    configRepo domain.ConfigRepository
    logger     domain.Logger
}

// NewUseCase はUseCaseのコンストラクタ（依存性注入）
func NewUseCase(
    prRepo domain.Repository,
    configRepo domain.ConfigRepository,
    logger domain.Logger,
) *UseCase {
    return &UseCase{
        prRepo:     prRepo,
        configRepo: configRepo,
        logger:     logger,
    }
}

// GetPullRequests はPull Requestsを取得し、ビジネスルールを適用
func (uc *UseCase) GetPullRequests(ctx context.Context, req domain.GetPullRequestsRequest) ([]domain.PullRequest, error) {
    // バリデーション
    if err := req.Validate(); err != nil {
        uc.logger.Error("validation failed", err, "request", req)
        return nil, NewValidationError("invalid request parameters", err)
    }
    
    uc.logger.Info("fetching pull requests", 
        "developers", req.Developers,
        "startDate", req.StartDate,
        "endDate", req.EndDate)
    
    // リポジトリから取得（抽象に依存）
    pullRequests, err := uc.prRepo.GetPullRequests(ctx, req)
    if err != nil {
        uc.logger.Error("repository access failed", err, "request", req)
        return nil, NewRepositoryError("failed to fetch pull requests", err)
    }
    
    // ビジネスルール適用
    filtered := uc.applyBusinessRules(pullRequests)
    
    uc.logger.Info("pull requests retrieved successfully",
        "totalCount", len(pullRequests),
        "filteredCount", len(filtered))
    
    return filtered, nil
}

// ValidateConfiguration は設定の妥当性を検証
func (uc *UseCase) ValidateConfiguration(ctx context.Context) error {
    // GitHubトークンの存在確認
    token, err := uc.configRepo.GetGitHubToken()
    if err != nil || token == "" {
        return NewConfigurationError("GitHub token is not configured", err)
    }
    
    // 対象リポジトリの存在確認
    repos, err := uc.configRepo.GetTargetRepositories()
    if err != nil || len(repos) == 0 {
        return NewConfigurationError("target repositories are not configured", err)
    }
    
    uc.logger.Info("configuration validated successfully",
        "repositoryCount", len(repos))
    
    return nil
}

// GetAvailableDevelopers は利用可能な開発者一覧を取得
func (uc *UseCase) GetAvailableDevelopers(ctx context.Context) ([]string, error) {
    repos, err := uc.configRepo.GetTargetRepositories()
    if err != nil {
        return nil, NewConfigurationError("failed to get target repositories", err)
    }
    
    developers, err := uc.prRepo.GetDevelopers(ctx, repos)
    if err != nil {
        return nil, NewRepositoryError("failed to get developers", err)
    }
    
    return developers, nil
}

// applyBusinessRules はビジネスルールを適用
func (uc *UseCase) applyBusinessRules(pullRequests []domain.PullRequest) []domain.PullRequest {
    var filtered []domain.PullRequest
    
    for _, pr := range pullRequests {
        if !uc.shouldExcludePR(pr) {
            filtered = append(filtered, pr)
        }
    }
    
    return filtered
}

// shouldExcludePR はPR除外判定のビジネスルール
func (uc *UseCase) shouldExcludePR(pr domain.PullRequest) bool {
    // ビジネスルール1: epicブランチは除外
    if isEpicBranch(pr.HeadRefName) {
        uc.logger.Debug("excluding epic branch PR", "prId", pr.ID, "branch", pr.HeadRefName)
        return true
    }
    
    // ビジネスルール2: マージされていないPRは除外
    if !pr.IsMerged() {
        uc.logger.Debug("excluding unmerged PR", "prId", pr.ID)
        return true
    }
    
    return false
}

func isEpicBranch(branchName string) bool {
    return len(branchName) > 5 && branchName[:5] == "epic/"
}
```

#### `application/pull_request/errors.go`（拡張版）
```go
package pull_request

import "fmt"

// UseCaseError はUseCase層のエラー基底型
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

// エラータイプ定数
const (
    ErrorTypeValidation    = "VALIDATION_ERROR"
    ErrorTypeRepository    = "REPOSITORY_ERROR"
    ErrorTypeConfiguration = "CONFIGURATION_ERROR"
    ErrorTypeBusinessRule  = "BUSINESS_RULE_ERROR"
)

// NewValidationError はバリデーションエラーを生成
func NewValidationError(message string, cause error) UseCaseError {
    return UseCaseError{
        Type:    ErrorTypeValidation,
        Message: message,
        Cause:   cause,
    }
}

// NewRepositoryError はリポジトリエラーを生成
func NewRepositoryError(message string, cause error) UseCaseError {
    return UseCaseError{
        Type:    ErrorTypeRepository,
        Message: message,
        Cause:   cause,
    }
}

// NewConfigurationError は設定エラーを生成
func NewConfigurationError(message string, cause error) UseCaseError {
    return UseCaseError{
        Type:    ErrorTypeConfiguration,
        Message: message,
        Cause:   cause,
    }
}

// NewBusinessRuleError はビジネスルールエラーを生成
func NewBusinessRuleError(message string, cause error) UseCaseError {
    return UseCaseError{
        Type:    ErrorTypeBusinessRule,
        Message: message,
        Cause:   cause,
    }
}
```

### 3. Infrastructure層の抽象実装

#### `infrastructure/github_api/repository.go`（インターフェース実装）
```go
package github_api

import (
    "context"
    "fmt"
    "strings"
    
    domain "github-stats-metrics/domain/pull_request"
    "github.com/shurcooL/githubv4"
    "golang.org/x/oauth2"
)

// repository はdomain.Repositoryインターフェースの実装
type repository struct {
    client     *githubv4.Client
    configRepo domain.ConfigRepository
    logger     domain.Logger
}

// NewRepository はGitHub APIを使用するRepository実装を作成
func NewRepository(configRepo domain.ConfigRepository, logger domain.Logger) domain.Repository {
    return &repository{
        configRepo: configRepo,
        logger:     logger,
    }
}

// 遅延初期化でクライアントを作成
func (r *repository) getClient() (*githubv4.Client, error) {
    if r.client == nil {
        token, err := r.configRepo.GetGitHubToken()
        if err != nil {
            return nil, fmt.Errorf("failed to get GitHub token: %w", err)
        }
        
        if token == "" {
            return nil, fmt.Errorf("GitHub token is empty")
        }
        
        src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
        httpClient := oauth2.NewClient(context.Background(), src)
        r.client = githubv4.NewClient(httpClient)
        
        r.logger.Info("GitHub API client initialized")
    }
    
    return r.client, nil
}

// GetPullRequests はGitHub APIからPull Requestsを取得
func (r *repository) GetPullRequests(ctx context.Context, req domain.GetPullRequestsRequest) ([]domain.PullRequest, error) {
    client, err := r.getClient()
    if err != nil {
        return nil, fmt.Errorf("failed to initialize GitHub client: %w", err)
    }
    
    query := graphqlQuery{}
    variables := map[string]interface{}{
        "searchType": githubv4.SearchTypeIssue,
        "cursor":     (*githubv4.String)(nil),
        "query":      githubv4.String(r.createQuery(req)),
    }
    
    var allPRs []domain.PullRequest
    
    for {
        if err := client.Query(ctx, &query, variables); err != nil {
            r.logger.Error("GitHub API query failed", err, "variables", variables)
            return nil, fmt.Errorf("github api query failed: %w", err)
        }
        
        // レート制限チェック
        if query.RateLimit.Remaining < 10 {
            r.logger.Warn("GitHub API rate limit approaching",
                "remaining", query.RateLimit.Remaining,
                "resetAt", query.RateLimit.ResetAt)
        }
        
        // レスポンス変換
        for _, node := range query.Search.Nodes {
            domainPR := r.convertToDomain(node.Pr)
            allPRs = append(allPRs, domainPR)
        }
        
        if !query.Search.PageInfo.HasNextPage {
            break
        }
        
        variables["cursor"] = githubv4.NewString(query.Search.PageInfo.EndCursor)
    }
    
    r.logger.Info("pull requests fetched from GitHub API",
        "count", len(allPRs),
        "developers", req.Developers)
    
    return allPRs, nil
}

// GetPullRequestByID は特定IDのPull Requestを取得
func (r *repository) GetPullRequestByID(ctx context.Context, id string) (*domain.PullRequest, error) {
    // 実装は必要に応じて追加
    return nil, fmt.Errorf("not implemented yet")
}

// GetRepositories は対象リポジトリ一覧を取得
func (r *repository) GetRepositories(ctx context.Context) ([]string, error) {
    return r.configRepo.GetTargetRepositories()
}

// GetDevelopers は開発者一覧を取得
func (r *repository) GetDevelopers(ctx context.Context, repositories []string) ([]string, error) {
    // 実際の実装では、GitHub APIから開発者一覧を取得
    // 現在は設定ベースで実装
    return []string{"developer1", "developer2", "developer3"}, nil
}

// createQuery はGitHub GraphQL検索クエリを生成
func (r *repository) createQuery(req domain.GetPullRequestsRequest) string {
    query := fmt.Sprintf("merged:%s..%s ", req.StartDate, req.EndDate)
    
    // リポジトリ設定
    repositories, err := r.configRepo.GetTargetRepositories()
    if err != nil {
        r.logger.Error("failed to get target repositories", err)
        return query
    }
    
    query += "repo:" + strings.Join(repositories, " repo:") + " "
    query += "author:" + strings.Join(req.Developers, " author:")
    
    r.logger.Debug("GitHub search query created", "query", query)
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
    
    // nil値の適切な処理
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
```

#### `infrastructure/config/repository.go`（新規作成）
```go
package config

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    
    domain "github-stats-metrics/domain/pull_request"
)

// repository はdomain.ConfigRepositoryインターフェースの実装
type repository struct{}

// NewConfigRepository は設定リポジトリを作成
func NewConfigRepository() domain.ConfigRepository {
    return &repository{}
}

// GetGitHubToken はGitHubトークンを取得
func (r *repository) GetGitHubToken() (string, error) {
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        return "", fmt.Errorf("GITHUB_TOKEN environment variable is not set")
    }
    return token, nil
}

// GetTargetRepositories は対象リポジトリを取得
func (r *repository) GetTargetRepositories() ([]string, error) {
    reposStr := os.Getenv("GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES")
    if reposStr == "" {
        return nil, fmt.Errorf("GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES environment variable is not set")
    }
    
    repos := strings.Split(reposStr, ",")
    for i, repo := range repos {
        repos[i] = strings.TrimSpace(repo)
    }
    
    return repos, nil
}

// GetServerPort はサーバーポートを取得
func (r *repository) GetServerPort() (int, error) {
    portStr := os.Getenv("SERVER_PORT")
    if portStr == "" {
        return 8080, nil // デフォルト値
    }
    
    port, err := strconv.Atoi(portStr)
    if err != nil {
        return 0, fmt.Errorf("invalid SERVER_PORT value: %w", err)
    }
    
    return port, nil
}
```

#### `infrastructure/logger/logger.go`（新規作成）
```go
package logger

import (
    "log"
    "os"
    
    domain "github-stats-metrics/domain/pull_request"
)

// logger はdomain.Loggerインターフェースの実装
type logger struct {
    infoLogger  *log.Logger
    errorLogger *log.Logger
    debugLogger *log.Logger
    warnLogger  *log.Logger
    debugMode   bool
}

// NewLogger はロガーを作成
func NewLogger() domain.Logger {
    debugMode := os.Getenv("DEBUG") == "true"
    
    return &logger{
        infoLogger:  log.New(os.Stdout, "INFO: ", log.LstdFlags|log.Lshortfile),
        errorLogger: log.New(os.Stderr, "ERROR: ", log.LstdFlags|log.Lshortfile),
        debugLogger: log.New(os.Stdout, "DEBUG: ", log.LstdFlags|log.Lshortfile),
        warnLogger:  log.New(os.Stdout, "WARN: ", log.LstdFlags|log.Lshortfile),
        debugMode:   debugMode,
    }
}

// Info は情報ログを出力
func (l *logger) Info(message string, fields ...interface{}) {
    l.infoLogger.Printf(message+" %v", fields...)
}

// Error はエラーログを出力
func (l *logger) Error(message string, err error, fields ...interface{}) {
    l.errorLogger.Printf(message+" error=%v %v", err, fields)
}

// Debug はデバッグログを出力
func (l *logger) Debug(message string, fields ...interface{}) {
    if l.debugMode {
        l.debugLogger.Printf(message+" %v", fields...)
    }
}

// Warn は警告ログを出力
func (l *logger) Warn(message string, fields ...interface{}) {
    l.warnLogger.Printf(message+" %v", fields...)
}
```

### 4. 依存性注入の完全実装

#### `cmd/main.go`（完全版）
```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/joho/godotenv"
    
    // Application層
    pullRequestUseCase "github-stats-metrics/application/pull_request"
    
    // Infrastructure層
    configRepo "github-stats-metrics/infrastructure/config"
    githubRepo "github-stats-metrics/infrastructure/github_api"
    loggerImpl "github-stats-metrics/infrastructure/logger"
    
    // Presentation層
    "github-stats-metrics/server"
)

func main() {
    // 環境変数読み込み
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: Error loading .env file: %v", err)
    }
    
    // 依存関係の構築
    dependencies := buildDependencies()
    
    // 設定検証
    if err := validateConfiguration(dependencies); err != nil {
        log.Fatalf("Configuration validation failed: %v", err)
    }
    
    // サーバー起動
    server := server.NewServer(dependencies)
    
    // Graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // シグナルハンドリング
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        log.Println("Shutdown signal received")
        cancel()
    }()
    
    // サーバー実行
    if err := server.Run(ctx); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}

// buildDependencies は依存関係を構築
func buildDependencies() *server.Dependencies {
    // Infrastructure層の構築
    logger := loggerImpl.NewLogger()
    configRepository := configRepo.NewConfigRepository()
    prRepository := githubRepo.NewRepository(configRepository, logger)
    
    // Application層の構築
    prUseCase := pullRequestUseCase.NewUseCase(prRepository, configRepository, logger)
    
    return &server.Dependencies{
        PullRequestUseCase: prUseCase,
        ConfigRepository:   configRepository,
        Logger:             logger,
    }
}

// validateConfiguration は起動時設定検証
func validateConfiguration(deps *server.Dependencies) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return deps.PullRequestUseCase.ValidateConfiguration(ctx)
}
```

#### `server/server.go`（更新版）
```go
package server

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "github.com/gorilla/mux"
    
    pullRequestUseCase "github-stats-metrics/application/pull_request"
    domain "github-stats-metrics/domain/pull_request"
    pullRequestHandler "github-stats-metrics/presentation/pull_request"
)

// Dependencies は依存関係を管理
type Dependencies struct {
    PullRequestUseCase *pullRequestUseCase.UseCase
    ConfigRepository   domain.ConfigRepository
    Logger             domain.Logger
}

// Server はWebサーバーを表現
type Server struct {
    deps   *Dependencies
    server *http.Server
}

// NewServer はサーバーを作成
func NewServer(deps *Dependencies) *Server {
    return &Server{
        deps: deps,
    }
}

// Run はサーバーを実行
func (s *Server) Run(ctx context.Context) error {
    // ポート取得
    port, err := s.deps.ConfigRepository.GetServerPort()
    if err != nil {
        return fmt.Errorf("failed to get server port: %w", err)
    }
    
    // ルーター設定
    router := s.setupRoutes()
    
    // HTTPサーバー設定
    s.server = &http.Server{
        Addr:         fmt.Sprintf(":%d", port),
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    s.deps.Logger.Info("server starting", "port", port)
    
    // Graceful shutdown対応
    go func() {
        <-ctx.Done()
        s.deps.Logger.Info("shutting down server")
        
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        
        if err := s.server.Shutdown(shutdownCtx); err != nil {
            s.deps.Logger.Error("server shutdown failed", err)
        }
    }()
    
    // サーバー起動
    if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return fmt.Errorf("server failed to start: %w", err)
    }
    
    s.deps.Logger.Info("server stopped")
    return nil
}

// setupRoutes はルーティングを設定
func (s *Server) setupRoutes() *mux.Router {
    // Handler作成
    prHandler := pullRequestHandler.NewHandler(s.deps.PullRequestUseCase)
    
    router := mux.NewRouter().StrictSlash(true)
    
    // API routes
    api := router.PathPrefix("/api").Subrouter()
    api.HandleFunc("/pull_requests", prHandler.GetPullRequests).Methods("GET")
    api.HandleFunc("/pull_requests/metrics", prHandler.GetPullRequestsWithMetrics).Methods("GET")
    
    // Health check
    router.HandleFunc("/health", s.healthCheck).Methods("GET")
    
    // CORS middleware（本番では適切に設定）
    router.Use(s.corsMiddleware)
    
    return router
}

// healthCheck はヘルスチェックエンドポイント
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok","service":"github-stats-metrics"}`))
}

// corsMiddleware はCORSミドルウェア
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## 実装チェックリスト

- [ ] Domain層に抽象インターフェース定義
- [ ] Application層の抽象依存への変更
- [ ] Infrastructure層でのインターフェース実装
- [ ] 依存性注入コンテナの実装
- [ ] 設定管理の分離
- [ ] ログ出力の抽象化
- [ ] Graceful shutdown対応
- [ ] テストの実装（各層独立）
- [ ] 設定検証の実装

## 期待される効果

1. **完全な疎結合**: 各層が抽象にのみ依存
2. **テスタビリティ最大化**: 全ての依存関係をモック可能
3. **実装の交換可能性**: インターフェース準拠であれば任意の実装に交換可能
4. **開発効率向上**: 各層を独立して開発・テスト可能
5. **保守性向上**: 変更影響範囲の完全な限定化

## 移行戦略

1. **インターフェース定義**: 抽象レイヤーの先行実装
2. **Infrastructure実装**: 既存ロジックをインターフェース実装に移行
3. **Application層修正**: 抽象依存への変更
4. **DI実装**: 依存性注入機構の構築
5. **段階的テスト**: 各層の独立テスト実装
6. **完全移行**: 旧実装の削除

## リスク対策

1. **複雑性管理**: 適切な抽象レベルの設定
2. **パフォーマンス**: インターフェース呼び出しによるオーバーヘッド測定
3. **過度な抽象化**: YAGNIの原則に基づく適切な境界設定
4. **学習コスト**: チーム向けドキュメント作成と研修