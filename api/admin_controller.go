package api

import (
	"strconv"

	"backend/middleware"
	"backend/models"
	"backend/services"
	"backend/utils"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	adminService *services.AdminService
}

func NewAdminController() *AdminController {
	return &AdminController{
		adminService: services.NewAdminService(),
	}
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"remember,omitempty"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Admin *models.Admin `json:"admin"`
	Token string        `json:"token"`
}

// CreateAdminRequest 创建管理员请求结构
type CreateAdminRequest struct {
	Username string           `json:"username" binding:"required"`
	Account  string           `json:"account" binding:"required,email"`
	Password string           `json:"password" binding:"required"`
	Role     models.AdminRole `json:"role" binding:"required"`
}

// UpdateAdminRequest 更新管理员请求结构
type UpdateAdminRequest struct {
	Username string           `json:"username"`
	Account  string           `json:"account"`
	Role     models.AdminRole `json:"role"`
}

// UpdatePasswordRequest 更新密码请求结构
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UpdateStatusRequest 更新状态请求结构
type UpdateStatusRequest struct {
	Status models.AdminStatus `json:"status" binding:"required"`
}

// ResetPasswordRequest 重置密码请求结构
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

// Login 管理员登录
func (ctrl *AdminController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	admin, token, err := ctrl.adminService.Login(req.Username, req.Password)
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
// 这是 /api/auth/me 接口的实现
// 返回数据结构:
//
//	{
//	  "admin": {...},           // 管理员基本信息
//	  "roles": [...],           // 管理员的角色列表
//	  "menus": [...],           // 管理员可访问的菜单树（来自permissions表）
//	  "permissions": [...]      // 管理员的权限代码列表
//	}
func (ctrl *AdminController) GetProfile(c *gin.Context) {
	// 获取当前用户信息
	userID, _, role, exists := middleware.GetCurrentUser(c)
	if !exists {
		utils.Unauthorized(c, "用户信息不存在")
		return
	}

	// 获取管理员基本信息
	admin, err := ctrl.adminService.GetByID(userID)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	// 创建权限服务实例以获取菜单和权限
	// 注意: 现在所有菜单都来自统一的permissions表
	// 前后端通过type字段(menu/page/button/api)来区分
	permissionService := services.NewPermissionService()

	// 获取用户的菜单树(包含menu和page类型)
	// 这是基于用户的角色和role_permissions关联表
	menuTree, err := permissionService.GetUserMenuTree(userID)
	if err != nil {
		// 如果获取菜单失败，返回空菜单而不是错误
		// 这样前端至少能看到Dashboard等基础菜单
		menuTree = []models.Permission{}
	}

	// 转换菜单格式为前端期望的BackendMenuItem结构
	menus := convertMenusToResponse(menuTree)

	// 获取用户的所有权限代码列表
	// 用于前端按钮级别的权限控制
	permissions, err := permissionService.GetUserPermissions(userID)
	if err != nil {
		// 如果获取权限失败，返回空列表
		permissions = []string{}
	}

	// 构建用户的角色信息
	roles := []map[string]interface{}{
		{
			"id":   role,
			"name": getRoleName(role),
		},
	}

	// 构建并返回响应数据
	utils.Success(c, gin.H{
		"admin":       admin,
		"roles":       roles,
		"menus":       menus,
		"permissions": permissions,
	})
}

// Create 创建管理员
func (ctrl *AdminController) Create(c *gin.Context) {
	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	admin, err := ctrl.adminService.Create(req.Username, req.Account, req.Password, req.Role)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "创建成功", admin)
}

// Update 更新管理员信息
func (ctrl *AdminController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的管理员ID")
		return
	}

	var req UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	admin, err := ctrl.adminService.Update(uint(id), req.Username, req.Account, req.Role)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", admin)
}

// UpdatePassword 更新密码
func (ctrl *AdminController) UpdatePassword(c *gin.Context) {
	userID, _, _, exists := middleware.GetCurrentUser(c)
	if !exists {
		utils.Unauthorized(c, "用户信息不存在")
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err := ctrl.adminService.UpdatePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "密码更新成功", nil)
}

