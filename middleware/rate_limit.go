package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"backend/database"
	"backend/utils"
	"github.com/gin-gonic/gin"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	KeyPrefix string        // Redis键前缀
	Limit     int           // 限制次数
	Window    time.Duration // 时间窗口
}

// RateLimitMiddleware 基于Redis的限流中间件
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Redis客户端
		redisClient := database.GetRedis()

		// 如果Redis不可用，跳过限流直接放行
		if redisClient == nil {
			c.Next()
			return
		}

		// 获取客户端IP作为限流键的一部分
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%s:%s", config.KeyPrefix, clientIP)
		ctx := context.Background()

		// 使用Redis的INCR命令实现计数器
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			// Redis错误，允许请求通过
			c.Next()
			return
		}

		// 如果是第一次访问，设置过期时间
		if count == 1 {
			redisClient.Expire(ctx, key, config.Window)
		}

		// 检查是否超过限制
		if count > int64(config.Limit) {
			// 获取剩余时间
			ttl, _ := redisClient.TTL(ctx, key).Result()
			
			// 设置响应头
			c.Header("X-Rate-Limit-Limit", strconv.Itoa(config.Limit))
			c.Header("X-Rate-Limit-Remaining", "0")
			c.Header("X-Rate-Limit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))
			
			utils.Error(c, 429, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		// 设置响应头
		c.Header("X-Rate-Limit-Limit", strconv.Itoa(config.Limit))
		c.Header("X-Rate-Limit-Remaining", strconv.FormatInt(int64(config.Limit)-count, 10))

		c.Next()
	}
}

// LoginRateLimitMiddleware 登录限流中间件
func LoginRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		KeyPrefix: "login_limit",
		Limit:     5,                // 5次尝试
		Window:    15 * time.Minute, // 15分钟窗口
	})
}

// APIRateLimitMiddleware API限流中间件
func APIRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		KeyPrefix: "api_limit",
		Limit:     100,             // 100次请求
		Window:    1 * time.Minute, // 1分钟窗口
	})
}

// UserBasedRateLimitMiddleware 基于用户的限流中间件
func UserBasedRateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Redis客户端
		redisClient := database.GetRedis()

		// 如果Redis不可用，跳过限流直接放行
		if redisClient == nil {
			c.Next()
			return
		}

		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			// 如果没有用户信息，使用IP限流
			APIRateLimitMiddleware()(c)
			return
		}

		key := fmt.Sprintf("user_limit:%v", userID)
		ctx := context.Background()

		// 使用Redis的INCR命令实现计数器
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		// 如果是第一次访问，设置过期时间
		if count == 1 {
			redisClient.Expire(ctx, key, window)
		}

		// 检查是否超过限制
		if count > int64(limit) {
			ttl, _ := redisClient.TTL(ctx, key).Result()
			
			c.Header("X-Rate-Limit-Limit", strconv.Itoa(limit))
			c.Header("X-Rate-Limit-Remaining", "0")
			c.Header("X-Rate-Limit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))
			
			utils.Error(c, 429, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Header("X-Rate-Limit-Limit", strconv.Itoa(limit))
		c.Header("X-Rate-Limit-Remaining", strconv.FormatInt(int64(limit)-count, 10))

		c.Next()
	}
}