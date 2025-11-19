package middleware

import (
	"fmt"
	"log"
	"runtime/debug"

	"backend/utils"
	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware 错误恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 记录详细的错误信息
		err := fmt.Sprintf("Panic recovered: %v", recovered)
		stack := string(debug.Stack())
		
		// 输出到日志
		log.Printf("PANIC: %s\nStack trace:\n%s", err, stack)
		
		// 返回500错误
		utils.ServerError(c, "服务器内部错误")
	})
}

// ErrorHandlerMiddleware 统一错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// 记录错误日志
			log.Printf("Request error: %s %s - %v", c.Request.Method, c.Request.URL.Path, err.Error())
			
			// 如果还没有响应，返回错误响应
			if !c.Writer.Written() {
				utils.ServerError(c, "请求处理出错")
			}
		}
	}
}