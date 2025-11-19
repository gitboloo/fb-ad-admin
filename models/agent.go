package models

import (
	"math/rand"
	"time"

	"gorm.io/gorm"
)

// AgentStatus 代理商状态
type AgentStatus int

const (
	AgentStatusDisabled AgentStatus = 0 // 禁用
	AgentStatusActive   AgentStatus = 1 // 正常
)

// AgentLevel 代理商等级
type AgentLevel int

const (
	AgentLevelFirst  AgentLevel = 1 // 一级代理
	AgentLevelSecond AgentLevel = 2 // 二级代理
	AgentLevelThird  AgentLevel = 3 // 三级代理
)

// Agent 代理商模型 - 与Admin一一对应的代理身份配置
type Agent struct {
	ID         uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	AdminID    uint   `json:"admin_id" gorm:"type:bigint;uniqueIndex;not null;comment:关联的管理员ID"`
	InviteCode string `json:"invite_code" gorm:"type:varchar(50);uniqueIndex;not null;default:'';comment:邀请码"`

	// 代理商等级
	AgentLevel AgentLevel `json:"agent_level" gorm:"type:tinyint;not null;default:1;comment:代理等级(1=一级,2=二级,3=三级)"`

	// 上级代理（通过parent_id引用其他Agent的admin_id来确定层级关系）
	ParentAdminID *uint `json:"parent_admin_id" gorm:"type:bigint;index;comment:上级代理的AdminID"`

	// 功能权限
	EnableGoogleAuth          bool `json:"enable_google_auth" gorm:"type:tinyint(1);not null;default:0;comment:是否开启google验证"`
	CanDispatchOrders         bool `json:"can_dispatch_orders" gorm:"type:tinyint(1);not null;default:0;comment:是否可以派单"`
	CanModifyCustomerBankCard bool `json:"can_modify_customer_bank_card" gorm:"type:tinyint(1);not null;default:0;comment:是否可以修改客户银行卡信息"`

	// 备注
	Remark string `json:"remark" gorm:"type:text;comment:备注"`

	// 状态（继承自关联Admin，这里冗余存储以便查询）
	Status AgentStatus `json:"status" gorm:"type:tinyint;not null;default:1;comment:状态(0=禁用,1=正常)"`

	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Admin       *Admin  `json:"admin,omitempty" gorm:"foreignKey:AdminID;references:ID"`
	ParentAgent *Agent  `json:"parent,omitempty" gorm:"foreignKey:ParentAdminID;references:AdminID"`
	Children    []Agent `json:"children,omitempty" gorm:"foreignKey:ParentAdminID;references:AdminID"`
}

func (Agent) TableName() string {
	return "agents"
}

// BeforeCreate 创建前钩子 - 生成邀请码
func (a *Agent) BeforeCreate(tx *gorm.DB) error {
	// 生成邀请码（如果未设置）
	if a.InviteCode == "" {
		a.InviteCode = generateInviteCode()
	}
	return nil
}

// generateInviteCode 生成唯一的邀请码
func generateInviteCode() string {
	return "INV" + randomString(12)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// IsActive 检查是否激活
func (a *Agent) IsActive() bool {
	return a.Status == AgentStatusActive
}

// GetStatusString 获取状态字符串
func (a *Agent) GetStatusString() string {
	switch a.Status {
	case AgentStatusActive:
		return "正常"
	case AgentStatusDisabled:
		return "禁用"
	default:
		return "未知"
	}
}

// GetLevelString 获取等级字符串
func (a *Agent) GetLevelString() string {
	switch a.AgentLevel {
	case AgentLevelFirst:
		return "一级代理"
	case AgentLevelSecond:
		return "二级代理"
	case AgentLevelThird:
		return "三级代理"
	default:
		return "未知"
	}
}
