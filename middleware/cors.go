package middleware

import (
	"backend/configs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     configs.AppConfig.CORS.AllowedOrigins,
		AllowMethods:     configs.AppConfig.CORS.AllowedMethods,
		AllowHeaders:     configs.AppConfig.CORS.AllowedHeaders,
		AllowCredentials: true,
	}

	// 如果允许所有来源，设置为通配符
	for _, origin := range config.AllowOrigins {
		if origin == "*" {
			config.AllowAllOrigins = true
			config.AllowOrigins = nil
			break
		}
	}

	return cors.New(config)
}