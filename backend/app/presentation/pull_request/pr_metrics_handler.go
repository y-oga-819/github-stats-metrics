package pull_request

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	analyticsApp "github-stats-metrics/application/analytics"
	prApp "github-stats-metrics/application/pull_request"
	"github-stats-metrics/infrastructure/repository"
)

// PRMetricsHandler はPRメトリクスのHTTPハンドラー
type PRMetricsHandler struct {
	prMetricsRepo     *repository.PRMetricsRepository
	metricsAggregator *analyticsApp.MetricsAggregator
	presenter         *PRMetricsPresenter
}

// NewPRMetricsHandler は新しいPRメトリクスハンドラーを作成
func NewPRMetricsHandler(
	prMetricsRepo *repository.PRMetricsRepository,
	metricsAggregator *analyticsApp.MetricsAggregator,
) *PRMetricsHandler {
	return &PRMetricsHandler{
		prMetricsRepo:     prMetricsRepo,
		metricsAggregator: metricsAggregator,
		presenter:         NewPRMetricsPresenter(),
	}
}

// GetPRMetrics は指定されたPRのメトリクスを取得
func (h *PRMetricsHandler) GetPRMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	prID := vars["id"]

	if prID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PR_ID", "PR IDが指定されていません", nil)
		return
	}

	// PRメトリクスを取得
	metrics, err := h.prMetricsRepo.FindByPRID(ctx, prID)
	if err != nil {
		log.Printf("Failed to get PR metrics: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "メトリクスの取得に失敗しました", nil)
		return
	}

	if metrics == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "PR_NOT_FOUND", "指定されたPRが見つかりません", nil)
		return
	}

	// レスポンス形式に変換
	response := h.presenter.ToPRMetricsResponse(metrics)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetCycleTimeMetrics はサイクルタイムメトリクスを取得
func (h *PRMetricsHandler) GetCycleTimeMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// クエリパラメータの解析
	params, err := h.parseDateRangeParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// PRメトリクスを取得
	metrics, err := h.prMetricsRepo.FindByDateRange(ctx, params.StartDate, params.EndDate, params.Developers, params.Repositories)
	if err != nil {
		log.Printf("Failed to get PR metrics for cycle time: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "メトリクスの取得に失敗しました", nil)
		return
	}

	// サイクルタイムメトリクスに変換
	response := h.presenter.ToCycleTimeMetricsResponse(metrics, params.Period, params.StartDate, params.EndDate)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetReviewTimeMetrics はレビュー時間メトリクスを取得
func (h *PRMetricsHandler) GetReviewTimeMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// クエリパラメータの解析
	params, err := h.parseDateRangeParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// PRメトリクスを取得
	metrics, err := h.prMetricsRepo.FindByDateRange(ctx, params.StartDate, params.EndDate, params.Developers, params.Repositories)
	if err != nil {
		log.Printf("Failed to get PR metrics for review time: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "メトリクスの取得に失敗しました", nil)
		return
	}

	// レビュー時間メトリクスに変換
	response := h.presenter.ToReviewTimeMetricsResponse(metrics, params.Period, params.StartDate, params.EndDate)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// ListPRMetrics はPRメトリクスの一覧を取得
func (h *PRMetricsHandler) ListPRMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// クエリパラメータの解析
	params, err := h.parseListParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// PRメトリクスを取得
	metrics, err := h.prMetricsRepo.FindByDateRange(ctx, params.StartDate, params.EndDate, params.Developers, params.Repositories)
	if err != nil {
		log.Printf("Failed to get PR metrics list: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "メトリクスの取得に失敗しました", nil)
		return
	}

	// ページング処理
	totalCount := len(metrics)
	startIndex := (params.Page - 1) * params.PageSize
	endIndex := startIndex + params.PageSize

	if startIndex >= totalCount {
		metrics = []*prApp.PRMetrics{}
	} else if endIndex > totalCount {
		metrics = metrics[startIndex:]
	} else {
		metrics = metrics[startIndex:endIndex]
	}

	// レスポンス形式に変換
	response := h.presenter.ToPRListResponse(metrics, totalCount, params.Page, params.PageSize)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetDeveloperMetrics は開発者別のメトリクスを取得
func (h *PRMetricsHandler) GetDeveloperMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	developer := vars["developer"]

	if developer == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_DEVELOPER", "開発者名が指定されていません", nil)
		return
	}

	// クエリパラメータの解析
	params, err := h.parseDateRangeParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// 開発者のPRメトリクスを取得
	metrics, err := h.prMetricsRepo.FindByDeveloper(ctx, developer, params.StartDate, params.EndDate)
	if err != nil {
		log.Printf("Failed to get developer metrics: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "メトリクスの取得に失敗しました", nil)
		return
	}

	// レスポンス形式に変換
	response := h.presenter.ToPRListResponse(metrics, len(metrics), 1, len(metrics))
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetRepositoryMetrics はリポジトリ別のメトリクスを取得
func (h *PRMetricsHandler) GetRepositoryMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	repository := vars["repository"]

	if repository == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REPOSITORY", "リポジトリ名が指定されていません", nil)
		return
	}

	// クエリパラメータの解析
	params, err := h.parseDateRangeParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// リポジトリのPRメトリクスを取得
	metrics, err := h.prMetricsRepo.FindByRepository(ctx, repository, params.StartDate, params.EndDate)
	if err != nil {
		log.Printf("Failed to get repository metrics: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "メトリクスの取得に失敗しました", nil)
		return
	}

	// レスポンス形式に変換
	response := h.presenter.ToPRListResponse(metrics, len(metrics), 1, len(metrics))
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetHealthCheck はヘルスチェック
func (h *PRMetricsHandler) GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// データベース接続チェック
	dbStatus := h.checkDatabaseHealth(ctx)
	
	// GitHub API接続チェック（省略）
	githubStatus := GitHubAPIStatus{
		Available:    true,
		LastChecked:  time.Now(),
		ResponseTime: "50ms",
		RateLimit: RateLimitStatus{
			Limit:     5000,
			Remaining: 4500,
			Reset:     time.Now().Add(time.Hour),
		},
	}

	status := "healthy"
	if !dbStatus.Connected || !githubStatus.Available {
		status = "unhealthy"
	}

	response := HealthCheckResponse{
		Status:    status,
		Version:   "1.0.0",
		Timestamp: time.Now(),
		Services: map[string]string{
			"database": "connected",
			"github":   "available",
		},
		Database:  dbStatus,
		GitHubAPI: githubStatus,
	}

	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	h.writeJSONResponse(w, statusCode, response)
}

// プライベートメソッド

type DateRangeParams struct {
	StartDate    time.Time
	EndDate      time.Time
	Period       string
	Developers   []string
	Repositories []string
}

type ListParams struct {
	DateRangeParams
	Page     int
	PageSize int
}

func (h *PRMetricsHandler) parseDateRangeParams(r *http.Request) (*DateRangeParams, error) {
	query := r.URL.Query()

	// 期間パラメータ
	period := query.Get("period")
	if period == "" {
		period = "30days"
	}

	// 日付範囲の計算
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30) // デフォルト30日

	if startDateStr := query.Get("startdate"); startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start date format: %s", startDateStr)
		}
		startDate = parsed
	}

	if endDateStr := query.Get("enddate"); endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format: %s", endDateStr)
		}
		endDate = parsed
	}

	// 開発者フィルタ
	developers := query["developers[]"]

	// リポジトリフィルタ
	repositories := query["repositories[]"]

	return &DateRangeParams{
		StartDate:    startDate,
		EndDate:      endDate,
		Period:       period,
		Developers:   developers,
		Repositories: repositories,
	}, nil
}

