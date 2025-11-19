package admin

import (
	"fmt"
	"strconv"

	"backend/models"
	"backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RoleHandler 管理员-角色管理
type RoleHandler struct {
	db *gorm.DB
}

// NewRoleHandler 创建角色管理handler
func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required"`
	Code          string `json:"code" binding:"required"`
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description"`
	ParentRoleID  *uint  `json:"parent_role_id"` // 上级角色(用于权限继承验证)
	PermissionIDs []uint `json:"permission_ids"` // 权限ID列表
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name          string `json:"name"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Status        int    `json:"status"`
	PermissionIDs []uint `json:"permission_ids"`
}

// List 获取角色列表
// GET /api/admin/roles
func (h *RoleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	offset := (page - 1) * pageSize
	query := h.db.Model(&models.Role{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var roles []models.Role
	var total int64
	query.Count(&total)
	query.Preload("Permissions").Offset(offset).Limit(pageSize).Find(&roles)

	utils.Success(c, gin.H{
		"list":  roles,
		"total": total,
	})
}

// Create 创建角色
// POST /api/admin/roles
func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 检查角色code是否已存在
	var existingRole models.Role
	if err := h.db.Where("code = ?", req.Code).First(&existingRole).Error; err == nil {
		utils.BadRequest(c, "角色代码已存在")
		return
	} else if err != gorm.ErrRecordNotFound {
		utils.ServerError(c, "查询失败")
		return
	}

	// 如果有ParentRoleID，验证权限继承
	if req.ParentRoleID != nil {
		if err := h.validatePermissionInheritance(c, *req.ParentRoleID, req.PermissionIDs); err != nil {
			utils.BadRequest(c, err.Error())
			return
		}
	}

	// 获取创建者ID（当前登录用户）
	creatorID := uint(0)
	if adminIDVal, exists := c.Get("admin_id"); exists {
		if id, ok := adminIDVal.(float64); ok {
			creatorID = uint(id)
		}
	}

	// 验证当前用户是否有权分配这些权限（权限必须在用户权限范围内）
	if creatorID > 0 && len(req.PermissionIDs) > 0 {
		if err := h.validateUserCanAssignPermissions(c, creatorID, req.PermissionIDs); err != nil {
			utils.Forbidden(c, err.Error())
			return
		}
	}

	// 创建角色
	role := models.Role{
		Name:        req.Name,
		Code:        req.Code,
		Title:       req.Title,
		Description: req.Description,
		Status:      1,
		CreatorID:   creatorID,
	}

	if err := h.db.Create(&role).Error; err != nil {
		utils.ServerError(c, "创建失败")
		return
	}

	// 分配权限
	if len(req.PermissionIDs) > 0 {
		permissions := make([]models.Permission, 0)
		if err := h.db.Where("id IN ?", req.PermissionIDs).Find(&permissions).Error; err != nil {
			utils.ServerError(c, "查询权限失败")
			return
		}

		if err := h.db.Model(&role).Association("Permissions").Append(permissions); err != nil {
			utils.ServerError(c, "分配权限失败")
			return
		}
	}

	// 重新加载权限
	h.db.Preload("Permissions").First(&role)

	utils.Success(c, gin.H{"data": role})
}

// Detail 获取角色详情
// GET /api/admin/roles/:id
func (h *RoleHandler) Detail(c *gin.Context) {
	id := c.Param("id")

	var role models.Role
	if err := h.db.Preload("Permissions").First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "角色不存在")
		} else {
			utils.ServerError(c, "查询失败")
		}
		return
	}

	utils.Success(c, gin.H{"data": role})
}

// Update 更新角色
// PUT /api/admin/roles/:id
func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	var role models.Role
	if err := h.db.First(&role, id).Error; err != nil {
		utils.NotFound(c, "角色不存在")
		return
	}

	// 构建更新map
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status > 0 {
		updates["status"] = req.Status
	}

	if err := h.db.Model(&role).Updates(updates).Error; err != nil {
		utils.ServerError(c, "更新失败")
		return
	}

	// 更新权限
	if len(req.PermissionIDs) > 0 {
		// 验证当前用户是否有权分配这些权限
		adminIDVal, exists := c.Get("admin_id")
		if exists {
			if id, ok := adminIDVal.(float64); ok {
				if err := h.validateUserCanAssignPermissions(c, uint(id), req.PermissionIDs); err != nil {
					utils.Forbidden(c, err.Error())
					return
				}
			}
		}

		permissions := make([]models.Permission, 0)
		if err := h.db.Where("id IN ?", req.PermissionIDs).Find(&permissions).Error; err != nil {
			utils.ServerError(c, "查询权限失败")
			return
		}

		// 清除旧权限并添加新权限
		if err := h.db.Model(&role).Association("Permissions").Replace(permissions); err != nil {
			utils.ServerError(c, "更新权限失败")
			return
		}
	}

	// 重新加载
	h.db.Preload("Permissions").First(&role)

	utils.Success(c, gin.H{"data": role})
}

// AssignPermissions 分配权限给角色
// POST /api/admin/roles/:id/permissions
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		PermissionIDs []uint `json:"permission_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	var role models.Role
	if err := h.db.First(&role, id).Error; err != nil {
		utils.NotFound(c, "角色不存在")
		return
	}

	// 验证当前用户是否有权分配这些权限
	adminIDVal, exists := c.Get("admin_id")
	if exists {
		if adminID, ok := adminIDVal.(float64); ok {
			if err := h.validateUserCanAssignPermissions(c, uint(adminID), req.PermissionIDs); err != nil {
				utils.Forbidden(c, err.Error())
				return
			}
		}
	}

	// 查询权限
	permissions := make([]models.Permission, 0)
	if err := h.db.Where("id IN ?", req.PermissionIDs).Find(&permissions).Error; err != nil {
		utils.ServerError(c, "查询权限失败")
		return
	}

	// 替换权限
	if err := h.db.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		utils.ServerError(c, "分配权限失败")
		return
	}

	// 重新加载
	h.db.Preload("Permissions").First(&role)

	utils.Success(c, gin.H{"data": role})
}

