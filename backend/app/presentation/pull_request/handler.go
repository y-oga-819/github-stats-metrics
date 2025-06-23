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
	"github-stats-metrics/shared/errors"
	"github-stats-metrics/shared/logger"
)

// Handler はPull Request APIのHTTPハンドラー
type Handler struct {
	useCase      *usecase.UseCase
	decoder      *schema.Decoder
	errorHandler *errors.ErrorHandler
}

// NewHandler はHandlerのコンストラクタ
func NewHandler(useCase *usecase.UseCase) *Handler {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true) // 不明なクエリパラメータを無視
	
	// 統一エラーハンドラーの初期化
	logger := logger.NewStandardLogger()
	errorHandler := errors.NewErrorHandler(logger)
	
	return &Handler{
		useCase:      useCase,
		decoder:      decoder,
		errorHandler: errorHandler,
	}
}

// GetPullRequests はPull RequestsのGETエンドポイント
func (h *Handler) GetPullRequests(w http.ResponseWriter, r *http.Request) {
	// リクエスト解析
	req, err := h.parseRequest(r)
	if err != nil {
		appErr := errors.NewValidationError(errors.ErrCodeInvalidRequest, "Invalid request parameters", err.Error())
		h.errorHandler.HandleError(w, appErr)
		return
	}
	
	// タイムアウト設定
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	// UseCase呼び出し
	pullRequests, err := h.useCase.GetPullRequests(ctx, *req)
	if err != nil {
		h.errorHandler.HandleError(w, err)
		return
	}
	
	// レスポンス返却
	h.respondSuccess(w, pullRequests)
}

// parseRequest はHTTPリクエストからDomainリクエストを生成
func (h *Handler) parseRequest(r *http.Request) (*domain.GetPullRequestsRequest, error) {
	req := &domain.GetPullRequestsRequest{}
	
	if err := h.decoder.Decode(req, r.URL.Query()); err != nil {
		return nil, fmt.Errorf("failed to decode query parameters: %w", err)
	}
	
	return req, nil
}

// respondSuccess は成功レスポンスを返却
func (h *Handler) respondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		appErr := errors.NewInternalError("Failed to encode response", err)
		h.errorHandler.HandleError(w, appErr)
	}
}