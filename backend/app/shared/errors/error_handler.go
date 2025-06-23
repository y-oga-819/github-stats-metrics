package errors

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// AppError はアプリケーション全体で使用する統一エラー型
type AppError struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	StatusCode int         `json:"-"`
	Internal   error       `json:"-"`
}

func (e AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (internal: %v)", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e AppError) Unwrap() error {
	return e.Internal
}

// エラーコード定数
const (
	// バリデーションエラー
	ErrCodeInvalidRequest     = "INVALID_REQUEST"
	ErrCodeMissingParameter   = "MISSING_PARAMETER"
	ErrCodeInvalidDateRange   = "INVALID_DATE_RANGE"
	ErrCodeInvalidDeveloper   = "INVALID_DEVELOPER"
	
	// リポジトリエラー
	ErrCodeDatabaseError      = "DATABASE_ERROR"
	ErrCodeExternalAPIError   = "EXTERNAL_API_ERROR"
	ErrCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	
	// ビジネスルールエラー
	ErrCodeBusinessRuleViolation = "BUSINESS_RULE_VIOLATION"
	ErrCodeResourceNotFound      = "RESOURCE_NOT_FOUND"
	
	// システムエラー
	ErrCodeInternalError         = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable    = "SERVICE_UNAVAILABLE"
)

// NewValidationError はバリデーションエラーを生成
func NewValidationError(code, message string, details interface{}) AppError {
	return AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusBadRequest,
	}
}

// NewRepositoryError はリポジトリエラーを生成
func NewRepositoryError(code, message string, internal error) AppError {
	statusCode := http.StatusInternalServerError
	
	// エラーコードに応じたHTTPステータスコードの設定
	switch code {
	case ErrCodeUnauthorized:
		statusCode = http.StatusUnauthorized
	case ErrCodeRateLimitExceeded:
		statusCode = http.StatusTooManyRequests
	case ErrCodeExternalAPIError:
		statusCode = http.StatusBadGateway
	}
	
	return AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Internal:   internal,
	}
}

// NewBusinessRuleError はビジネスルールエラーを生成
func NewBusinessRuleError(code, message string, details interface{}) AppError {
	statusCode := http.StatusBadRequest
	
	if code == ErrCodeResourceNotFound {
		statusCode = http.StatusNotFound
	}
	
	return AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
	}
}

// NewInternalError はシステムエラーを生成
func NewInternalError(message string, internal error) AppError {
	return AppError{
		Code:       ErrCodeInternalError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Internal:   internal,
	}
}

// ErrorHandler は統一エラーハンドリング
type ErrorHandler struct {
	logger Logger
}

// Logger インターフェース
type Logger interface {
	Error(message string, fields ...interface{})
	Warn(message string, fields ...interface{})
	Info(message string, fields ...interface{})
}

// NewErrorHandler はErrorHandlerのコンストラクタ
func NewErrorHandler(logger Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError はエラーを適切に処理してHTTPレスポンスを返す
func (eh *ErrorHandler) HandleError(w http.ResponseWriter, err error) {
	var appErr AppError
	
	// AppErrorへの変換を試行
	if e, ok := err.(AppError); ok {
		appErr = e
	} else {
		// 未知のエラーは内部エラーとして扱う
		appErr = NewInternalError("An unexpected error occurred", err)
	}
	
	// ログ出力
	eh.logError(appErr)
	
	// HTTPレスポンス
	eh.writeErrorResponse(w, appErr)
}

// logError はエラーレベルに応じてログ出力
func (eh *ErrorHandler) logError(err AppError) {
	logFields := []interface{}{
		"code", err.Code,
		"message", err.Message,
		"statusCode", err.StatusCode,
	}
	
	if err.Details != nil {
		logFields = append(logFields, "details", err.Details)
	}
	
	if err.Internal != nil {
		logFields = append(logFields, "internal", err.Internal.Error())
	}
	
	// エラーレベルの判定
	if err.StatusCode >= 500 {
		eh.logger.Error("Server error occurred", logFields...)
	} else if err.StatusCode >= 400 {
		eh.logger.Warn("Client error occurred", logFields...)
	} else {
		eh.logger.Info("Error handled", logFields...)
	}
}

// writeErrorResponse はHTTPエラーレスポンスを書き込み
func (eh *ErrorHandler) writeErrorResponse(w http.ResponseWriter, err AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	
	// 内部エラー情報は除外してレスポンス
	response := struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	}{
		Code:    err.Code,
		Message: err.Message,
		Details: err.Details,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// JSON エンコードに失敗した場合のフォールバック
		log.Printf("Failed to encode error response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// WrapDomainError はドメイン層のエラーをAppErrorに変換
func WrapDomainError(err error) AppError {
	if err == nil {
		return AppError{}
	}
	
	// ドメインエラー型の判定（型アサーション）
	switch e := err.(type) {
	case interface{ IsValidationError() bool }:
		if e.IsValidationError() {
			return NewValidationError(ErrCodeInvalidRequest, err.Error(), nil)
		}
	case interface{ IsBusinessRuleError() bool }:
		if e.IsBusinessRuleError() {
			return NewBusinessRuleError(ErrCodeBusinessRuleViolation, err.Error(), nil)
		}
	}
	
	// その他のエラーは内部エラーとして扱う
	return NewInternalError("Domain operation failed", err)
}