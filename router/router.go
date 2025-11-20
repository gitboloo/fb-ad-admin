package router

import (
	"backend/controllers/admin"
	"backend/controllers/client"
	"backend/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handlers 所有handler的集合
type Handlers struct {
	// Admin handlers
	AdminAgent      *admin.AgentHandler
	AdminRole       *admin.RoleHandler
	AdminDashboard  *admin.DashboardHandler
	AdminAuth       *admin.AuthHandler
	AdminProduct    *admin.ProductHandler
	AdminCampaign   *admin.CampaignHandler
	AdminCustomer   *admin.CustomerHandler
	AdminCoupon     *admin.CouponHandler
	AdminAuthCode   *admin.AuthCodeHandler
	AdminFinance    *admin.FinanceHandler
	AdminPermission *admin.PermissionHandler
	AdminStatistics *admin.StatisticsHandler
	AdminSystem     *admin.SystemHandler

	// Client handlers
	ClientAuth     *client.AuthHandler
	ClientCustomer *client.CustomerHandler
	ClientFinance  *client.FinanceHandler
	ClientCoupon   *client.CouponHandler
	ClientAuthCode *client.AuthCodeHandler
}

// NewHandlers 创建所有handlers
func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{
		// Admin handlers
		AdminAgent:      admin.NewAgentHandler(db),
		AdminRole:       admin.NewRoleHandler(db),
		AdminDashboard:  admin.NewDashboardHandler(),
		AdminAuth:       admin.NewAuthHandler(),
		AdminProduct:    admin.NewProductHandler(),
		AdminCampaign:   admin.NewCampaignHandler(),
		AdminCustomer:   admin.NewCustomerHandler(),
		AdminCoupon:     admin.NewCouponHandler(),
		AdminAuthCode:   admin.NewAuthCodeHandler(),
		AdminFinance:    admin.NewFinanceHandler(),
		AdminPermission: admin.NewPermissionHandler(),
		AdminStatistics: admin.NewStatisticsHandler(),
		AdminSystem:     admin.NewSystemHandler(),

		// Client handlers
		ClientAuth:     client.NewAuthHandler(),
		ClientCustomer: client.NewCustomerHandler(),
		ClientFinance:  client.NewFinanceHandler(),
		ClientCoupon:   client.NewCouponHandler(),
		ClientAuthCode: client.NewAuthCodeHandler(),
	}
}

// SetupRouter 设置所有路由
// 路由规范:
//   - 管理后台 API: /api/admin/*  (后台管理员使用)
//   - 客户端 API:   /api/cli/*    (客户端用户使用)
func SetupRouter(r *gin.Engine, db *gorm.DB) {
	handlers := NewHandlers(db)

	// API路由组
	api := r.Group("/api")
	{
		// 管理后台路由 (后台管理系统使用)
		SetupAdminRoutes(api, handlers)

		// 客户端路由 (客户端应用使用)
		SetupClientRoutes(api, handlers)
	}
}

