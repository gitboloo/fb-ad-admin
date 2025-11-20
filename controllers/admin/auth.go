package admin

import (
	"backend/middleware"
	"backend/models"
	"backend/services"
	"backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证管理
type AuthHandler struct {
	adminService      *services.AdminService
	permissionService *services.PermissionService
}

// NewAuthHandler 创建认证管理handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		adminService:      services.NewAdminService(),
		permissionService: services.NewPermissionService(),
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"remember,omitempty"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Admin *models.Admin `json:"admin"`
	Token string        `json:"token"`
}

// Login 管理员登录
// POST /api/admin/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	admin, token, err := h.adminService.Login(req.Account, req.Password)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, LoginResponse{
		Admin: admin,
		Token: token,
	})
}

// GetProfile 获取当前管理员信息(包含角色、菜单、权限)
// GET /api/admin/auth/me
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// 获取当前用户信息
	userID, _, role, exists := middleware.GetCurrentAdmin(c)
	if !exists {
		utils.Unauthorized(c, "用户信息不存在")
		return
	}

	// 获取管理员基本信息
	admin, err := h.adminService.GetByID(userID)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	// 获取用户的菜单树
	menuTree, err := h.permissionService.GetUserMenuTree(userID)
	if err != nil {
		menuTree = []models.Permission{}
	}

	// 转换菜单格式
	menus := h.convertMenusToResponse(menuTree)

	// 获取用户的所有权限代码列表
	permissions, err := h.permissionService.GetUserPermissions(userID)
	if err != nil {
		permissions = []string{}
	}

	// 构建角色信息
	roles := []map[string]interface{}{
		{
			"id":   role,
			"name": h.getRoleName(models.AdminRole(role)),
		},
	}

	utils.Success(c, gin.H{
		"admin":       admin,
		"roles":       roles,
		"menus":       menus,
		"permissions": permissions,
	})
}

// Logout 退出登录
// POST /api/admin/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: 实现token失效逻辑（如果使用Redis存储token）
	utils.Success(c, gin.H{
		"message": "退出成功",
	})
}

// GetPermissions 获取用户权限列表
// GET /api/admin/auth/permissions
func (h *AuthHandler) GetPermissions(c *gin.Context) {
	userID, _, _, exists := middleware.GetCurrentAdmin(c)
	if !exists {
		utils.Unauthorized(c, "用户信息不存在")
		return
	}

	permissions, err := h.permissionService.GetUserPermissions(userID)
	if err != nil {
		utils.ServerError(c, "获取权限失败")
		return
	}

	utils.Success(c, gin.H{
		"permissions": permissions,
	})
}

// GetMenus 获取用户菜单树
// GET /api/admin/auth/menus
func (h *AuthHandler) GetMenus(c *gin.Context) {
	userID, _, _, exists := middleware.GetCurrentAdmin(c)
	if !exists {
		utils.Unauthorized(c, "用户信息不存在")
		return
	}

	menuTree, err := h.permissionService.GetUserMenuTree(userID)
	if err != nil {
		utils.ServerError(c, "获取菜单失败")
		return
	}

	menus := h.convertMenusToResponse(menuTree)
	utils.Success(c, gin.H{
		"menus": menus,
	})
}

// UpdatePasswordRequest 更新密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UpdatePassword 修改密码
// PUT /api/admin/auth/password
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	userID, _, _, exists := middleware.GetCurrentAdmin(c)
	if !exists {
		utils.Unauthorized(c, "用户信息不存在")
		return
	}

	if err := h.adminService.UpdatePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message": "密码修改成功",
	})
}

// convertMenusToResponse 转换菜单格式
func (h *AuthHandler) convertMenusToResponse(menus []models.Permission) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, menu := range menus {
		item := map[string]interface{}{
			"id":        menu.ID,
			"name":      menu.Name,
			"title":     menu.Title, // 添加 title 字段用于菜单显示
			"path":      menu.Path,
			"component": menu.Component,
			"icon":      menu.Icon,
			"sort":      menu.Sort,
			"hidden":    menu.IsHidden,
		}

		if len(menu.Children) > 0 {
			item["children"] = h.convertMenusToResponse(menu.Children)
		}

		result = append(result, item)
	}
	return result
}

// getRoleName 获取角色名称
func (h *AuthHandler) getRoleName(role models.AdminRole) string {
	switch role {
	case models.AdminRoleSuperAdmin:
		return "超级管理员"
	case models.AdminRoleAdmin:
		return "管理员"
	case models.AdminRoleUser:
		return "普通用户"
	default:
		return "未知角色"
	}
}
