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

// エラータイプ定数
const (
	ErrorTypeValidation    = "VALIDATION_ERROR"
	ErrorTypeRepository    = "REPOSITORY_ERROR"
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

// NewBusinessRuleError はビジネスルールエラーを生成
func NewBusinessRuleError(message string, cause error) UseCaseError {
	return UseCaseError{
		Type:    ErrorTypeBusinessRule,
		Message: message,
		Cause:   cause,
	}
}