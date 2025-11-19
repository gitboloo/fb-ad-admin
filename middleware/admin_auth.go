package middleware

import (
	"strings"

	"backend/utils"
	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
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

		// 验证是否为管理员角色 (role=1为超级管理员, role=2为运营)
		if claims.Role > 2 {
			utils.Forbidden(c, "仅限管理员访问")
			c.Abort()
			return
		}

		// 将管理员信息存储到上下文中
		c.Set("admin_id", claims.UserID)
		c.Set("admin_username", claims.Username)
		c.Set("admin_role", claims.Role)

		c.Next()
	}
}

// RequireAdminRole 管理员角色权限中间件
// role=1: 超级管理员, role=2: 运营
func RequireAdminRole(requiredRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminRole, exists := c.Get("admin_role")
		if !exists {
			utils.Unauthorized(c, "认证信息不完整")
			c.Abort()
			return
		}

		role, ok := adminRole.(int)
		if !ok {
			utils.ServerError(c, "管理员角色信息错误")
			c.Abort()
			return
		}

		// role值越小权限越高
		if role > requiredRole {
			utils.Forbidden(c, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentAdmin 获取当前管理员信息的辅助函数
func GetCurrentAdmin(c *gin.Context) (adminID uint, username string, role int, exists bool) {
	adminIDVal, exists1 := c.Get("admin_id")
	usernameVal, exists2 := c.Get("admin_username")
	roleVal, exists3 := c.Get("admin_role")

	if !exists1 || !exists2 || !exists3 {
		return 0, "", 0, false
	}

	adminID, ok1 := adminIDVal.(uint)
	username, ok2 := usernameVal.(string)
	role, ok3 := roleVal.(int)

	if !ok1 || !ok2 || !ok3 {
		return 0, "", 0, false
	}

	return adminID, username, role, true
}
