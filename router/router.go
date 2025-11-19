package router

import (
	"backend/controllers/admin"
	"backend/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handlers 所有handler的集合
type Handlers struct {
	// Admin handlers
	AdminAgent *admin.AgentHandler
	AdminRole  *admin.RoleHandler
}

// NewHandlers 创建所有handlers
func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{
		AdminAgent: admin.NewAgentHandler(db),
		AdminRole:  admin.NewRoleHandler(db),
	}
}

// SetupRouter 设置所有路由
func SetupRouter(r *gin.Engine, db *gorm.DB) {
	handlers := NewHandlers(db)

	// API路由组
	api := r.Group("/api")
	{
		// 管理后台路由
		SetupAdminRoutes(api, handlers)
	}
}

// SetupAdminRoutes 设置管理后台路由
func SetupAdminRoutes(api *gin.RouterGroup, h *Handlers) {
	admin := api.Group("/admin")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		// 代理商管理
		agents := admin.Group("/agents")
		{
			agents.GET("", h.AdminAgent.List)          // 列表
			agents.POST("", h.AdminAgent.Create)       // 创建
			agents.GET("/:id", h.AdminAgent.Detail)    // 详情
			agents.PUT("/:id", h.AdminAgent.Update)    // 更新
			agents.DELETE("/:id", h.AdminAgent.Delete) // 删除
		}

		// 角色管理
		roles := admin.Group("/roles")
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
	}
}
