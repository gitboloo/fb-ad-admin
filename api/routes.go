package api

import (
	"time"

	"backend/database"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// SetupRoutes 设置路由
// 注意: 此函数已废弃，所有业务路由已迁移到 router 包
// 保留此文件仅为了健康检查和静态文件服务
func SetupRoutes(r *gin.Engine) {
	// 静态文件服务
	r.Static("/uploads", "./uploads")
	r.Static("/static", "./static")

	// 健康检查
	r.GET("/health", healthCheck)

	// 详细健康检查（包含数据库和Redis状态）
	r.GET("/api/health", detailedHealthCheck)

	// 所有 API 路由已迁移到 router.SetupRouter()
}

// healthCheck 简单健康检查
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "ok",
		"message":   "API is running",
		"timestamp": time.Now().Unix(),
	})
}

// detailedHealthCheck 详细健康检查
func detailedHealthCheck(c *gin.Context) {
	health := gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"services":  gin.H{},
	}

	// 检查数据库连接
	dbStatus := "ok"
	dbMessage := "Database is connected"
	if err := database.HealthCheck(); err != nil {
		dbStatus = "error"
		dbMessage = err.Error()
		health["status"] = "degraded"
	}

	// 获取数据库连接池状态
	var poolStats interface{}
	if stats, err := database.GetPoolStats(); err == nil {
		poolStats = gin.H{
			"maxOpenConnections": stats.MaxOpenConnections,
			"openConnections":    stats.OpenConnections,
			"inUse":              stats.InUse,
			"idle":               stats.Idle,
			"waitCount":          stats.WaitCount,
			"waitDuration":       stats.WaitDuration.String(),
			"maxIdleClosed":      stats.MaxIdleClosed,
			"maxIdleTimeClosed":  stats.MaxIdleTimeClosed,
			"maxLifetimeClosed":  stats.MaxLifetimeClosed,
		}
	}

	health["services"].(gin.H)["database"] = gin.H{
		"status":  dbStatus,
		"message": dbMessage,
		"pool":    poolStats,
	}

	// 检查Redis连接
	redisStatus := "ok"
	redisMessage := "Redis is connected"
	if err := database.RedisHealthCheck(); err != nil {
		redisStatus = "error"
		redisMessage = err.Error()
		health["status"] = "degraded"
	}

	health["services"].(gin.H)["redis"] = gin.H{
		"status":  redisStatus,
		"message": redisMessage,
	}

	// 系统信息
	health["system"] = gin.H{
		"version": "1.0.0",
		"uptime":  time.Since(startTime).String(),
	}

	statusCode := 200
	if health["status"] != "ok" {
		statusCode = 503
	}

	c.JSON(statusCode, health)
}
