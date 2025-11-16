package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

type AuthCodeStatus int

const (
	AuthCodeStatusUnused  AuthCodeStatus = 1
	AuthCodeStatusUsed    AuthCodeStatus = 2
	AuthCodeStatusExpired AuthCodeStatus = 3
)

type AuthCode struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Code      string         `json:"code" gorm:"type:varchar(64);uniqueIndex;not null"`
	Status    AuthCodeStatus `json:"status" gorm:"type:tinyint;not null;default:1"`
	UsedBy    *uint          `json:"used_by" gorm:"index"`           // 使用者用户ID
	UsedAt    *time.Time     `json:"used_at"`                       // 使用时间
	ExpiredAt time.Time      `json:"expired_at"`                    // 过期时间
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (AuthCode) TableName() string {
	return "auth_codes"
}

// GenerateCode 生成授权码
func GenerateAuthCode() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为种子
		return hex.EncodeToString([]byte(time.Now().String()))[:32]
	}
	return hex.EncodeToString(bytes)
}

// BeforeCreate 在创建前生成授权码
func (ac *AuthCode) BeforeCreate(tx *gorm.DB) error {
	if ac.Code == "" {
		ac.Code = GenerateAuthCode()
	}
	if ac.ExpiredAt.IsZero() {
		// 默认7天有效期
		ac.ExpiredAt = time.Now().AddDate(0, 0, 7)
	}
	return nil
}

// IsUsable 检查授权码是否可用
func (ac *AuthCode) IsUsable() bool {
	if ac.Status != AuthCodeStatusUnused {
		return false
	}
	
	// 检查是否过期
	now := time.Now()
	if now.After(ac.ExpiredAt) {
		return false
	}
	
	return true
}

// Use 使用授权码
func (ac *AuthCode) Use(userID uint) {
	now := time.Now()
	ac.Status = AuthCodeStatusUsed
	ac.UsedBy = &userID
	ac.UsedAt = &now
}

// Expire 使授权码过期
func (ac *AuthCode) Expire() {
	ac.Status = AuthCodeStatusExpired
}

// GetRemainingTime 获取剩余有效时间
func (ac *AuthCode) GetRemainingTime() time.Duration {
	if ac.ExpiredAt.Before(time.Now()) {
		return 0
	}
	return ac.ExpiredAt.Sub(time.Now())
}