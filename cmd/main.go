package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"backend/api"
	"backend/configs"
	"backend/database"
	"backend/middleware"
	"backend/models"
	"backend/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	configs.LoadConfig()
	log.Println("Configuration loaded successfully")

	// 初始化数据库连接
	database.InitMySQL()
	database.InitRedis()
	log.Println("Database connections initialized")

	// 数据库迁移
	models.AutoMigrate()
	models.CreateIndexes()
	models.SeedDefaultData()

	// 设置Gin模式
	if configs.AppConfig.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.New()

	// 添加中间件
	setupMiddleware(r)

	// 设置路由
	api.SetupRoutes(r)                 // 健康检查和静态文件路由
	router.SetupRouter(r, database.DB) // 业务路由

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", configs.AppConfig.Server.Host, configs.AppConfig.Server.Port)
	log.Printf("Server starting on %s", addr)

	// 优雅关闭
	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 关闭数据库连接
	if err := database.CloseDB(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
	if err := database.CloseRedis(); err != nil {
		log.Printf("Error closing Redis: %v", err)
	}

	log.Println("Server shutdown complete")
}

// setupMiddleware 设置中间件
func setupMiddleware(r *gin.Engine) {
	// 恢复中间件
	r.Use(middleware.RecoveryMiddleware())

	// 日志中间件
	if configs.AppConfig.Server.Env != "production" {
		r.Use(middleware.RequestLoggerMiddleware())
	} else {
		r.Use(middleware.LoggerMiddleware())
	}

	// CORS中间件
	r.Use(middleware.CORSMiddleware())

	// 错误处理中间件
	r.Use(middleware.ErrorHandlerMiddleware())

	// API限流中间件
	r.Use(middleware.APIRateLimitMiddleware())
}