// Delete 删除角色
// DELETE /api/admin/roles/:id
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	// 检查是否有管理员使用该角色
	var adminCount int64
	h.db.Model(&models.Admin{}).Joins("JOIN admin_roles ON admin_roles.admin_id = admins.id").
		Where("admin_roles.role_id = ?", id).Count(&adminCount)

	if adminCount > 0 {
		utils.BadRequest(c, "角色已被管理员使用，无法删除")
		return
	}

	// 删除角色关联的权限
	if err := h.db.Model(&models.Role{}).Where("id = ?", id).Association("Permissions").Clear(); err != nil {
		utils.ServerError(c, "删除权限关联失败")
		return
	}

	// 删除角色
	if err := h.db.Delete(&models.Role{}, id).Error; err != nil {
		utils.ServerError(c, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// GetAssignableRoles 获取当前用户可分配的角色列表
// GET /api/admin/roles/assignable
// 返回：当前用户拥有的角色下属角色（排除自己拥有的角色）
func (h *RoleHandler) GetAssignableRoles(c *gin.Context) {
	// 获取当前登录用户ID
	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	currentAdminID := uint(0)
	if id, ok := adminIDVal.(float64); ok {
		currentAdminID = uint(id)
	}

	if currentAdminID == 0 {
		utils.BadRequest(c, "获取用户信息失败")
		return
	}

	// 获取当前用户的所有角色
	var admin models.Admin
	if err := h.db.Preload("Roles").Preload("Roles.Permissions").First(&admin, currentAdminID).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	// 如果用户没有角色，返回空列表
	if len(admin.Roles) == 0 {
		utils.Success(c, gin.H{
			"data": []models.Role{},
		})
		return
	}

	// 获取当前用户拥有的所有权限ID
	userPermissionIDs := make(map[uint]bool)
	for _, role := range admin.Roles {
		for _, perm := range role.Permissions {
			userPermissionIDs[perm.ID] = true
		}
	}

	// 获取当前用户拥有的所有角色ID
	userRoleIDs := make(map[uint]bool)
	for _, role := range admin.Roles {
		userRoleIDs[role.ID] = true
	}

	// 查询所有角色及其权限
	var allRoles []models.Role
	if err := h.db.Preload("Permissions").Find(&allRoles).Error; err != nil {
		utils.ServerError(c, "查询角色失败")
		return
	}

	// 过滤规则：
	// 1. 排除当前用户拥有的角色
	// 2. 排除权限超出当前用户权限范围的角色
	var assignableRoles []models.Role
	for _, role := range allRoles {
		// 规则1：排除自己拥有的角色
		if userRoleIDs[role.ID] {
			continue
		}

		// 规则2：检查角色的所有权限是否都在当前用户的权限范围内
		isAssignable := true
		for _, perm := range role.Permissions {
			if !userPermissionIDs[perm.ID] {
				isAssignable = false
				break
			}
		}

		if isAssignable {
			assignableRoles = append(assignableRoles, role)
		}
	}

	utils.Success(c, gin.H{
		"data": assignableRoles,
	})
}

// GetPermissions 获取当前用户可赋予的权限树
// GET /api/admin/roles/permissions/tree
// 返回：当前用户拥有的权限树（用于创建/编辑角色时选择）
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	// 获取当前登录用户ID
	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	currentAdminID := uint(0)
	if id, ok := adminIDVal.(float64); ok {
		currentAdminID = uint(id)
	}

	if currentAdminID == 0 {
		utils.BadRequest(c, "获取用户信息失败")
		return
	}

	// 获取当前用户的所有角色及其权限
	var admin models.Admin
	if err := h.db.Preload("Roles").Preload("Roles.Permissions").First(&admin, currentAdminID).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	// 收集当前用户拥有的所有权限
	userPermissions := make(map[uint]*models.Permission)
	for _, role := range admin.Roles {
		for _, perm := range role.Permissions {
			userPermissions[perm.ID] = &perm
		}
	}

	// 如果用户没有任何权限，返回空树
	if len(userPermissions) == 0 {
		utils.Success(c, gin.H{"data": []models.Permission{}})
		return
	}

	// 查询所有权限
	var allPermissions []models.Permission
	if err := h.db.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&allPermissions).Error; err != nil {
		utils.ServerError(c, "查询权限失败")
		return
	}

	// 过滤出用户拥有的权限
	var userOwnedPerms []models.Permission
	for _, perm := range allPermissions {
		if userPermissions[perm.ID] != nil {
			userOwnedPerms = append(userOwnedPerms, perm)
		}
	}

	// 构建树形结构（只包含用户拥有的权限）
	tree := models.BuildPermissionTree(userOwnedPerms)

	utils.Success(c, gin.H{"data": tree})
}

