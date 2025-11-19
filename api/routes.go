package api

import (
	"time"

	"backend/configs"
	"backend/database"
	"backend/middleware"
	"backend/models"
	"backend/router"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
	// 静态文件服务
	r.Static("/uploads", "./uploads")
	r.Static("/static", "./static")

	// 健康检查
	r.GET("/health", healthCheck)

	// 详细健康检查（包含数据库和Redis状态）
	r.GET("/api/health", detailedHealthCheck)

	// API版本分组
	api := r.Group("/api")
	{
		// 公开接口（不需要认证）
		public := api.Group("")
		setupPublicRoutes(public)

		// 需要认证的接口
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		setupProtectedRoutes(protected)

		// 管理员接口（需要管理员权限）
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.RequireRole(int(models.AdminRoleAdmin)))
		setupAdminRoutes(admin)

		// 超级管理员接口（需要超级管理员权限）
		superAdmin := api.Group("/super-admin")
		superAdmin.Use(middleware.AuthMiddleware())
		superAdmin.Use(middleware.RequireRole(int(models.AdminRoleSuperAdmin)))
		setupSuperAdminRoutes(superAdmin)
	}

	// 设置新的代理商路由系统(使用独立的router包)
	router.SetupRouter(r, database.GetDB())
}

// setupPublicRoutes 设置公开路由
func setupPublicRoutes(r *gin.RouterGroup) {
	adminController := NewAdminController()

	// 管理员登录
	auth := r.Group("/auth")
	{
		auth.POST("/login", middleware.LoginRateLimitMiddleware(), adminController.Login)
	}

	// 代理商登录（公开接口） - 已迁移到新的router系统
	// agentAuth := r.Group("/agent/auth")
	// {
	// 	agentAuth.POST("/login", middleware.LoginRateLimitMiddleware(), AgentLogin(database.GetDB()))
	// }
}

// setupProtectedRoutes 设置需要认证的路由
func setupProtectedRoutes(r *gin.RouterGroup) {
	adminController := NewAdminController()
	permissionController := NewPermissionController()
	customerController := NewCustomerController()
	financeController := NewFinanceController()
	couponController := NewCouponController()
	authCodeController := NewAuthCodeController()

	// 认证相关路由
	auth := r.Group("/auth")
	{
		auth.GET("/me", adminController.GetProfile)                       // 获取当前用户信息
		auth.POST("/logout", adminController.Logout)                      // 退出登录
		auth.GET("/permissions", permissionController.GetUserPermissions) // 获取用户权限
		auth.GET("/menus", permissionController.GetUserMenus)             // 获取用户菜单
	}

	// 当前用户相关
	r.GET("/profile", adminController.GetProfile)
	r.PUT("/profile/password", adminController.UpdatePassword)

	// 客户个人资料管理
	r.GET("/customer/profile", customerController.GetProfile)
	r.PUT("/customer/profile", customerController.UpdateProfile)

	// 财务管理（用户）
	finance := r.Group("/finance")
	{
		finance.POST("/recharge", financeController.Recharge)
		finance.POST("/withdraw", financeController.Withdraw)
		finance.GET("/transactions", financeController.GetTransactions)
		finance.GET("/balance", financeController.GetBalance)
		finance.GET("/statistics", financeController.GetStatistics)
	}

	// 优惠券管理（用户）
	coupons := r.Group("/coupons")
	{
		coupons.GET("/user-coupons", couponController.GetUserCoupons)
		coupons.POST("/claim", couponController.ClaimCoupon)
		coupons.POST("/user-coupons/:id/use", couponController.UseCoupon)
		coupons.GET("/available", couponController.GetAvailableCoupons)
	}

	// 授权码验证（用户）
	authcodes := r.Group("/authcodes")
	{
		authcodes.POST("/verify", authCodeController.Verify)
		authcodes.GET("/my-used", authCodeController.GetMyUsedCodes)
		authcodes.GET("/check", authCodeController.CheckCodeAvailability)
		authcodes.POST("/validate-format", authCodeController.ValidateCodeFormat)
	}

	// ============================================
	// 代理商自服务路由 - 已迁移到新的router系统
	// ============================================
	// agent := r.Group("/agent")
	// {
	// 	// 个人信息管理
	// 	agent.GET("/profile", GetAgentProfile(database.GetDB()))
	// 	agent.PUT("/profile", UpdateAgentProfile(database.GetDB()))

	// 	// 仪表板
	// 	agent.GET("/dashboard", GetAgentDashboard(database.GetDB()))

	// 	// 客户管理
	// 	agent.GET("/customers", GetAgentCustomers(database.GetDB()))

	// 	// 佣金管理
	// 	agent.GET("/commissions", GetAgentCommissions(database.GetDB()))
	// 	agent.GET("/commissions/summary", GetCommissionSummary(database.GetDB()))

	// 	// 提现管理
	// 	agent.POST("/withdraw", CreateWithdrawal(database.GetDB()))
	// 	agent.GET("/withdrawals", GetAgentWithdrawals(database.GetDB()))
	// }
}

