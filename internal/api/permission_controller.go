package api

import (
	"fmt"
	
	"github.com/ad-platform/backend/internal/middleware"
	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/service"
	"github.com/ad-platform/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type PermissionController struct {
	permissionService *service.PermissionService
}

func NewPermissionController() *PermissionController {
	return &PermissionController{
		permissionService: service.NewPermissionService(),
	}
}

// GetUserPermissions 获取当前用户的权限列表
func (ctrl *PermissionController) GetUserPermissions(c *gin.Context) {
	userID, _, _, exists := middleware.GetCurrentUser(c)
	if !exists {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	permissions, err := ctrl.permissionService.GetUserPermissions(userID)
	if err != nil {
		utils.ServerError(c, "获取权限失败")
		return
	}

	utils.Success(c, permissions)
}

// GetUserMenus 获取当前用户的菜单树
func (ctrl *PermissionController) GetUserMenus(c *gin.Context) {
	userID, username, role, exists := middleware.GetCurrentUser(c)
	if !exists {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	// 调试信息
	fmt.Printf("GetUserMenus called for user: %s (ID: %d, Role: %d)\n", username, userID, role)
	
	// 根据角色返回默认菜单
	menus := ctrl.getMenusByRole(models.AdminRole(role))
	fmt.Printf("Returning %d menu items\n", len(menus))
	
	utils.Success(c, menus)
}

// GetAllPermissions 获取所有权限列表（管理员用）
func (ctrl *PermissionController) GetAllPermissions(c *gin.Context) {
	permissions, err := ctrl.permissionService.GetAllPermissions()
	if err != nil {
		utils.ServerError(c, "获取权限列表失败")
		return
	}

	utils.Success(c, permissions)
}

// CreatePermission 创建权限
func (ctrl *PermissionController) CreatePermission(c *gin.Context) {
	var req models.Permission
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	permission, err := ctrl.permissionService.CreatePermission(&req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, permission)
}

// UpdatePermission 更新权限
func (ctrl *PermissionController) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	
	var req models.Permission
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err := ctrl.permissionService.UpdatePermission(id, &req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// DeletePermission 删除权限
func (ctrl *PermissionController) DeletePermission(c *gin.Context) {
	id := c.Param("id")

	err := ctrl.permissionService.DeletePermission(id)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetAllRoles 获取所有角色
func (ctrl *PermissionController) GetAllRoles(c *gin.Context) {
	roles, err := ctrl.permissionService.GetAllRoles()
	if err != nil {
		utils.ServerError(c, "获取角色列表失败")
		return
	}

	utils.Success(c, roles)
}

// CreateRole 创建角色
func (ctrl *PermissionController) CreateRole(c *gin.Context) {
	var req models.Role
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	role, err := ctrl.permissionService.CreateRole(&req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, role)
}

// UpdateRole 更新角色
func (ctrl *PermissionController) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	
	var req models.Role
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err := ctrl.permissionService.UpdateRole(id, &req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// AssignPermissionsToRole 给角色分配权限
func (ctrl *PermissionController) AssignPermissionsToRole(c *gin.Context) {
	roleID := c.Param("id")
	
	var req struct {
		PermissionIDs []uint `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err := ctrl.permissionService.AssignPermissionsToRole(roleID, req.PermissionIDs)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// AssignRolesToUser 给用户分配角色
func (ctrl *PermissionController) AssignRolesToUser(c *gin.Context) {
	userID := c.Param("id")
	
	var req struct {
		RoleIDs []uint `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err := ctrl.permissionService.AssignRolesToUser(userID, req.RoleIDs)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// getMenusByRole 根据角色返回默认菜单
func (ctrl *PermissionController) getMenusByRole(role models.AdminRole) []map[string]interface{} {
	// 超级管理员的完整菜单
	if role == models.AdminRoleSuperAdmin {
		return []map[string]interface{}{
			{
				"id": "dashboard",
				"title": "仪表盘",
				"path": "/dashboard",
				"icon": "Dashboard",
			},
			{
				"id": "products",
				"title": "产品管理",
				"icon": "Goods",
				"children": []map[string]interface{}{
					{"id": "product-list", "title": "产品列表", "path": "/products/list"},
					{"id": "product-create", "title": "创建产品", "path": "/products/create"},
				},
			},
			{
				"id": "campaigns",
				"title": "计划管理",
				"icon": "Promotion",
				"children": []map[string]interface{}{
					{"id": "campaign-list", "title": "计划列表", "path": "/campaigns/list"},
					{"id": "campaign-create", "title": "创建计划", "path": "/campaigns/create"},
					{"id": "campaign-stats", "title": "计划统计", "path": "/campaigns/stats"},
				},
			},
			{
				"id": "coupons",
				"title": "优惠券管理",
				"icon": "Ticket",
				"children": []map[string]interface{}{
					{"id": "coupon-list", "title": "优惠券列表", "path": "/coupons/list"},
					{"id": "coupon-create", "title": "创建优惠券", "path": "/coupons/create"},
					{"id": "user-coupon-list", "title": "用户优惠券", "path": "/coupons/user-coupons"},
				},
			},
			{
				"id": "authcodes",
				"title": "授权码管理",
				"icon": "Key",
				"children": []map[string]interface{}{
					{"id": "authcode-list", "title": "授权码列表", "path": "/authcodes/list"},
					{"id": "authcode-generate", "title": "生成授权码", "path": "/authcodes/generate"},
				},
			},
			{
				"id": "finance",
				"title": "财务管理",
				"icon": "Money",
				"children": []map[string]interface{}{
					{"id": "transaction-list", "title": "交易记录", "path": "/finance/transactions"},
					{"id": "recharge-form", "title": "充值管理", "path": "/finance/recharge"},
					{"id": "withdraw-form", "title": "提现管理", "path": "/finance/withdraw"},
					{"id": "finance-stats", "title": "财务统计", "path": "/finance/stats"},
				},
			},
			{
				"id": "customers",
				"title": "客户管理",
				"icon": "User",
				"children": []map[string]interface{}{
					{"id": "customer-list", "title": "客户列表", "path": "/customers/list"},
				},
			},
			{
				"id": "system",
				"title": "系统管理",
				"icon": "Setting",
				"children": []map[string]interface{}{
					{"id": "system-config", "title": "系统配置", "path": "/system/config"},
					{"id": "admin-list", "title": "管理员管理", "path": "/system/admins"},
				},
			},
			{
				"id": "statistics",
				"title": "数据统计",
				"icon": "DataAnalysis",
				"children": []map[string]interface{}{
					{"id": "statistics-overview", "title": "总览统计", "path": "/statistics/overview"},
					{"id": "product-stats", "title": "产品统计", "path": "/statistics/products"},
					{"id": "revenue-stats", "title": "收入统计", "path": "/statistics/revenue"},
				},
			},
		}
	}
	
	// 管理员的菜单（没有系统管理和授权码管理）
	if role == models.AdminRoleAdmin {
		return []map[string]interface{}{
			{
				"id": "dashboard",
				"title": "仪表盘",
				"path": "/dashboard",
				"icon": "Dashboard",
			},
			{
				"id": "products",
				"title": "产品管理",
				"icon": "Goods",
				"children": []map[string]interface{}{
					{"id": "product-list", "title": "产品列表", "path": "/products/list"},
					{"id": "product-create", "title": "创建产品", "path": "/products/create"},
				},
			},
			{
				"id": "campaigns",
				"title": "计划管理",
				"icon": "Promotion",
				"children": []map[string]interface{}{
					{"id": "campaign-list", "title": "计划列表", "path": "/campaigns/list"},
					{"id": "campaign-create", "title": "创建计划", "path": "/campaigns/create"},
					{"id": "campaign-stats", "title": "计划统计", "path": "/campaigns/stats"},
				},
			},
			{
				"id": "coupons",
				"title": "优惠券管理",
				"icon": "Ticket",
				"children": []map[string]interface{}{
					{"id": "coupon-list", "title": "优惠券列表", "path": "/coupons/list"},
					{"id": "coupon-create", "title": "创建优惠券", "path": "/coupons/create"},
					{"id": "user-coupon-list", "title": "用户优惠券", "path": "/coupons/user-coupons"},
				},
			},
			{
				"id": "finance",
				"title": "财务管理",
				"icon": "Money",
				"children": []map[string]interface{}{
					{"id": "transaction-list", "title": "交易记录", "path": "/finance/transactions"},
					{"id": "recharge-form", "title": "充值管理", "path": "/finance/recharge"},
					{"id": "withdraw-form", "title": "提现管理", "path": "/finance/withdraw"},
					{"id": "finance-stats", "title": "财务统计", "path": "/finance/stats"},
				},
			},
			{
				"id": "customers",
				"title": "客户管理",
				"icon": "User",
				"children": []map[string]interface{}{
					{"id": "customer-list", "title": "客户列表", "path": "/customers/list"},
				},
			},
			{
				"id": "statistics",
				"title": "数据统计",
				"icon": "DataAnalysis",
				"children": []map[string]interface{}{
					{"id": "statistics-overview", "title": "总览统计", "path": "/statistics/overview"},
					{"id": "product-stats", "title": "产品统计", "path": "/statistics/products"},
					{"id": "revenue-stats", "title": "收入统计", "path": "/statistics/revenue"},
				},
			},
		}
	}
	
	// 普通用户只有查看权限的菜单
	return []map[string]interface{}{
		{
			"id": "dashboard",
			"title": "仪表盘",
			"path": "/dashboard",
			"icon": "Dashboard",
		},
		{
			"id": "products",
			"title": "产品管理",
			"icon": "Goods",
			"children": []map[string]interface{}{
				{"id": "product-list", "title": "产品列表", "path": "/products/list"},
			},
		},
		{
			"id": "campaigns",
			"title": "计划管理",
			"icon": "Promotion",
			"children": []map[string]interface{}{
				{"id": "campaign-list", "title": "计划列表", "path": "/campaigns/list"},
				{"id": "campaign-stats", "title": "计划统计", "path": "/campaigns/stats"},
			},
		},
		{
			"id": "statistics",
			"title": "数据统计",
			"icon": "DataAnalysis",
			"children": []map[string]interface{}{
				{"id": "statistics-overview", "title": "总览统计", "path": "/statistics/overview"},
				{"id": "product-stats", "title": "产品统计", "path": "/statistics/products"},
				{"id": "revenue-stats", "title": "收入统计", "path": "/statistics/revenue"},
			},
		},
	}
}