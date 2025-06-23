package pull_request

import "fmt"

// DomainError はドメイン層で発生するエラー
type DomainError struct {
	Type    string
	Message string
	Details interface{}
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// IsValidationError はバリデーションエラーかどうかを判定
func (e *DomainError) IsValidationError() bool {
	return e.Type == "VALIDATION_ERROR"
}

// IsBusinessRuleError はビジネスルールエラーかどうかを判定
func (e *DomainError) IsBusinessRuleError() bool {
	return e.Type == "BUSINESS_RULE_ERROR"
}

// NewValidationError はバリデーションエラーを生成
func NewValidationError(message string, details interface{}) *DomainError {
	return &DomainError{
		Type:    "VALIDATION_ERROR",
		Message: message,
		Details: details,
	}
}

// NewBusinessRuleError はビジネスルールエラーを生成
func NewBusinessRuleError(message string, details interface{}) *DomainError {
	return &DomainError{
		Type:    "BUSINESS_RULE_ERROR",
		Message: message,
		Details: details,
	}
}