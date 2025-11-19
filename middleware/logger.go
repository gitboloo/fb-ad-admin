package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 自定义日志格式
		return fmt.Sprintf("[%s] %s %s %s %d %s \"%s\" %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.ClientIP,
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RequestLoggerMiddleware 请求日志中间件（更详细的日志）
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(startTime)

		// 获取请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 获取用户信息（如果已认证）
		userID, exists := c.Get("user_id")
		var userIDStr string
		if exists {
			userIDStr = fmt.Sprintf(" UserID:%v", userID)
		}

		// 输出日志
		fmt.Printf("[%s] %s %s %s %d %v%s \"%s\"\n",
			time.Now().Format("2006-01-02 15:04:05"),
			clientIP,
			method,
			path,
			statusCode,
			latency,
			userIDStr,
			userAgent,
		)
	}
}