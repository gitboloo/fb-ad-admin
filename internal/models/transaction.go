package models

import (
	"time"

	"gorm.io/gorm"
)

type TransactionType int
type TransactionStatus int

const (
	TransactionTypeRecharge  TransactionType = 1 // 充值
	TransactionTypeWithdraw  TransactionType = 2 // 提现
	TransactionTypeConsume   TransactionType = 3 // 消费
	TransactionTypeRefund    TransactionType = 4 // 退款
	TransactionTypeReward    TransactionType = 5 // 奖励
)

const (
	TransactionStatusPending   TransactionStatus = 1 // 待处理
	TransactionStatusSuccess   TransactionStatus = 2 // 成功
	TransactionStatusFailed    TransactionStatus = 3 // 失败
	TransactionStatusCancelled TransactionStatus = 4 // 已取消
	TransactionStatusProcessing TransactionStatus = 5 // 处理中
)

type Transaction struct {
	ID            uint              `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID        uint              `json:"user_id" gorm:"not null;index"`
	Type          TransactionType   `json:"type" gorm:"type:tinyint;not null"`
	Amount        float64           `json:"amount" gorm:"type:decimal(15,2);not null"`
	Status        TransactionStatus `json:"status" gorm:"type:tinyint;not null;default:1"`
	Description   string            `json:"description" gorm:"type:varchar(500)"`
	OrderNo       string            `json:"order_no" gorm:"type:varchar(64);uniqueIndex"`     // 订单号
	PaymentMethod string            `json:"payment_method" gorm:"type:varchar(50)"`           // 支付方式
	PaymentID     string            `json:"payment_id" gorm:"type:varchar(100)"`              // 第三方支付ID
	BalanceBefore float64           `json:"balance_before" gorm:"type:decimal(15,2);default:0"` // 交易前余额
	BalanceAfter  float64           `json:"balance_after" gorm:"type:decimal(15,2);default:0"`  // 交易后余额
	ProcessedAt   *time.Time        `json:"processed_at"`                                     // 处理时间
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `json:"-" gorm:"index"`
	
	// 关联
	Customer      Customer          `json:"customer,omitempty" gorm:"foreignKey:UserID"`
}

func (Transaction) TableName() string {
	return "transactions"
}

// BeforeCreate 在创建前生成订单号
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.OrderNo == "" {
		t.OrderNo = generateOrderNo()
	}
	return nil
}

// generateOrderNo 生成订单号
func generateOrderNo() string {
	// 格式：TXN + 年月日时分秒 + 6位随机数
	now := time.Now()
	return now.Format("TXN20060102150405") + generateRandomString(6)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// IsSuccess 检查交易是否成功
func (t *Transaction) IsSuccess() bool {
	return t.Status == TransactionStatusSuccess
}

// IsPending 检查交易是否待处理
func (t *Transaction) IsPending() bool {
	return t.Status == TransactionStatusPending || t.Status == TransactionStatusProcessing
}

// CanCancel 检查是否可以取消
func (t *Transaction) CanCancel() bool {
	return t.Status == TransactionStatusPending
}

// Complete 完成交易
func (t *Transaction) Complete(balanceAfter float64) {
	now := time.Now()
	t.Status = TransactionStatusSuccess
	t.BalanceAfter = balanceAfter
	t.ProcessedAt = &now
}

// Fail 交易失败
func (t *Transaction) Fail() {
	now := time.Now()
	t.Status = TransactionStatusFailed
	t.ProcessedAt = &now
}

// Cancel 取消交易
func (t *Transaction) Cancel() {
	now := time.Now()
	t.Status = TransactionStatusCancelled
	t.ProcessedAt = &now
}

// GetTypeString 获取交易类型字符串
func (t *Transaction) GetTypeString() string {
	switch t.Type {
	case TransactionTypeRecharge:
		return "充值"
	case TransactionTypeWithdraw:
		return "提现"
	case TransactionTypeConsume:
		return "消费"
	case TransactionTypeRefund:
		return "退款"
	case TransactionTypeReward:
		return "奖励"
	default:
		return "未知"
	}
}

// GetStatusString 获取交易状态字符串
func (t *Transaction) GetStatusString() string {
	switch t.Status {
	case TransactionStatusPending:
		return "待处理"
	case TransactionStatusSuccess:
		return "成功"
	case TransactionStatusFailed:
		return "失败"
	case TransactionStatusCancelled:
		return "已取消"
	case TransactionStatusProcessing:
		return "处理中"
	default:
		return "未知"
	}
}