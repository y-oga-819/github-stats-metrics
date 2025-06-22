package todo

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	usecase "github-stats-metrics/application/todo"
)

// Handler はTodo APIのHTTPハンドラー
type Handler struct {
	useCase *usecase.UseCase
}

// NewHandler はHandlerのコンストラクタ
func NewHandler(useCase *usecase.UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

// GetTodos はTodosのGETエンドポイント
func (h *Handler) GetTodos(w http.ResponseWriter, r *http.Request) {
	// タイムアウト設定
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// UseCase呼び出し
	todos, err := h.useCase.GetTodos(ctx)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get todos", err)
		return
	}

	// レスポンス返却
	h.respondSuccess(w, todos)
}

// respondSuccess は成功レスポンスを返却
func (h *Handler) respondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to encode response", err)
	}
}

// respondError はエラーレスポンスを返却
func (h *Handler) respondError(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
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