// setupAdminRoutes 设置管理员路由
func setupAdminRoutes(r *gin.RouterGroup) {
	adminController := NewAdminController()
	permissionController := NewPermissionController()
	productController := NewProductController()
	campaignController := NewCampaignController()
	couponController := NewCouponController()
	customerController := NewCustomerController()
	financeController := NewFinanceController()
	authCodeController := NewAuthCodeController()
	statisticsController := NewStatisticsController()

	// 管理员管理
	admins := r.Group("/admins")
	{
		admins.GET("", adminController.List)
		admins.GET("/:id", adminController.GetByID)
		admins.PUT("/:id", adminController.Update)
		admins.PUT("/:id/status", adminController.UpdateStatus)
		admins.DELETE("/:id", adminController.Delete)
		admins.POST("/:id/reset-password", adminController.ResetPassword)
		admins.POST("/:id/roles", permissionController.AssignRolesToUser)
	}

	// 权限管理
	permissions := r.Group("/permissions")
	{
		permissions.GET("", permissionController.GetAllPermissions)
		permissions.POST("", permissionController.CreatePermission)
		permissions.PUT("/:id", permissionController.UpdatePermission)
		permissions.DELETE("/:id", permissionController.DeletePermission)
	}

	// 角色管理已迁移到 router/router.go 中的新 RoleController

	// 产品管理
	products := r.Group("/products")
	{
		products.GET("", productController.List)
		products.GET("/:id", productController.GetByID)
		products.POST("", productController.Create)
		products.PUT("/:id", productController.Update)
		products.DELETE("/:id", productController.Delete)
		products.POST("/:id/upload-logo", productController.UploadLogo)
		products.POST("/:id/upload-images", productController.UploadImages)
		products.GET("/statistics", productController.GetStatistics)
	}

	// 计划管理
	campaigns := r.Group("/campaigns")
	{
		campaigns.GET("", campaignController.List)
		campaigns.GET("/:id", campaignController.GetByID)
		campaigns.POST("", campaignController.Create)
		campaigns.PUT("/:id", campaignController.Update)
		campaigns.DELETE("/:id", campaignController.Delete)
		campaigns.POST("/:id/upload-image", campaignController.UploadMainImage)
		campaigns.POST("/:id/upload-video", campaignController.UploadVideo)
		campaigns.GET("/:id/stats", campaignController.GetStatistics)
		campaigns.PUT("/:id/status", campaignController.UpdateStatus)
		campaigns.POST("/:id/pause", campaignController.Pause)
		campaigns.POST("/:id/resume", campaignController.Resume)
	}

	// 优惠券管理
	coupons := r.Group("/coupons")
	{
		coupons.GET("", couponController.List)
		coupons.GET("/:id", couponController.GetByID)
		coupons.POST("", couponController.Create)
		coupons.PUT("/:id", couponController.Update)
		coupons.DELETE("/:id", couponController.Delete)
		coupons.POST("/:id/distribute", couponController.Distribute)
		coupons.GET("/statistics", couponController.GetStatistics)
	}

	// 客户管理
	customers := r.Group("/customers")
	{
		customers.GET("", customerController.List)
		customers.GET("/:id", customerController.GetByID)
		customers.POST("", customerController.Create)
		customers.PUT("/:id", customerController.Update)
		customers.DELETE("/:id", customerController.Delete)
		customers.PUT("/:id/status", customerController.UpdateStatus)
		customers.POST("/:id/block", customerController.Block)
		customers.POST("/:id/unblock", customerController.Unblock)
		customers.GET("/:id/transactions", customerController.GetTransactions)
		customers.GET("/:id/coupons", customerController.GetCoupons)
		customers.PUT("/:id/balance", customerController.UpdateBalance)
		customers.GET("/statistics", customerController.GetStatistics)
		customers.GET("/export", customerController.Export)
		customers.POST("/batch-status", customerController.BatchUpdateStatus)
	}

	// 财务管理（管理员）
	finance := r.Group("/finance")
	{
		finance.GET("/transactions", financeController.AdminGetAllTransactions)
		finance.GET("/statistics", financeController.AdminGetStatistics)
		finance.POST("/transactions/:id/process", financeController.ProcessTransaction)
		finance.GET("/pending", financeController.GetPendingTransactions)
		finance.GET("/type/:type", financeController.GetTransactionsByType)
		finance.GET("/dashboard", financeController.GetDashboardStats)
		finance.GET("/export", financeController.ExportTransactions)
		finance.POST("/batch-process", financeController.BatchProcessTransactions)
	}

	// 授权码管理
	authcodes := r.Group("/authcodes")
	{
		authcodes.GET("", authCodeController.List)
		authcodes.GET("/:id", authCodeController.GetByID)
		authcodes.POST("/generate", authCodeController.Generate)
		authcodes.PUT("/:id/revoke", authCodeController.Revoke)
		authcodes.POST("/batch-revoke", authCodeController.BatchRevoke)
		authcodes.GET("/export", authCodeController.Export)
		authcodes.GET("/statistics", authCodeController.GetStatistics)
		authcodes.GET("/usage-history", authCodeController.GetUsageHistory)
		authcodes.GET("/expired", authCodeController.GetExpiredCodes)
		authcodes.POST("/clean-expired", authCodeController.CleanExpired)
		authcodes.GET("/code/:code", authCodeController.GetCodeByCode)
	}

	// 统计分析
	statistics := r.Group("/statistics")
	{
		statistics.GET("/overview", statisticsController.GetOverview)
		statistics.GET("/products", statisticsController.GetProducts)
		statistics.GET("/campaigns", statisticsController.GetCampaigns)
		statistics.GET("/coupons", statisticsController.GetCoupons)
		statistics.GET("/revenue", statisticsController.GetRevenue)
		statistics.GET("/users", statisticsController.GetUsers)
		statistics.GET("/trends", statisticsController.GetTrends)
		statistics.GET("/realtime", statisticsController.GetRealtime)
		statistics.GET("/top-performers", statisticsController.GetTopPerformers)
		statistics.POST("/comparison", statisticsController.GetComparison)
		statistics.GET("/forecast", statisticsController.GetForecast)
		statistics.POST("/custom-report", statisticsController.GetCustomReport)
		statistics.POST("/export-report", statisticsController.ExportReport)
		statistics.GET("/metrics", statisticsController.GetMetrics)
		statistics.GET("/dimensions", statisticsController.GetDimensions)
	}

	// ============================================
	// 代理商管理（管理员权限）- 已迁移到新的router系统
	// ============================================
	// agents := r.Group("/agents")
	// {
	// 	// 代理商CRUD
	// 	agents.GET("", GetAgentList(database.GetDB()))
	// 	agents.POST("", CreateAgent(database.GetDB()))
	// 	agents.GET("/:id", GetAgentDetail(database.GetDB()))
	// 	agents.PUT("/:id", UpdateAgent(database.GetDB()))

	// 	// 代理商审核
	// 	agents.POST("/:id/approve", ApproveAgent(database.GetDB()))
	// 	agents.POST("/:id/reject", RejectAgent(database.GetDB()))
	// 	agents.POST("/:id/freeze", FreezeAgent(database.GetDB()))
	// 	agents.POST("/:id/unfreeze", UnfreezeAgent(database.GetDB()))
	// }

	// // 提现审核（管理员权限）
	// withdrawals := r.Group("/withdrawals")
	// {
	// 	withdrawals.GET("", GetWithdrawalList(database.GetDB()))
	// 	withdrawals.POST("/:id/approve", ApproveWithdrawal(database.GetDB()))
	// 	withdrawals.POST("/:id/reject", RejectWithdrawal(database.GetDB()))
	// 	withdrawals.POST("/:id/complete", CompleteWithdrawal(database.GetDB()))
	// }
}

