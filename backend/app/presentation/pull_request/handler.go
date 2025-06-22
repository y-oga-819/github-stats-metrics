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
		case usecase.ErrorTypeValidation:
			h.respondError(w, http.StatusBadRequest, "Validation failed", err)
		case usecase.ErrorTypeRepository:
			h.respondError(w, http.StatusInternalServerError, "Data access failed", err)
		case usecase.ErrorTypeBusinessRule:
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