// SetupAdminRoutes 设置管理后台路由
// 所有路由前缀: /api/admin
// 用于后台管理系统，包括管理员登录、代理商管理、角色权限管理等
func SetupAdminRoutes(api *gin.RouterGroup, h *Handlers) {
	admin := api.Group("/admin")
	{
		// 公开路由（不需要认证）
		auth := admin.Group("/auth")
		{
			auth.POST("/login", h.AdminAuth.Login) // 登录
		}

		// 受保护路由（需要认证）
		protected := admin.Group("")
		protected.Use(middleware.AdminAuthMiddleware())
		{
			// 认证相关
			authProtected := protected.Group("/auth")
			{
				authProtected.GET("/me", h.AdminAuth.GetProfile)              // 获取当前用户信息
				authProtected.POST("/logout", h.AdminAuth.Logout)             // 退出登录
				authProtected.GET("/permissions", h.AdminAuth.GetPermissions) // 获取用户权限
				authProtected.GET("/menus", h.AdminAuth.GetMenus)             // 获取用户菜单
				authProtected.PUT("/password", h.AdminAuth.UpdatePassword)    // 修改密码
			}

			// 代理商管理
			agents := protected.Group("/agents")
			{
				agents.GET("", h.AdminAgent.List)          // 列表
				agents.POST("", h.AdminAgent.Create)       // 创建
				agents.GET("/:id", h.AdminAgent.Detail)    // 详情
				agents.PUT("/:id", h.AdminAgent.Update)    // 更新
				agents.DELETE("/:id", h.AdminAgent.Delete) // 删除
			}

			// 角色管理
			roles := protected.Group("/roles")
			{
				roles.GET("", h.AdminRole.List)                               // 列表
				roles.POST("", h.AdminRole.Create)                            // 创建
				roles.GET("/:id", h.AdminRole.Detail)                         // 详情
				roles.PUT("/:id", h.AdminRole.Update)                         // 更新
				roles.DELETE("/:id", h.AdminRole.Delete)                      // 删除
				roles.POST("/:id/permissions", h.AdminRole.AssignPermissions) // 分配权限
				roles.GET("/permissions/tree", h.AdminRole.GetPermissions)    // 获取权限树
				roles.GET("/assignable", h.AdminRole.GetAssignableRoles)      // 获取可分配角色列表
			}

			// 仪表盘
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/stats", h.AdminDashboard.GetStats)                              // 统计数据
				dashboard.GET("/activities", h.AdminDashboard.GetActivities)                    // 最近活动
				dashboard.GET("/revenue-trend", h.AdminDashboard.GetRevenueTrend)               // 收入趋势
				dashboard.GET("/user-growth", h.AdminDashboard.GetUserGrowth)                   // 用户增长
				dashboard.GET("/campaign-performance", h.AdminDashboard.GetCampaignPerformance) // 广告计划表现
				dashboard.GET("/realtime", h.AdminDashboard.GetRealtime)                        // 实时数据
			}

			// 产品管理
			products := protected.Group("/products")
			{
				products.GET("", h.AdminProduct.List)                            // 列表
				products.GET("/:id", h.AdminProduct.GetByID)                     // 详情
				products.POST("", h.AdminProduct.Create)                         // 创建
				products.PUT("/:id", h.AdminProduct.Update)                      // 更新
				products.PATCH("/:id/status", h.AdminProduct.UpdateStatus)       // 更新状态
				products.DELETE("/:id", h.AdminProduct.Delete)                   // 删除
				products.POST("/:id/upload-logo", h.AdminProduct.UploadLogo)     // 上传Logo（需要ID）
				products.POST("/:id/upload-images", h.AdminProduct.UploadImages) // 上传图片（需要ID）
				products.POST("/upload", h.AdminProduct.UploadFile)              // 通用单文件上传
				products.POST("/upload-multiple", h.AdminProduct.UploadFiles)    // 通用多文件上传
				products.GET("/statistics", h.AdminProduct.GetStatistics)        // 统计
			}

			// 广告计划管理
			campaigns := protected.Group("/campaigns")
			{
				campaigns.GET("", h.AdminCampaign.List)                              // 列表
				campaigns.GET("/:id", h.AdminCampaign.GetByID)                       // 详情
				campaigns.POST("", h.AdminCampaign.Create)                           // 创建
				campaigns.PUT("/:id", h.AdminCampaign.Update)                        // 更新
				campaigns.DELETE("/:id", h.AdminCampaign.Delete)                     // 删除
				campaigns.POST("/:id/upload-image", h.AdminCampaign.UploadMainImage) // 上传主图
				campaigns.POST("/:id/upload-video", h.AdminCampaign.UploadVideo)     // 上传视频
				campaigns.GET("/:id/stats", h.AdminCampaign.GetStatistics)           // 统计
				campaigns.PUT("/:id/status", h.AdminCampaign.UpdateStatus)           // 更新状态
				campaigns.POST("/:id/pause", h.AdminCampaign.Pause)                  // 暂停
				campaigns.POST("/:id/resume", h.AdminCampaign.Resume)                // 恢复
			}

			// 客户管理
			customers := protected.Group("/customers")
			{
				customers.GET("", h.AdminCustomer.List)                             // 列表
				customers.GET("/:id", h.AdminCustomer.GetByID)                      // 详情
				customers.POST("", h.AdminCustomer.Create)                          // 创建
				customers.PUT("/:id", h.AdminCustomer.Update)                       // 更新
				customers.DELETE("/:id", h.AdminCustomer.Delete)                    // 删除
				customers.PUT("/:id/status", h.AdminCustomer.UpdateStatus)          // 更新状态
				customers.POST("/:id/block", h.AdminCustomer.Block)                 // 封禁
				customers.POST("/:id/unblock", h.AdminCustomer.Unblock)             // 解封
				customers.GET("/:id/transactions", h.AdminCustomer.GetTransactions) // 交易记录
				customers.GET("/:id/coupons", h.AdminCustomer.GetCoupons)           // 优惠券
				customers.PUT("/:id/balance", h.AdminCustomer.UpdateBalance)        // 更新余额
				customers.GET("/statistics", h.AdminCustomer.GetStatistics)         // 统计
				customers.GET("/export", h.AdminCustomer.Export)                    // 导出
				customers.POST("/batch-status", h.AdminCustomer.BatchUpdateStatus)  // 批量更新状态
			}

			// 优惠券管理
			coupons := protected.Group("/coupons")
			{
				coupons.GET("", h.AdminCoupon.List)                       // 列表
				coupons.GET("/:id", h.AdminCoupon.GetByID)                // 详情
				coupons.POST("", h.AdminCoupon.Create)                    // 创建
				coupons.PUT("/:id", h.AdminCoupon.Update)                 // 更新
				coupons.DELETE("/:id", h.AdminCoupon.Delete)              // 删除
				coupons.POST("/:id/distribute", h.AdminCoupon.Distribute) // 分发
				coupons.GET("/statistics", h.AdminCoupon.GetStatistics)   // 统计
			}

			// 授权码管理
			authcodes := protected.Group("/authcodes")
			{
				authcodes.GET("", h.AdminAuthCode.List)                          // 列表
				authcodes.GET("/:id", h.AdminAuthCode.GetByID)                   // 详情
				authcodes.POST("/generate", h.AdminAuthCode.Generate)            // 生成
				authcodes.POST("/verify", h.AdminAuthCode.Verify)                // 验证
				authcodes.PUT("/:id/revoke", h.AdminAuthCode.Revoke)             // 撤销
				authcodes.POST("/batch-revoke", h.AdminAuthCode.BatchRevoke)     // 批量撤销
				authcodes.GET("/export", h.AdminAuthCode.Export)                 // 导出
				authcodes.GET("/statistics", h.AdminAuthCode.GetStatistics)      // 统计
				authcodes.GET("/usage-history", h.AdminAuthCode.GetUsageHistory) // 使用历史
				authcodes.GET("/expired", h.AdminAuthCode.GetExpiredCodes)       // 过期授权码
				authcodes.POST("/clean-expired", h.AdminAuthCode.CleanExpired)   // 清理过期
				authcodes.GET("/code/:code", h.AdminAuthCode.GetCodeByCode)      // 按code查询
			}

			// 财务管理
			finance := protected.Group("/finance")
			{
				finance.GET("/transactions", h.AdminFinance.AdminGetAllTransactions)         // 所有交易
				finance.GET("/statistics", h.AdminFinance.AdminGetStatistics)                // 统计
				finance.POST("/transactions/:id/process", h.AdminFinance.ProcessTransaction) // 处理交易
				finance.GET("/pending", h.AdminFinance.GetPendingTransactions)               // 待处理交易
				finance.GET("/type/:type", h.AdminFinance.GetTransactionsByType)             // 按类型查询
				finance.GET("/dashboard", h.AdminFinance.GetDashboardStats)                  // 仪表盘统计
				finance.GET("/export", h.AdminFinance.ExportTransactions)                    // 导出交易
				finance.POST("/batch-process", h.AdminFinance.BatchProcessTransactions)      // 批量处理
			}

			// 权限管理
			permissions := protected.Group("/permissions")
			{
				permissions.GET("/tree", h.AdminPermission.GetPermissionTree)  // 权限树
				permissions.GET("", h.AdminPermission.GetAllPermissions)       // 所有权限
				permissions.GET("/:id", h.AdminPermission.GetPermissionByID)   // 详情
				permissions.POST("", h.AdminPermission.CreatePermission)       // 创建
				permissions.PUT("/:id", h.AdminPermission.UpdatePermission)    // 更新
				permissions.DELETE("/:id", h.AdminPermission.DeletePermission) // 删除
			}

			// 统计分析
			statistics := protected.Group("/statistics")
			{
				statistics.GET("/overview", h.AdminStatistics.GetOverview) // 概览
				statistics.GET("/products", h.AdminStatistics.GetProducts) // 产品统计
			}

			// 系统管理
			system := protected.Group("/system")
			{
				system.GET("/configs", h.AdminSystem.GetConfigs)              // 获取配置
				system.PUT("/configs", h.AdminSystem.UpdateConfigs)           // 更新配置
				system.GET("/config/:key", h.AdminSystem.GetConfig)           // 获取单个配置
				system.PUT("/config/:key", h.AdminSystem.UpdateConfig)        // 更新单个配置
				system.GET("/stats", h.AdminSystem.GetStats)                  // 统计
				system.GET("/dashboard", h.AdminSystem.GetDashboard)          // 仪表盘
				system.GET("/info", h.AdminSystem.GetSystemInfo)              // 系统信息
				system.PUT("/info", h.AdminSystem.UpdateSystemInfo)           // 更新系统信息
				system.GET("/maintenance", h.AdminSystem.GetMaintenanceMode)  // 维护模式
				system.POST("/maintenance", h.AdminSystem.SetMaintenanceMode) // 设置维护模式
				system.GET("/health", h.AdminSystem.GetHealth)                // 健康检查
			}
		}
	}
}