// setupSuperAdminRoutes 设置超级管理员路由
func setupSuperAdminRoutes(r *gin.RouterGroup) {
	adminController := NewAdminController()
	systemController := NewSystemController()

	// 管理员创建（仅超级管理员）
	r.POST("/admins", adminController.Create)

	// 系统管理
	system := r.Group("/system")
	{
		system.GET("/configs", systemController.GetConfigs)
		system.PUT("/configs", systemController.UpdateConfigs)
		system.GET("/config/:key", systemController.GetConfig)
		system.PUT("/config/:key", systemController.UpdateConfig)
		system.GET("/stats", systemController.GetStats)
		system.GET("/dashboard", systemController.GetDashboard)
		system.POST("/backup", systemController.Backup)
		system.POST("/restore", systemController.Restore)
		system.GET("/backups", systemController.GetBackups)
		system.DELETE("/backup/:filename", systemController.DeleteBackup)
		system.GET("/info", systemController.GetSystemInfo)
		system.PUT("/info", systemController.UpdateSystemInfo)
		system.GET("/maintenance", systemController.GetMaintenanceMode)
		system.POST("/maintenance", systemController.SetMaintenanceMode)
		system.POST("/clean", systemController.CleanSystem)
		system.GET("/health", systemController.GetHealth)
		system.POST("/init", systemController.InitSystem)
		system.POST("/reset", systemController.ResetSystem)
		system.POST("/export", systemController.ExportData)
		system.POST("/import", systemController.ImportData)
	}
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
		"version":     "1.0.0",
		"environment": configs.AppConfig.Server.Env,
		"uptime":      time.Since(startTime).String(),
	}

	statusCode := 200
	if health["status"] != "ok" {
		statusCode = 503
	}

	c.JSON(statusCode, health)
}

var startTime = time.Now()
