package services

import "fmt"

// ServiceError 服务层错误
type ServiceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ServiceError) Error() string {
	return e.Message
}

// NewServiceError 创建服务错误
func NewServiceError(code int, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

// 常用错误
var (
	ErrRecordNotFound = &ServiceError{Code: 404, Message: "记录不存在"}
	ErrInvalidParams  = &ServiceError{Code: 400, Message: "参数无效"}
	ErrUnauthorized   = &ServiceError{Code: 401, Message: "未授权"}
	ErrForbidden      = &ServiceError{Code: 403, Message: "禁止访问"}
	ErrInternalServer = &ServiceError{Code: 500, Message: "内部服务器错误"}
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors 验证错误集合
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	return e[0].Error()
}

// HasErrors 是否有错误
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// Add 添加错误
func (e *ValidationErrors) Add(field, message string) {
	*e = append(*e, ValidationError{
		Field:   field,
		Message: message,
	})
}

// BusinessError 业务逻辑错误
type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *BusinessError) Error() string {
	return e.Message
}

// NewBusinessError 创建业务错误
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// WithDetails 添加详细信息
func (e *BusinessError) WithDetails(details map[string]interface{}) *BusinessError {
	e.Details = details
	return e
}