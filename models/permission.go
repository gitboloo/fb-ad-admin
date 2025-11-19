package models

import (
	"gorm.io/gorm"
)

// Permission 统一的权限菜单模型 (合并了原有的Menu和Permission)
type Permission struct {
	gorm.Model

	// 基础信息
	Name  string `json:"name" gorm:"type:varchar(100);not null;index;comment:权限标识/菜单名称(英文)"`
	Code  string `json:"code" gorm:"type:varchar(100);uniqueIndex;not null;comment:权限代码(用于后端验证)"`
	Title string `json:"title" gorm:"type:varchar(100);not null;comment:显示标题(中文)"`
	Type  string `json:"type" gorm:"type:varchar(20);not null;default:'menu';index;comment:类型(menu=菜单/page=页面/button=按钮/api=API接口)"`

	// 树形结构
	ParentID uint `json:"parent_id" gorm:"default:0;index;comment:父权限ID,0表示顶级"`

	// 路由相关(菜单类型)
	Path      string `json:"path,omitempty" gorm:"type:varchar(200);comment:路由路径"`
	Component string `json:"component,omitempty" gorm:"type:varchar(200);comment:组件路径"`
	Redirect  string `json:"redirect,omitempty" gorm:"type:varchar(200);comment:重定向路径"`
	Icon      string `json:"icon,omitempty" gorm:"type:varchar(50);comment:图标"`

	// 显示控制
	Sort     int  `json:"sort" gorm:"default:0;comment:排序"`
	IsHidden bool `json:"is_hidden" gorm:"default:0;comment:是否隐藏(0=显示,1=隐藏)"`
	IsCache  bool `json:"is_cache" gorm:"default:1;comment:是否缓存(0=不缓存,1=缓存)"`
	IsAffix  bool `json:"is_affix" gorm:"default:0;comment:是否固定标签(0=不固定,1=固定)"`

	// API权限相关(api/button类型)
	APIPath   string `json:"api_path,omitempty" gorm:"type:varchar(200);comment:API路径"`
	APIMethod string `json:"api_method,omitempty" gorm:"type:varchar(10);comment:API方法(GET/POST/PUT/DELETE)"`

	// 状态和描述
	Status      int    `json:"status" gorm:"default:1;index;comment:状态(0=禁用,1=启用)"`
	Description string `json:"description,omitempty" gorm:"type:varchar(500);comment:描述/备注"`

	// 关联
	Roles    []Role       `json:"roles,omitempty" gorm:"many2many:role_permissions;"`
	Children []Permission `json:"children,omitempty" gorm:"-"` // 子节点,不存储到数据库
	Parent   *Permission  `json:"parent,omitempty" gorm:"-"`   // 父节点,不存储到数据库
}

// Role 角色模型
type Role struct {
	gorm.Model
	Name        string `json:"name" gorm:"type:varchar(100);uniqueIndex;not null;comment:角色名称"`
	Code        string `json:"code" gorm:"type:varchar(50);uniqueIndex;not null;comment:角色代码"`
	Title       string `json:"title" gorm:"type:varchar(100);comment:显示标题"`
	Description string `json:"description" gorm:"type:varchar(500);comment:描述"`
	Status      int    `json:"status" gorm:"default:1;comment:状态(0:禁用 1:启用)"`
	CreatorID   uint   `json:"creator_id" gorm:"index;comment:创建者ID(Admin的ID,0表示系统创建)"`

	// 关联
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	Admins      []Admin      `json:"admins" gorm:"many2many:admin_roles;"`
}

// AdminRoleAssoc 管理员角色关联表
type AdminRoleAssoc struct {
	AdminID uint `json:"admin_id" gorm:"primaryKey"`
	RoleID  uint `json:"role_id" gorm:"primaryKey"`
}

// RolePermission 角色权限关联表
type RolePermission struct {
	RoleID       uint `json:"role_id" gorm:"primaryKey"`
	PermissionID uint `json:"permission_id" gorm:"primaryKey"`
}

