package constants

// 错误码定义
const (
	// 成功
	SUCCESS = 200

	// 客户端错误 4xx
	ERROR_INVALID_PARAMS     = 400
	ERROR_UNAUTHORIZED      = 401
	ERROR_FORBIDDEN         = 403
	ERROR_NOT_FOUND         = 404
	ERROR_METHOD_NOT_ALLOWED = 405
	ERROR_CONFLICT          = 409
	ERROR_RATE_LIMITED      = 429

	// 服务器错误 5xx
	ERROR_INTERNAL_SERVER   = 500
	ERROR_DATABASE         = 501
	ERROR_REDIS           = 502
	ERROR_THIRD_PARTY     = 503
)

// 错误消息
var ErrorMessages = map[int]string{
	SUCCESS:                 "成功",
	ERROR_INVALID_PARAMS:    "参数错误",
	ERROR_UNAUTHORIZED:      "未授权",
	ERROR_FORBIDDEN:         "权限不足",
	ERROR_NOT_FOUND:         "资源不存在",
	ERROR_METHOD_NOT_ALLOWED: "方法不允许",
	ERROR_CONFLICT:          "资源冲突",
	ERROR_RATE_LIMITED:      "请求过于频繁",
	ERROR_INTERNAL_SERVER:   "服务器内部错误",
	ERROR_DATABASE:         "数据库错误",
	ERROR_REDIS:           "缓存服务错误",
	ERROR_THIRD_PARTY:     "第三方服务错误",
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(code int) string {
	if msg, ok := ErrorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}