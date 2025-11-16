package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminStatus int
type AdminRole int

const (
	AdminStatusInactive AdminStatus = 0
	AdminStatusActive   AdminStatus = 1
	AdminStatusLocked   AdminStatus = 2
)

const (
	AdminRoleSuperAdmin AdminRole = 1  // 超级管理员
	AdminRoleAdmin      AdminRole = 2  // 管理员
	AdminRoleUser       AdminRole = 3  // 普通用户
)

type Admin struct {
	ID           uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Account      string         `json:"account" gorm:"type:varchar(100);uniqueIndex;not null"`
	Password     string         `json:"-" gorm:"type:varchar(255);not null"`
	Role         AdminRole      `json:"role" gorm:"type:tinyint;not null;default:3"`
	Status       AdminStatus    `json:"status" gorm:"type:tinyint;not null;default:1"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty" gorm:"type:datetime;comment:最后登录时间"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	Roles []Role `json:"roles" gorm:"many2many:admin_roles;"`
}

func (Admin) TableName() string {
	return "admins"
}

// BeforeCreate 在创建前加密密码
func (a *Admin) BeforeCreate(tx *gorm.DB) error {
	if a.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		a.Password = string(hashedPassword)
	}
	return nil
}

// CheckPassword 验证密码
func (a *Admin) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}

// SetPassword 设置密码
func (a *Admin) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.Password = string(hashedPassword)
	return nil
}

// IsActive 检查账户是否激活
func (a *Admin) IsActive() bool {
	return a.Status == AdminStatusActive
}

// HasRole 检查是否有指定角色
func (a *Admin) HasRole(role AdminRole) bool {
	return a.Role <= role  // 角色值越小权限越高
}