package utils

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"

	"github.com/ad-platform/backend/internal/types"
	"github.com/gin-gonic/gin"
)

// jsonResponse 发送不转义Unicode的JSON响应
func jsonResponse(c *gin.Context, status int, obj interface{}) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(obj); err != nil {
		c.JSON(status, gin.H{"error": "编码错误"})
		return
	}
	
	c.Data(status, "application/json; charset=utf-8", buffer.Bytes())
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	jsonResponse(c, http.StatusOK, types.Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应带消息
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, types.Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// PagedSuccess 分页成功响应
func PagedSuccess(c *gin.Context, data interface{}, total int64, page, size int) {
	pages := int(math.Ceil(float64(total) / float64(size)))
	c.JSON(http.StatusOK, types.PagedResponse{
		Code:    200,
		Message: "success",
		Data:    data,
		Total:   total,
		Page:    page,
		Size:    size,
		Pages:   pages,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, types.Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalServerError 500错误
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// ServerError 服务器错误（500）- 别名
func ServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// ValidateError 参数验证错误
func ValidateError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, types.ErrorResponse{
		Error: "参数验证失败",
		Details: map[string]interface{}{
			"validation_error": err.Error(),
		},
	})
}

// Created 201创建成功
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, types.Response{
		Code:    201,
		Message: "创建成功",
		Data:    data,
	})
}

// Updated 更新成功
func Updated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, types.Response{
		Code:    200,
		Message: "更新成功",
		Data:    data,
	})
}

// Deleted 删除成功
func Deleted(c *gin.Context) {
	c.JSON(http.StatusOK, types.Response{
		Code:    200,
		Message: "删除成功",
	})
}