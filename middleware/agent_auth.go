package middleware

import (
	"strings"

	"backend/utils"
	"github.com/gin-gonic/gin"
)

// AgentAuthMiddleware 代理商认证中间件
func AgentAuthMiddleware() gin.HandlerFunc {
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

		// 这里可以添加额外的代理商验证逻辑
		// 例如检查代理商状态是否激活等

		// 将代理商信息存储到上下文中
		c.Set("agent_id", claims.UserID)
		c.Set("agent_username", claims.Username)
		c.Set("agent_level", claims.Role) // 使用Role字段存储代理商等级

		c.Next()
	}
}

// GetCurrentAgent 获取当前代理商信息的辅助函数
func GetCurrentAgent(c *gin.Context) (agentID uint, username string, level int, exists bool) {
	agentIDVal, exists1 := c.Get("agent_id")
	usernameVal, exists2 := c.Get("agent_username")
	levelVal, exists3 := c.Get("agent_level")

	if !exists1 || !exists2 || !exists3 {
		return 0, "", 0, false
	}

	agentID, ok1 := agentIDVal.(uint)
	username, ok2 := usernameVal.(string)
	level, ok3 := levelVal.(int)

	if !ok1 || !ok2 || !ok3 {
		return 0, "", 0, false
	}

	return agentID, username, level, true
}
