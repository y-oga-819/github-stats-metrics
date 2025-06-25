package health

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler はヘルスチェック用のHTTPハンドラー
type Handler struct{}

// NewHandler はHandlerのコンストラクタ
func NewHandler() *Handler {
	return &Handler{}
}

// HealthResponse はヘルスチェックレスポンスの構造
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
}

// GetHealth はヘルスチェックエンドポイント
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "github-stats-metrics",
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode health response", http.StatusInternalServerError)
	}
}