package models

import (
	"gorm.io/gorm"
)

// Permission 权限模型
type Permission struct {
	gorm.Model
	Name        string `json:"name" gorm:"type:varchar(100);uniqueIndex;not null;comment:权限名称"`
	Code        string `json:"code" gorm:"type:varchar(100);uniqueIndex;not null;comment:权限代码"`
	Type        string `json:"type" gorm:"type:varchar(20);not null;comment:权限类型(menu/page/button/api)"`
	ParentID    uint   `json:"parent_id" gorm:"default:0;comment:父权限ID"`
	Path        string `json:"path" gorm:"type:varchar(200);comment:路由路径"`
	Component   string `json:"component" gorm:"type:varchar(200);comment:组件路径"`
	Icon        string `json:"icon" gorm:"type:varchar(50);comment:图标"`
	Sort        int    `json:"sort" gorm:"default:0;comment:排序"`
	Description string `json:"description" gorm:"type:varchar(500);comment:描述"`
	Status      int    `json:"status" gorm:"default:1;comment:状态(0:禁用 1:启用)"`

	// 关联
	Roles    []Role       `json:"roles" gorm:"many2many:role_permissions;"`
	Children []Permission `json:"children" gorm:"foreignKey:ParentID"`
}

// Role 角色模型
type Role struct {
	gorm.Model
	Name        string `json:"name" gorm:"type:varchar(100);uniqueIndex;not null;comment:角色名称"`
	Code        string `json:"code" gorm:"type:varchar(50);uniqueIndex;not null;comment:角色代码"`
	Description string `json:"description" gorm:"type:varchar(500);comment:描述"`
	Status      int    `json:"status" gorm:"default:1;comment:状态(0:禁用 1:启用)"`

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
	PermissionTypeMenu   = "menu"   // 菜单
	PermissionTypePage   = "page"   // 页面
	PermissionTypeButton = "button" // 按钮
	PermissionTypeAPI    = "api"    // API接口
)

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