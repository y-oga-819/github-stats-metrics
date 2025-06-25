package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	analyticsApp "github-stats-metrics/application/analytics"
	"github-stats-metrics/infrastructure/repository"
)

// AnalyticsHandler は集計データのHTTPハンドラー
type AnalyticsHandler struct {
	aggregatedRepo    *repository.AggregatedMetricsRepository
	metricsAggregator *analyticsApp.MetricsAggregator
	presenter         *AnalyticsPresenter
}

// NewAnalyticsHandler は新しい集計データハンドラーを作成
func NewAnalyticsHandler(
	aggregatedRepo *repository.AggregatedMetricsRepository,
	metricsAggregator *analyticsApp.MetricsAggregator,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		aggregatedRepo:    aggregatedRepo,
		metricsAggregator: metricsAggregator,
		presenter:         NewAnalyticsPresenter(),
	}
}

// GetTeamMetrics はチームメトリクスを取得
func (h *AnalyticsHandler) GetTeamMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// パラメータ解析
	params, err := h.parseAnalyticsParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// チームメトリクスを取得
	metricsList, err := h.aggregatedRepo.FindTeamMetrics(ctx, params.Period, params.StartDate, params.EndDate)
	if err != nil {
		log.Printf("Failed to get team metrics: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "チームメトリクスの取得に失敗しました", nil)
		return
	}

	if len(metricsList) == 0 {
		h.writeErrorResponse(w, http.StatusNotFound, "NO_DATA", "指定された期間のデータが見つかりません", nil)
		return
	}

	// 最新のメトリクスを返す
	latest := metricsList[0]
	response := h.presenter.ToTeamMetricsResponse(latest)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetDeveloperMetrics は開発者メトリクスを取得
func (h *AnalyticsHandler) GetDeveloperMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	developer := vars["developer"]

	if developer == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_DEVELOPER", "開発者名が指定されていません", nil)
		return
	}

	// パラメータ解析
	params, err := h.parseAnalyticsParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// 開発者メトリクスを取得
	metricsList, err := h.aggregatedRepo.FindDeveloperMetrics(ctx, developer, params.Period, params.StartDate, params.EndDate)
	if err != nil {
		log.Printf("Failed to get developer metrics: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "開発者メトリクスの取得に失敗しました", nil)
		return
	}

	if len(metricsList) == 0 {
		h.writeErrorResponse(w, http.StatusNotFound, "NO_DATA", "指定された開発者のデータが見つかりません", nil)
		return
	}

	// 最新のメトリクスを返す
	latest := metricsList[0]
	response := h.presenter.ToDeveloperMetricsResponse(latest)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetRepositoryMetrics はリポジトリメトリクスを取得
func (h *AnalyticsHandler) GetRepositoryMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	repository := vars["repository"]

	if repository == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REPOSITORY", "リポジトリ名が指定されていません", nil)
		return
	}

	// パラメータ解析
	params, err := h.parseAnalyticsParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// リポジトリメトリクスを取得
	metricsList, err := h.aggregatedRepo.FindRepositoryMetrics(ctx, repository, params.Period, params.StartDate, params.EndDate)
	if err != nil {
		log.Printf("Failed to get repository metrics: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "リポジトリメトリクスの取得に失敗しました", nil)
		return
	}

	if len(metricsList) == 0 {
		h.writeErrorResponse(w, http.StatusNotFound, "NO_DATA", "指定されたリポジトリのデータが見つかりません", nil)
		return
	}

	// 最新のメトリクスを返す
	latest := metricsList[0]
	response := h.presenter.ToRepositoryMetricsResponse(latest)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// ListTeamMetrics はチームメトリクスの一覧を取得
func (h *AnalyticsHandler) ListTeamMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// パラメータ解析
	params, err := h.parseListParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// チームメトリクス一覧を取得
	metricsList, err := h.aggregatedRepo.FindTeamMetrics(ctx, params.Period, params.StartDate, params.EndDate)
	if err != nil {
		log.Printf("Failed to get team metrics list: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "チームメトリクス一覧の取得に失敗しました", nil)
		return
	}

	// ページング処理
	totalCount := len(metricsList)
	startIndex := (params.Page - 1) * params.PageSize
	endIndex := startIndex + params.PageSize

	if startIndex >= totalCount {
		metricsList = []*analyticsApp.TeamMetrics{}
	} else if endIndex > totalCount {
		metricsList = metricsList[startIndex:]
	} else {
		metricsList = metricsList[startIndex:endIndex]
	}

	// フィルタ情報の構築
	filters := AppliedFiltersResponse{
		Period:    string(params.Period),
		DateRange: DateRangeResponse{Start: params.StartDate, End: params.EndDate},
	}

	// レスポンス形式に変換
	response := h.presenter.ToTeamMetricsListResponse(metricsList, totalCount, params.Page, params.PageSize, filters)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// ListDeveloperMetrics は開発者メトリクスの一覧を取得
func (h *AnalyticsHandler) ListDeveloperMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// パラメータ解析
	params, err := h.parseListParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// 全開発者メトリクスを取得
	developerMetricsMap, err := h.aggregatedRepo.FindAllDeveloperMetrics(ctx, params.Period, params.StartDate, params.EndDate)
	if err != nil {
		log.Printf("Failed to get developer metrics list: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "開発者メトリクス一覧の取得に失敗しました", nil)
		return
	}

	// マップをスライスに変換
	metricsList := make([]*analyticsApp.DeveloperMetrics, 0, len(developerMetricsMap))
	for _, metrics := range developerMetricsMap {
		metricsList = append(metricsList, metrics)
	}

	// ページング処理
	totalCount := len(metricsList)
	startIndex := (params.Page - 1) * params.PageSize
	endIndex := startIndex + params.PageSize

	if startIndex >= totalCount {
		metricsList = []*analyticsApp.DeveloperMetrics{}
	} else if endIndex > totalCount {
		metricsList = metricsList[startIndex:]
	} else {
		metricsList = metricsList[startIndex:endIndex]
	}

	// レスポンス形式に変換
	response := h.presenter.ToDeveloperMetricsListResponse(metricsList, totalCount, params.Page, params.PageSize)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// ListRepositoryMetrics はリポジトリメトリクスの一覧を取得
func (h *AnalyticsHandler) ListRepositoryMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// パラメータ解析
	params, err := h.parseListParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// 全リポジトリメトリクスを取得
	repositoryMetricsMap, err := h.aggregatedRepo.FindAllRepositoryMetrics(ctx, params.Period, params.StartDate, params.EndDate)
	if err != nil {
		log.Printf("Failed to get repository metrics list: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "リポジトリメトリクス一覧の取得に失敗しました", nil)
		return
	}

	// マップをスライスに変換
	metricsList := make([]*analyticsApp.RepositoryMetrics, 0, len(repositoryMetricsMap))
	for _, metrics := range repositoryMetricsMap {
		metricsList = append(metricsList, metrics)
	}

	// ページング処理
	totalCount := len(metricsList)
	startIndex := (params.Page - 1) * params.PageSize
	endIndex := startIndex + params.PageSize

	if startIndex >= totalCount {
		metricsList = []*analyticsApp.RepositoryMetrics{}
	} else if endIndex > totalCount {
		metricsList = metricsList[startIndex:]
	} else {
		metricsList = metricsList[startIndex:endIndex]
	}

	// レスポンス形式に変換
	response := h.presenter.ToRepositoryMetricsListResponse(metricsList, totalCount, params.Page, params.PageSize)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetTrends はトレンド分析を取得
func (h *AnalyticsHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// パラメータ解析
	params, err := h.parseAnalyticsParams(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PARAMETERS", err.Error(), nil)
		return
	}

	// チームメトリクスからトレンドを取得
	metricsList, err := h.aggregatedRepo.FindTeamMetrics(ctx, params.Period, params.StartDate, params.EndDate)
	if err != nil {
		log.Printf("Failed to get trends: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "トレンドデータの取得に失敗しました", nil)
		return
	}

	if len(metricsList) == 0 {
		h.writeErrorResponse(w, http.StatusNotFound, "NO_DATA", "トレンド分析に必要なデータが見つかりません", nil)
		return
	}

	// 最新のトレンド分析を返す
	latest := metricsList[0]
	response := h.presenter.toTrendAnalysisResponse(latest.TrendAnalysis)
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetAnalyticsHealthCheck は集計データシステムのヘルスチェック
func (h *AnalyticsHandler) GetAnalyticsHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 集計データの統計情報を取得してヘルスチェック
	stats, err := h.aggregatedRepo.GetAggregatedStatistics(ctx)
	
	status := "healthy"
	services := map[string]string{
		"aggregated_data": "available",
		"analytics":       "operational",
	}

	if err != nil {
		status = "unhealthy"
		services["aggregated_data"] = "unavailable"
		log.Printf("Analytics health check failed: %v", err)
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"services":  services,
		"statistics": stats,
	}

	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	h.writeJSONResponse(w, statusCode, response)
}

// プライベートメソッド

type AnalyticsParams struct {
	Period    analyticsApp.AggregationPeriod
	StartDate time.Time
	EndDate   time.Time
}

type ListParams struct {
	AnalyticsParams
	Page     int
	PageSize int
}

func (h *AnalyticsHandler) parseAnalyticsParams(r *http.Request) (*AnalyticsParams, error) {
	query := r.URL.Query()

	// 期間パラメータ
	periodStr := query.Get("period")
	if periodStr == "" {
		periodStr = "monthly"
	}

	var period analyticsApp.AggregationPeriod
	switch strings.ToLower(periodStr) {
	case "daily":
		period = analyticsApp.AggregationPeriodDaily
	case "weekly":
		period = analyticsApp.AggregationPeriodWeekly
	case "monthly":
		period = analyticsApp.AggregationPeriodMonthly
	default:
		return nil, fmt.Errorf("invalid period: %s", periodStr)
	}

	// 日付範囲の計算
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0) // デフォルト1ヶ月

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

	return &AnalyticsParams{
		Period:    period,
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

func (h *AnalyticsHandler) parseListParams(r *http.Request) (*ListParams, error) {
	analyticsParams, err := h.parseAnalyticsParams(r)
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

	pageSize := 20
	if pageSizeStr := query.Get("pageSize"); pageSizeStr != "" {
		parsed, err := strconv.Atoi(pageSizeStr)
		if err != nil || parsed < 1 || parsed > 100 {
			return nil, fmt.Errorf("invalid page size: %s", pageSizeStr)
		}
		pageSize = parsed
	}

	return &ListParams{
		AnalyticsParams: *analyticsParams,
		Page:            page,
		PageSize:        pageSize,
	}, nil
}

func (h *AnalyticsHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func (h *AnalyticsHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
	errorResponse := AnalyticsErrorResponse{
		Error:     http.StatusText(statusCode),
		Code:      code,
		Message:   message,
		Details:   details,
		RequestID: h.generateRequestID(),
		Timestamp: time.Now(),
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}

func (h *AnalyticsHandler) generateRequestID() string {
	// 簡単なリクエストID生成（実際にはUUIDを使用）
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// RegisterRoutes はルートを登録
func (h *AnalyticsHandler) RegisterRoutes(router *mux.Router) {
	// チームメトリクス
	router.HandleFunc("/api/analytics/team_metrics", h.GetTeamMetrics).Methods("GET")
	router.HandleFunc("/api/analytics/team_metrics/list", h.ListTeamMetrics).Methods("GET")
	
	// 開発者メトリクス
	router.HandleFunc("/api/analytics/developer_metrics/{developer}", h.GetDeveloperMetrics).Methods("GET")
	router.HandleFunc("/api/analytics/developer_metrics", h.ListDeveloperMetrics).Methods("GET")
	
	// リポジトリメトリクス
	router.HandleFunc("/api/analytics/repository_metrics/{repository}", h.GetRepositoryMetrics).Methods("GET")
	router.HandleFunc("/api/analytics/repository_metrics", h.ListRepositoryMetrics).Methods("GET")
	
	// トレンド分析
	router.HandleFunc("/api/analytics/trends", h.GetTrends).Methods("GET")
	
	// ヘルスチェック
	router.HandleFunc("/api/analytics/health", h.GetAnalyticsHealthCheck).Methods("GET")
}