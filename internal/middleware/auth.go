package middleware

import (
	"strings"

	"github.com/ad-platform/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "缺少认证信息")
			c.Abort()
			return
		}

		// 检查Bearer token格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		tokenString := authHeader[7:] // 去掉"Bearer "
		
		// 解析token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			utils.Unauthorized(c, "无效的认证信息")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		
		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(requiredRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.Unauthorized(c, "认证信息不完整")
			c.Abort()
			return
		}

		role, ok := userRole.(int)
		if !ok {
			utils.ServerError(c, "用户角色信息错误")
			c.Abort()
			return
		}

		if role < requiredRole {
			utils.Forbidden(c, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser 获取当前用户信息的辅助函数
func GetCurrentUser(c *gin.Context) (userID uint, username string, role int, exists bool) {
	userIDVal, exists1 := c.Get("user_id")
	usernameVal, exists2 := c.Get("username")
	roleVal, exists3 := c.Get("user_role")

	if !exists1 || !exists2 || !exists3 {
		return 0, "", 0, false
	}

	userID, ok1 := userIDVal.(uint)
	username, ok2 := usernameVal.(string)
	role, ok3 := roleVal.(int)

	if !ok1 || !ok2 || !ok3 {
		return 0, "", 0, false
	}

	return userID, username, role, true
}