// 权限类型常量
const (
	PermissionTypeMenu   = "menu"   // 菜单(目录)
	PermissionTypePage   = "page"   // 页面
	PermissionTypeButton = "button" // 按钮
	PermissionTypeAPI    = "api"    // API接口
)

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// IsMenu 判断是否为菜单类型
func (p *Permission) IsMenu() bool {
	return p.Type == PermissionTypeMenu || p.Type == PermissionTypePage
}

// IsAction 判断是否为操作类型(按钮/API)
func (p *Permission) IsAction() bool {
	return p.Type == PermissionTypeButton || p.Type == PermissionTypeAPI
}

// GetPermissionsByRoleID 根据角色ID获取权限列表
func GetPermissionsByRoleID(db *gorm.DB, roleID uint) ([]Permission, error) {
	var role Role
	err := db.Preload("Permissions").First(&role, roleID).Error
	if err != nil {
		return nil, err
	}
	return role.Permissions, nil
}

// GetPermissionsByAdminID 根据管理员ID获取权限列表
func GetPermissionsByAdminID(db *gorm.DB, adminID uint) ([]Permission, error) {
	var admin Admin
	err := db.Preload("Roles.Permissions").First(&admin, adminID).Error
	if err != nil {
		return nil, err
	}

	// 合并所有角色的权限
	permMap := make(map[uint]Permission)
	for _, role := range admin.Roles {
		for _, perm := range role.Permissions {
			permMap[perm.ID] = perm
		}
	}

	// 转换为切片
	var permissions []Permission
	for _, perm := range permMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// HasPermission 检查管理员是否有指定权限
func (a *Admin) HasPermission(db *gorm.DB, permissionCode string) bool {
	permissions, err := GetPermissionsByAdminID(db, a.ID)
	if err != nil {
		return false
	}

	for _, perm := range permissions {
		if perm.Code == permissionCode && perm.Status == 1 {
			return true
		}
	}
	return false
}

// GetRoleByCode 根据角色代码获取角色
func GetRoleByCode(db *gorm.DB, code string) (*Role, error) {
	var role Role
	err := db.Where("code = ?", code).First(&role).Error
	return &role, err
}

// BuildPermissionTree 构建权限菜单树
func BuildPermissionTree(permissions []Permission) []Permission {
	// 创建ID到权限的映射
	permMap := make(map[uint]*Permission)
	for i := range permissions {
		permMap[permissions[i].ID] = &permissions[i]
	}

	// 构建树形结构
	var tree []Permission
	for i := range permissions {
		perm := &permissions[i]
		if perm.ParentID == 0 {
			// 顶级节点
			tree = append(tree, *perm)
		} else {
			// 子节点,添加到父节点的Children中
			if parent, exists := permMap[perm.ParentID]; exists {
				parent.Children = append(parent.Children, *perm)
			}
		}
	}

	return tree
}

// GetMenuTree 获取菜单树(只包含菜单和页面类型)
func GetMenuTree(db *gorm.DB, roleID uint) ([]Permission, error) {
	var permissions []Permission

	// 查询角色的所有菜单权限
	err := db.Table("permissions").
		Select("DISTINCT permissions.*").
		Joins("INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Where("permissions.type IN ?", []string{PermissionTypeMenu, PermissionTypePage}).
		Where("permissions.status = ?", 1).
		Order("permissions.sort ASC, permissions.id ASC").
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	return BuildPermissionTree(permissions), nil
}

// GetAllMenuTree 获取所有菜单树(用于管理员配置)
func GetAllMenuTree(db *gorm.DB) ([]Permission, error) {
	var permissions []Permission

	err := db.Where("type IN ?", []string{PermissionTypeMenu, PermissionTypePage}).
		Where("status = ?", 1).
		Order("sort ASC, id ASC").
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	return BuildPermissionTree(permissions), nil
}

// GetPermissionCodes 获取权限代码列表(用于前端按钮权限控制)
func GetPermissionCodes(db *gorm.DB, roleID uint) ([]string, error) {
	var codes []string

	err := db.Table("permissions").
		Select("permissions.code").
		Joins("INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Where("permissions.status = ?", 1).
		Pluck("code", &codes).Error

	return codes, err
}