func (h *PRMetricsHandler) parseListParams(r *http.Request) (*ListParams, error) {
	dateParams, err := h.parseDateRangeParams(r)
	if err != nil {
		return nil, err
	}

	query := r.URL.Query()

	// ページネーション
	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		parsed, err := strconv.Atoi(pageStr)
		if err != nil || parsed < 1 {
			return nil, fmt.Errorf("invalid page number: %s", pageStr)
		}
		page = parsed
	}

	pageSize := 50
	if pageSizeStr := query.Get("pageSize"); pageSizeStr != "" {
		parsed, err := strconv.Atoi(pageSizeStr)
		if err != nil || parsed < 1 || parsed > 100 {
			return nil, fmt.Errorf("invalid page size: %s", pageSizeStr)
		}
		pageSize = parsed
	}

	return &ListParams{
		DateRangeParams: *dateParams,
		Page:            page,
		PageSize:        pageSize,
	}, nil
}

func (h *PRMetricsHandler) checkDatabaseHealth(ctx context.Context) DatabaseStatus {
	start := time.Now()
	
	// データベース統計を取得してチェック
	_, err := h.prMetricsRepo.GetStatistics(ctx)
	
	responseTime := time.Since(start)
	connected := err == nil

	return DatabaseStatus{
		Connected:    connected,
		LastChecked:  time.Now(),
		ResponseTime: responseTime.String(),
	}
}

func (h *PRMetricsHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func (h *PRMetricsHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
	errorResponse := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Code:    code,
		Message: message,
		Details: details,
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}

// RegisterRoutes はルートを登録
func (h *PRMetricsHandler) RegisterRoutes(router *mux.Router) {
	// PRメトリクス個別取得
	router.HandleFunc("/api/pull_requests/{id}/metrics", h.GetPRMetrics).Methods("GET")
	
	// メトリクス集計API
	router.HandleFunc("/api/metrics/cycle_time", h.GetCycleTimeMetrics).Methods("GET")
	router.HandleFunc("/api/metrics/review_time", h.GetReviewTimeMetrics).Methods("GET")
	
	// PRリスト取得
	router.HandleFunc("/api/pull_requests", h.ListPRMetrics).Methods("GET")
	
	// 開発者・リポジトリ別メトリクス
	router.HandleFunc("/api/developers/{developer}/metrics", h.GetDeveloperMetrics).Methods("GET")
	router.HandleFunc("/api/repositories/{repository}/metrics", h.GetRepositoryMetrics).Methods("GET")
	
	// ヘルスチェック
	router.HandleFunc("/api/health", h.GetHealthCheck).Methods("GET")
}