package repositories

import (
	"time"
	
	"backend/database"
	"backend/models"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository() *AdminRepository {
	return &AdminRepository{
		db: database.GetDB(),
	}
}

// Create 创建管理员
func (r *AdminRepository) Create(admin *models.Admin) error {
	return r.db.Create(admin).Error
}

// GetByID 根据ID获取管理员
func (r *AdminRepository) GetByID(id uint) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.First(&admin, id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetByUsername 根据用户名获取管理员
func (r *AdminRepository) GetByUsername(username string) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.Where("username = ?", username).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetByAccount 根据账号获取管理员
func (r *AdminRepository) GetByAccount(account string) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.Where("account = ?", account).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// Update 更新管理员信息
func (r *AdminRepository) Update(admin *models.Admin) error {
	return r.db.Save(admin).Error
}

// Delete 删除管理员（软删除）
func (r *AdminRepository) Delete(id uint) error {
	return r.db.Delete(&models.Admin{}, id).Error
}

// List 获取管理员列表
func (r *AdminRepository) List(page, pageSize int, status *models.AdminStatus, role *models.AdminRole) ([]*models.Admin, int64, error) {
	var admins []*models.Admin
	var total int64

	query := r.db.Model(&models.Admin{})
	
	// 添加过滤条件
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if role != nil {
		query = query.Where("role = ?", *role)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&admins).Error
	if err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

// UpdateStatus 更新管理员状态
func (r *AdminRepository) UpdateStatus(id uint, status models.AdminStatus) error {
	return r.db.Model(&models.Admin{}).Where("id = ?", id).Update("status", status).Error
}

// UpdatePassword 更新密码
func (r *AdminRepository) UpdatePassword(id uint, hashedPassword string) error {
	return r.db.Model(&models.Admin{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

// ExistsBy 检查记录是否存在
func (r *AdminRepository) ExistsBy(field, value string, excludeID ...uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Admin{}).Where(field+" = ?", value)
	
	// 排除指定ID（用于更新时检查重复）
	if len(excludeID) > 0 {
		query = query.Where("id != ?", excludeID[0])
	}
	
	err := query.Count(&count).Error
	return count > 0, err
}

// GetActiveAdmins 获取所有激活的管理员
func (r *AdminRepository) GetActiveAdmins() ([]*models.Admin, error) {
	var admins []*models.Admin
	err := r.db.Where("status = ?", models.AdminStatusActive).Find(&admins).Error
	return admins, err
}

// UpdateLastLogin 更新最后登录时间
func (r *AdminRepository) UpdateLastLogin(id uint) error {
	now := time.Now()
	return r.db.Model(&models.Admin{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_login_at": now,
		"updated_at": now,
	}).Error
}

// AssignRoles 给管理员分配角色
func (r *AdminRepository) AssignRoles(adminID uint, roleIDs []uint) error {
	var admin models.Admin
	if err := r.db.First(&admin, adminID).Error; err != nil {
		return err
	}

	// 清除现有角色关联
	if err := r.db.Model(&admin).Association("Roles").Clear(); err != nil {
		return err
	}

	// 添加新的角色关联
	if len(roleIDs) > 0 {
		var roles []models.Role
		if err := r.db.Find(&roles, roleIDs).Error; err != nil {
			return err
		}
		return r.db.Model(&admin).Association("Roles").Append(&roles)
	}

	return nil
}