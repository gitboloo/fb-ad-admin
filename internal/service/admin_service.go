package service

import (
	"errors"
	"fmt"

	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/repository"
	"github.com/ad-platform/backend/internal/utils"
	"gorm.io/gorm"
)

type AdminService struct {
	adminRepo *repository.AdminRepository
}

func NewAdminService() *AdminService {
	return &AdminService{
		adminRepo: repository.NewAdminRepository(),
	}
}

// Login 管理员登录
func (s *AdminService) Login(username, password string) (*models.Admin, string, error) {
	// 参数验证
	if username == "" || password == "" {
		return nil, "", errors.New("用户名和密码不能为空")
	}

	// 通过用户名查找管理员
	admin, err := s.adminRepo.GetByUsername(username)
	if err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 检查账户状态
	if !admin.IsActive() {
		return nil, "", errors.New("账户已被禁用")
	}

	// 验证密码
	if !admin.CheckPassword(password) {
		return nil, "", errors.New("账号或密码错误")
	}

	// 生成JWT token
	token, err := utils.GenerateToken(admin.ID, admin.Username, int(admin.Role))
	if err != nil {
		return nil, "", errors.New("生成认证令牌失败")
	}

	// 更新最后登录时间
	s.adminRepo.UpdateLastLogin(admin.ID)

	return admin, token, nil
}

// Create 创建管理员
func (s *AdminService) Create(username, account, password string, role models.AdminRole) (*models.Admin, error) {
	// 参数验证
	if valid, msg := utils.ValidateUsername(username); !valid {
		return nil, errors.New(msg)
	}
	if !utils.ValidateEmail(account) {
		return nil, errors.New("邮箱格式不正确")
	}
	if valid, msg := utils.ValidatePassword(password); !valid {
		return nil, errors.New(msg)
	}

	// 检查用户名是否已存在
	exists, err := s.adminRepo.ExistsBy("username", username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查账号是否已存在
	exists, err = s.adminRepo.ExistsBy("account", account)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("账号已存在")
	}

	// 创建管理员
	admin := &models.Admin{
		Username: username,
		Account:  account,
		Password: password, // 在模型的BeforeCreate中会自动加密
		Role:     role,
		Status:   models.AdminStatusActive,
	}

	err = s.adminRepo.Create(admin)
	if err != nil {
		return nil, fmt.Errorf("创建管理员失败: %w", err)
	}

	return admin, nil
}

// GetByID 根据ID获取管理员
func (s *AdminService) GetByID(id uint) (*models.Admin, error) {
	admin, err := s.adminRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("管理员不存在")
		}
		return nil, err
	}
	return admin, nil
}

// Update 更新管理员信息
func (s *AdminService) Update(id uint, username, account string, role models.AdminRole) (*models.Admin, error) {
	// 获取现有管理员
	admin, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 参数验证
	if username != "" {
		if valid, msg := utils.ValidateUsername(username); !valid {
			return nil, errors.New(msg)
		}
		// 检查用户名是否已被其他用户使用
		exists, err := s.adminRepo.ExistsBy("username", username, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("用户名已存在")
		}
		admin.Username = username
	}

	if account != "" {
		if !utils.ValidateEmail(account) {
			return nil, errors.New("邮箱格式不正确")
		}
		// 检查账号是否已被其他用户使用
		exists, err := s.adminRepo.ExistsBy("account", account, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("账号已存在")
		}
		admin.Account = account
	}

	admin.Role = role

	err = s.adminRepo.Update(admin)
	if err != nil {
		return nil, fmt.Errorf("更新管理员失败: %w", err)
	}

	return admin, nil
}

// UpdatePassword 更新密码
func (s *AdminService) UpdatePassword(id uint, oldPassword, newPassword string) error {
	admin, err := s.GetByID(id)
	if err != nil {
		return err
	}

	// 验证旧密码
	if !admin.CheckPassword(oldPassword) {
		return errors.New("原密码错误")
	}

	// 验证新密码
	if valid, msg := utils.ValidatePassword(newPassword); !valid {
		return errors.New(msg)
	}

	// 设置新密码
	err = admin.SetPassword(newPassword)
	if err != nil {
		return err
	}

	err = s.adminRepo.Update(admin)
	if err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// UpdateStatus 更新管理员状态
func (s *AdminService) UpdateStatus(id uint, status models.AdminStatus) error {
	// 检查管理员是否存在
	_, err := s.GetByID(id)
	if err != nil {
		return err
	}

	err = s.adminRepo.UpdateStatus(id, status)
	if err != nil {
		return fmt.Errorf("更新状态失败: %w", err)
	}

	return nil
}

// Delete 删除管理员
func (s *AdminService) Delete(id uint) error {
	// 检查管理员是否存在
	_, err := s.GetByID(id)
	if err != nil {
		return err
	}

	err = s.adminRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("删除管理员失败: %w", err)
	}

	return nil
}

// List 获取管理员列表
func (s *AdminService) List(page, pageSize int, status *models.AdminStatus, role *models.AdminRole) ([]*models.Admin, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return s.adminRepo.List(page, pageSize, status, role)
}

// ResetPassword 重置密码（管理员功能）
func (s *AdminService) ResetPassword(id uint, newPassword string) error {
	// 验证新密码
	if valid, msg := utils.ValidatePassword(newPassword); !valid {
		return errors.New(msg)
	}

	// 获取管理员
	admin, err := s.GetByID(id)
	if err != nil {
		return err
	}

	// 设置新密码
	err = admin.SetPassword(newPassword)
	if err != nil {
		return err
	}

	err = s.adminRepo.Update(admin)
	if err != nil {
		return fmt.Errorf("重置密码失败: %w", err)
	}

	return nil
}