// SetupClientRoutes 设置客户端路由
// 所有路由前缀: /api/cli
// 用于客户端应用，包括客户登录、个人资料、充值提现等
func SetupClientRoutes(api *gin.RouterGroup, h *Handlers) {
	cli := api.Group("/cli")
	{
		// 公开路由（不需要认证）
		auth := cli.Group("/auth")
		{
			auth.POST("/login", h.ClientAuth.Login) // 客户登录
		}

		// 受保护路由（需要认证）
		protected := cli.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// 认证相关
			authProtected := protected.Group("/auth")
			{
				authProtected.GET("/me", h.ClientAuth.GetProfile)           // 获取客户信息
				authProtected.POST("/logout", h.ClientAuth.Logout)          // 退出登录
				authProtected.PUT("/password", h.ClientAuth.UpdatePassword) // 修改密码
			}

			// 客户个人资料
			profile := protected.Group("/profile")
			{
				profile.GET("", h.ClientCustomer.GetProfile)    // 获取资料
				profile.PUT("", h.ClientCustomer.UpdateProfile) // 更新资料
			}

			// 财务管理（客户）
			finance := protected.Group("/finance")
			{
				finance.POST("/recharge", h.ClientFinance.Recharge)           // 充值
				finance.POST("/withdraw", h.ClientFinance.Withdraw)           // 提现
				finance.GET("/transactions", h.ClientFinance.GetTransactions) // 交易记录
				finance.GET("/balance", h.ClientFinance.GetBalance)           // 余额
				finance.GET("/statistics", h.ClientFinance.GetStatistics)     // 统计
			}

			// 优惠券（客户）
			coupons := protected.Group("/coupons")
			{
				coupons.GET("/my", h.ClientCoupon.GetUserCoupons)             // 我的优惠券
				coupons.POST("/claim", h.ClientCoupon.ClaimCoupon)            // 领取优惠券
				coupons.POST("/:id/use", h.ClientCoupon.UseCoupon)            // 使用优惠券
				coupons.GET("/available", h.ClientCoupon.GetAvailableCoupons) // 可用优惠券
			}

			// 授权码验证
			authcodes := protected.Group("/authcodes")
			{
				authcodes.POST("/verify", h.ClientAuthCode.Verify) // 验证授权码
			}
		}
	}
}