// UpdateStatus 更新管理员状态
func (ctrl *AdminController) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的管理员ID")
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err = ctrl.adminService.UpdateStatus(uint(id), req.Status)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "状态更新成功", nil)
}

// Delete 删除管理员
func (ctrl *AdminController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的管理员ID")
		return
	}

	err = ctrl.adminService.Delete(uint(id))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// List 获取管理员列表
func (ctrl *AdminController) List(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var status *models.AdminStatus
	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			statusVal := models.AdminStatus(s)
			status = &statusVal
		}
	}

	var role *models.AdminRole
	if roleStr := c.Query("role"); roleStr != "" {
		if r, err := strconv.Atoi(roleStr); err == nil {
			roleVal := models.AdminRole(r)
			role = &roleVal
		}
	}

	admins, total, err := ctrl.adminService.List(page, pageSize, status, role)
	if err != nil {
		utils.ServerError(c, "获取管理员列表失败")
		return
	}

	utils.PagedSuccess(c, admins, total, page, pageSize)
}

// GetByID 根据ID获取管理员
func (ctrl *AdminController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的管理员ID")
		return
	}

	admin, err := ctrl.adminService.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, admin)
}

// ResetPassword 重置密码（管理员功能）
func (ctrl *AdminController) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的管理员ID")
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err = ctrl.adminService.ResetPassword(uint(id), req.NewPassword)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "密码重置成功", nil)
}

// Logout 退出登录
func (ctrl *AdminController) Logout(c *gin.Context) {
	// 如果使用了Redis存储token，可以在这里将token加入黑名单
	// 这里简单返回成功，客户端清除token即可
	utils.SuccessWithMessage(c, "退出成功", nil)
}

// convertMenusToResponse 将数据库菜单模型转换为API响应格式（匹配前端 BackendMenuItem 结构）
func convertMenusToResponse(menuTree []models.Permission) []map[string]interface{} {
	var result []map[string]interface{}

	for _, menu := range menuTree {
		// 只处理菜单类型的权限
		if menu.Type == models.PermissionTypeMenu || menu.Type == models.PermissionTypePage {
			// 判断菜单类型: 1=目录(menu), 2=页面(page)
			menuType := 1
			if menu.Type == models.PermissionTypePage {
				menuType = 2
			}

			// 转换 is_hidden 布尔值为 0/1
			isHidden := 0
			if menu.IsHidden {
				isHidden = 1
			}

			// 转换 is_cache 布尔值为 0/1
			isCache := 1
			if !menu.IsCache {
				isCache = 0
			}

			// 转换 is_affix 布尔值为 0/1
			isAffix := 0
			if menu.IsAffix {
				isAffix = 1
			}

			item := map[string]interface{}{
				"id":        menu.ID,        // 数字ID
				"name":      menu.Code,      // 路由名称使用code（英文标识）
				"title":     menu.Title,     // 显示标题（中文）
				"path":      menu.Path,      // 路由路径
				"component": menu.Component, // 组件路径
				"redirect":  menu.Redirect,  // 重定向路径
				"icon":      menu.Icon,      // 图标
				"parent_id": menu.ParentID,  // 父级ID
				"is_hidden": isHidden,       // 是否隐藏: 0=显示, 1=隐藏
				"is_affix":  isAffix,        // 是否固定标签页
				"is_cache":  isCache,        // 是否缓存页面
				"type":      menuType,       // 类型: 1=目录(menu), 2=页面(page)
			}

			// 递归处理子菜单
			if len(menu.Children) > 0 {
				children := convertMenusToResponse(menu.Children)
				if len(children) > 0 {
					item["children"] = children
				}
			}

			result = append(result, item)
		}
	}

	return result
}

// getRoleName 根据角色ID获取角色名称
func getRoleName(role int) string {
	switch role {
	case 1:
		return "super_admin"
	case 2:
		return "admin"
	case 3:
		return "user"
	default:
		return "unknown"
	}
}
