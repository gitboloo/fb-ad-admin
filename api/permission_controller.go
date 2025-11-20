package api

import (
	"backend/middleware"
	"backend/models"
	"backend/services"
	"backend/utils"

	"github.com/gin-gonic/gin"
)

type PermissionController struct {
	permissionService *services.PermissionService
}

func NewPermissionController() *PermissionController {
	return &PermissionController{
		permissionService: services.NewPermissionService(),
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

// GetUserMenus 获取当前用户的菜单树（复用 /auth/me 的数据库逻辑）
func (ctrl *PermissionController) GetUserMenus(c *gin.Context) {
	userID, _, _, exists := middleware.GetCurrentUser(c)
	if !exists {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	// 直接复用 /auth/me 接口的菜单获取逻辑
	menus, err := utils.GetAdminMenus(userID)
	if err != nil {
		utils.ServerError(c, "获取菜单失败")
		return
	}

	// 构建菜单树
	menuTree := utils.BuildMenuTree(menus)

	utils.Success(c, menuTree)
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

// GetPermissionTree 获取权限树（管理员用）
func (ctrl *PermissionController) GetPermissionTree(c *gin.Context) {
	permissions, err := ctrl.permissionService.GetAllPermissions()
	if err != nil {
		utils.ServerError(c, "获取权限树失败")
		return
	}

	// 构建权限树结构
	tree := ctrl.buildPermissionTree(permissions, 0)
	utils.Success(c, tree)
}

// buildPermissionTree 构建权限树
func (ctrl *PermissionController) buildPermissionTree(permissions []models.Permission, parentID uint) []models.Permission {
	var tree []models.Permission
	for _, perm := range permissions {
		if perm.ParentID == parentID {
			children := ctrl.buildPermissionTree(permissions, perm.ID)
			if len(children) > 0 {
				perm.Children = children
			}
			tree = append(tree, perm)
		}
	}
	return tree
}

// GetPermissionByID 获取权限详情
func (ctrl *PermissionController) GetPermissionByID(c *gin.Context) {
	id := c.Param("id")

	permission, err := ctrl.permissionService.GetPermissionByID(id)
	if err != nil {
		utils.NotFound(c, "权限不存在")
		return
	}

	utils.Success(c, permission)
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

// 注意：所有硬编码的菜单配置已删除
// 系统现在统一使用数据库中的菜单数据，通过 /auth/me 接口返回
// 如果需要 GetUserMenus 接口，建议直接重定向到 /auth/me 或复用其逻辑
