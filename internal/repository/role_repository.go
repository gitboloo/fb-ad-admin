package repository

import (
	"github.com/ad-platform/backend/internal/database"
	"github.com/ad-platform/backend/internal/models"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{
		db: database.GetDB(),
	}
}

// GetAll 获取所有角色
func (r *RoleRepository) GetAll() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

// GetByID 根据ID获取角色
func (r *RoleRepository) GetByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	return &role, err
}

// GetByCode 根据代码获取角色
func (r *RoleRepository) GetByCode(code string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("code = ?", code).First(&role).Error
	return &role, err
}

// ExistsByCode 检查角色代码是否存在
func (r *RoleRepository) ExistsByCode(code string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Role{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

// Create 创建角色
func (r *RoleRepository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

// Update 更新角色
func (r *RoleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

// Delete 删除角色
func (r *RoleRepository) Delete(id uint) error {
	return r.db.Delete(&models.Role{}, id).Error
}

// AssignPermissions 给角色分配权限
func (r *RoleRepository) AssignPermissions(roleID uint, permissionIDs []uint) error {
	var role models.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	// 清除现有权限关联
	if err := r.db.Model(&role).Association("Permissions").Clear(); err != nil {
		return err
	}

	// 添加新的权限关联
	if len(permissionIDs) > 0 {
		var permissions []models.Permission
		if err := r.db.Find(&permissions, permissionIDs).Error; err != nil {
			return err
		}
		return r.db.Model(&role).Association("Permissions").Append(&permissions)
	}

	return nil
}