// validatePermissionInheritance 验证权限继承（下级权限必须是上级的子集）
func (h *RoleHandler) validatePermissionInheritance(c *gin.Context, parentRoleID uint, permissionIDs []uint) error {
	// 获取上级角色的权限
	var parentRole models.Role
	if err := h.db.Preload("Permissions").First(&parentRole, parentRoleID).Error; err != nil {
		return err
	}

	// 创建上级权限ID的集合
	parentPermMap := make(map[uint]bool)
	for _, perm := range parentRole.Permissions {
		parentPermMap[perm.ID] = true
	}

	// 检查新权限是否都在上级权限范围内
	for _, permID := range permissionIDs {
		if !parentPermMap[permID] {
			return fmt.Errorf("新权限超出上级角色权限范围")
		}
	}

	return nil
}

// validateUserCanAssignPermissions 验证当前用户是否有权分配这些权限
// 规则：用户只能分配自己拥有的权限
func (h *RoleHandler) validateUserCanAssignPermissions(c *gin.Context, adminID uint, permissionIDs []uint) error {
	// 获取当前用户的所有角色及权限
	var admin models.Admin
	if err := h.db.Preload("Roles").Preload("Roles.Permissions").First(&admin, adminID).Error; err != nil {
		return fmt.Errorf("获取用户信息失败")
	}

	// 收集用户拥有的权限ID
	userPermMap := make(map[uint]bool)
	for _, role := range admin.Roles {
		for _, perm := range role.Permissions {
			userPermMap[perm.ID] = true
		}
	}

	// 检查所有要分配的权限是否都在用户的权限范围内
	for _, permID := range permissionIDs {
		if !userPermMap[permID] {
			return fmt.Errorf("您没有权限分配该权限")
		}
	}

	return nil
}
