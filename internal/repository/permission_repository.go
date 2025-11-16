package repository

import (
	"github.com/ad-platform/backend/internal/database"
	"github.com/ad-platform/backend/internal/models"
	"gorm.io/gorm"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository() *PermissionRepository {
	return &PermissionRepository{
		db: database.GetDB(),
	}
}

// GetAll 获取所有权限
func (r *PermissionRepository) GetAll() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}

// GetByID 根据ID获取权限
func (r *PermissionRepository) GetByID(id uint) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.First(&permission, id).Error
	return &permission, err
}

// GetByCode 根据代码获取权限
func (r *PermissionRepository) GetByCode(code string) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.Where("code = ?", code).First(&permission).Error
	return &permission, err
}

// ExistsByCode 检查权限代码是否存在
func (r *PermissionRepository) ExistsByCode(code string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Permission{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

// Create 创建权限
func (r *PermissionRepository) Create(permission *models.Permission) error {
	return r.db.Create(permission).Error
}

// Update 更新权限
func (r *PermissionRepository) Update(permission *models.Permission) error {
	return r.db.Save(permission).Error
}

// Delete 删除权限
func (r *PermissionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Permission{}, id).Error
}

// GetPermissionsByRoleID 根据角色ID获取权限列表
func (r *PermissionRepository) GetPermissionsByRoleID(roleID uint) ([]models.Permission, error) {
	var role models.Role
	err := r.db.Preload("Permissions").First(&role, roleID).Error
	if err != nil {
		return nil, err
	}
	return role.Permissions, nil
}

// GetPermissionsByAdminID 根据管理员ID获取权限列表
func (r *PermissionRepository) GetPermissionsByAdminID(adminID uint) ([]models.Permission, error) {
	var admin models.Admin
	err := r.db.Preload("Roles.Permissions").First(&admin, adminID).Error
	if err != nil {
		return nil, err
	}

	// 合并所有角色的权限
	permMap := make(map[uint]models.Permission)
	for _, role := range admin.Roles {
		for _, perm := range role.Permissions {
			if perm.Status == 1 {
				permMap[perm.ID] = perm
			}
		}
	}

	// 转换为切片
	var permissions []models.Permission
	for _, perm := range permMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}