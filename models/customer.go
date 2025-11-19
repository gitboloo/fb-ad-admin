package models

import (
	"time"

	"gorm.io/gorm"
)

type CustomerStatus int

const (
	CustomerStatusInactive CustomerStatus = 0
	CustomerStatusActive   CustomerStatus = 1
	CustomerStatusBlocked  CustomerStatus = 2
)

type Customer struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string         `json:"name" gorm:"type:varchar(255);not null"`
	Email     string         `json:"email" gorm:"type:varchar(255);uniqueIndex"`
	Phone     string         `json:"phone" gorm:"type:varchar(20);index"`
	Company   string         `json:"company" gorm:"type:varchar(255)"`
	Status    CustomerStatus `json:"status" gorm:"type:tinyint;not null;default:1"`
	Avatar    string         `json:"avatar" gorm:"type:varchar(500)"` // 头像URL
	Address   string         `json:"address" gorm:"type:text"`        // 地址
	Notes     string         `json:"notes" gorm:"type:text"`          // 备注
	Balance   float64        `json:"balance" gorm:"type:decimal(15,2);default:0"` // 账户余额
	LastLoginAt *time.Time   `json:"last_login_at"`                   // 最后登录时间
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	Transactions []Transaction `json:"transactions,omitempty" gorm:"foreignKey:UserID"`
	UserCoupons  []UserCoupon  `json:"user_coupons,omitempty" gorm:"foreignKey:UserID"`
}

func (Customer) TableName() string {
	return "customers"
}

// IsActive 检查客户是否激活
func (c *Customer) IsActive() bool {
	return c.Status == CustomerStatusActive
}

// IsBlocked 检查客户是否被阻止
func (c *Customer) IsBlocked() bool {
	return c.Status == CustomerStatusBlocked
}

// Block 阻止客户
func (c *Customer) Block() {
	c.Status = CustomerStatusBlocked
}

// Activate 激活客户
func (c *Customer) Activate() {
	c.Status = CustomerStatusActive
}

// Deactivate 停用客户
func (c *Customer) Deactivate() {
	c.Status = CustomerStatusInactive
}

// UpdateBalance 更新余额
func (c *Customer) UpdateBalance(amount float64) {
	c.Balance += amount
	if c.Balance < 0 {
		c.Balance = 0
	}
}

// CanMakeTransaction 检查是否可以进行交易
func (c *Customer) CanMakeTransaction(amount float64) bool {
	if !c.IsActive() {
		return false
	}
	return c.Balance >= amount
}

// RecordLogin 记录登录时间
func (c *Customer) RecordLogin() {
	now := time.Now()
	c.LastLoginAt = &now
}

// GetStatusString 获取状态字符串
func (c *Customer) GetStatusString() string {
	switch c.Status {
	case CustomerStatusActive:
		return "激活"
	case CustomerStatusInactive:
		return "未激活"
	case CustomerStatusBlocked:
		return "已阻止"
	default:
		return "未知"
	}
}