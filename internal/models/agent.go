package models

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AgentStatus 代理商状态
type AgentStatus int

const (
	AgentStatusPending  AgentStatus = 0 // 待审核
	AgentStatusActive   AgentStatus = 1 // 正常
	AgentStatusFrozen   AgentStatus = 2 // 冻结
	AgentStatusDisabled AgentStatus = 3 // 禁用
)

// AgentLevel 代理商等级
type AgentLevel int

const (
	AgentLevelFirst  AgentLevel = 1 // 一级代理
	AgentLevelSecond AgentLevel = 2 // 二级代理
	AgentLevelThird  AgentLevel = 3 // 三级代理
)

// WithdrawalMethod 提现方式
type WithdrawalMethod int

const (
	WithdrawalMethodBank   WithdrawalMethod = 1 // 银行卡
	WithdrawalMethodAlipay WithdrawalMethod = 2 // 支付宝
	WithdrawalMethodWechat WithdrawalMethod = 3 // 微信
)

// Agent 代理商模型
type Agent struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Username string `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Account  string `json:"account" gorm:"type:varchar(100);uniqueIndex;not null"`
	Password string `json:"-" gorm:"type:varchar(255);not null"`

	// 基本信息
	RealName string `json:"real_name" gorm:"type:varchar(100);comment:真实姓名"`
	Phone    string `json:"phone" gorm:"type:varchar(20);index;comment:手机号"`
	Email    string `json:"email" gorm:"type:varchar(100);comment:邮箱"`
	Company  string `json:"company" gorm:"type:varchar(255);comment:公司名称"`

	// 代理商层级信息
	AgentLevel AgentLevel `json:"agent_level" gorm:"type:tinyint;not null;default:1;comment:代理等级"`
	ParentID   *uint      `json:"parent_id" gorm:"comment:上级代理ID"`
	AgentCode  string     `json:"agent_code" gorm:"type:varchar(50);uniqueIndex;not null;comment:代理商唯一编码"`

	// 分润配置
	CommissionRate     float64 `json:"commission_rate" gorm:"type:decimal(5,2);not null;default:0.00;comment:分润比例(%)"`
	SelfCommissionRate float64 `json:"self_commission_rate" gorm:"type:decimal(5,2);not null;default:0.00;comment:自购分润比例(%)"`

	// 财务信息
	Balance         float64 `json:"balance" gorm:"type:decimal(15,2);not null;default:0.00;comment:账户余额"`
	TotalCommission float64 `json:"total_commission" gorm:"type:decimal(15,2);not null;default:0.00;comment:累计佣金"`
	FrozenBalance   float64 `json:"frozen_balance" gorm:"type:decimal(15,2);not null;default:0.00;comment:冻结余额"`

	// 业绩统计
	TotalCustomers  uint    `json:"total_customers" gorm:"not null;default:0;comment:累计客户数"`
	ActiveCustomers uint    `json:"active_customers" gorm:"not null;default:0;comment:活跃客户数"`
	TotalOrders     uint    `json:"total_orders" gorm:"not null;default:0;comment:累计订单数"`
	TotalSales      float64 `json:"total_sales" gorm:"type:decimal(15,2);not null;default:0.00;comment:累计销售额"`

	// 状态管理
	Status     AgentStatus `json:"status" gorm:"type:tinyint;not null;default:0;comment:状态"`
	IsVerified bool        `json:"is_verified" gorm:"type:tinyint(1);not null;default:0;comment:是否实名认证"`

	// 认证信息
	Address         string `json:"address" gorm:"type:varchar(500);comment:联系地址"`
	IDCard          string `json:"id_card" gorm:"type:varchar(18);comment:身份证号"`
	BusinessLicense string `json:"business_license" gorm:"type:varchar(255);comment:营业执照URL"`

	// 备注
	Notes        string `json:"notes" gorm:"type:text;comment:备注信息"`
	RejectReason string `json:"reject_reason" gorm:"type:varchar(500);comment:拒绝原因"`

	// 时间戳
	LastLoginAt *time.Time     `json:"last_login_at" gorm:"type:datetime;comment:最后登录时间"`
	VerifiedAt  *time.Time     `json:"verified_at" gorm:"type:datetime;comment:认证时间"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Parent       *Agent           `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children     []Agent          `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Customers    []AgentCustomer  `json:"customers,omitempty" gorm:"foreignKey:AgentID"`
	Commissions  []Commission     `json:"commissions,omitempty" gorm:"foreignKey:AgentID"`
	Withdrawals  []Withdrawal     `json:"withdrawals,omitempty" gorm:"foreignKey:AgentID"`
	AgentAuthCodes []AgentAuthCode `json:"agent_auth_codes,omitempty" gorm:"foreignKey:AgentID"`
}

func (Agent) TableName() string {
	return "agents"
}

// BeforeCreate 创建前钩子
func (a *Agent) BeforeCreate(tx *gorm.DB) error {
	// 加密密码
	if a.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		a.Password = string(hashedPassword)
	}

	// 生成代理商编码（如果未设置）
	if a.AgentCode == "" {
		a.AgentCode = a.GenerateAgentCode(tx)
	}

	return nil
}

// GenerateAgentCode 生成代理商编码
func (a *Agent) GenerateAgentCode(tx *gorm.DB) string {
	if a.ParentID == nil {
		// 一级代理：A001, A002, ...
		var count int64
		tx.Model(&Agent{}).Where("parent_id IS NULL").Count(&count)
		return fmt.Sprintf("A%03d", count+1)
	}

	// 二级/三级代理：父级编码-B001 或 父级编码-C001
	var parent Agent
	if err := tx.First(&parent, a.ParentID).Error; err != nil {
		return fmt.Sprintf("TEMP-%d", time.Now().Unix())
	}

	var count int64
	tx.Model(&Agent{}).Where("parent_id = ?", a.ParentID).Count(&count)

	prefix := "B"
	if a.AgentLevel == AgentLevelThird {
		prefix = "C"
	}

	return fmt.Sprintf("%s-%s%03d", parent.AgentCode, prefix, count+1)
}

// CheckPassword 验证密码
func (a *Agent) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}

// SetPassword 设置密码
func (a *Agent) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.Password = string(hashedPassword)
	return nil
}

// IsActive 检查是否激活
func (a *Agent) IsActive() bool {
	return a.Status == AgentStatusActive
}

// IsPending 检查是否待审核
func (a *Agent) IsPending() bool {
	return a.Status == AgentStatusPending
}

// IsFrozen 检查是否冻结
func (a *Agent) IsFrozen() bool {
	return a.Status == AgentStatusFrozen
}

// Approve 审核通过
func (a *Agent) Approve() {
	a.Status = AgentStatusActive
	a.RejectReason = ""
}

// Reject 审核拒绝
func (a *Agent) Reject(reason string) {
	a.Status = AgentStatusDisabled
	a.RejectReason = reason
}

// Freeze 冻结账户
func (a *Agent) Freeze() {
	a.Status = AgentStatusFrozen
}

// Unfreeze 解冻账户
func (a *Agent) Unfreeze() {
	a.Status = AgentStatusActive
}

// Verify 实名认证
func (a *Agent) Verify() {
	a.IsVerified = true
	now := time.Now()
	a.VerifiedAt = &now
}

// RecordLogin 记录登录
func (a *Agent) RecordLogin() {
	now := time.Now()
	a.LastLoginAt = &now
}

// AddCommission 添加佣金
func (a *Agent) AddCommission(amount float64) {
	a.Balance += amount
	a.TotalCommission += amount
}

// FreezeAmount 冻结金额
func (a *Agent) FreezeAmount(amount float64) error {
	if a.Balance < amount {
		return fmt.Errorf("余额不足")
	}
	a.Balance -= amount
	a.FrozenBalance += amount
	return nil
}

// UnfreezeAmount 解冻金额
func (a *Agent) UnfreezeAmount(amount float64) {
	a.FrozenBalance -= amount
	a.Balance += amount
}

// DeductBalance 扣除余额
func (a *Agent) DeductBalance(amount float64) error {
	if a.Balance < amount {
		return fmt.Errorf("余额不足")
	}
	a.Balance -= amount
	return nil
}

// GetStatusString 获取状态字符串
func (a *Agent) GetStatusString() string {
	switch a.Status {
	case AgentStatusPending:
		return "待审核"
	case AgentStatusActive:
		return "正常"
	case AgentStatusFrozen:
		return "冻结"
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

// GetAvailableBalance 获取可用余额
func (a *Agent) GetAvailableBalance() float64 {
	return a.Balance - a.FrozenBalance
}

// CanWithdraw 检查是否可以提现
func (a *Agent) CanWithdraw(amount float64) bool {
	return a.IsActive() && a.GetAvailableBalance() >= amount
}

// AgentCustomer 代理商-客户关系
type AgentCustomer struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID        uint      `json:"agent_id" gorm:"not null;index;comment:代理商ID"`
	CustomerID     uint      `json:"customer_id" gorm:"uniqueIndex;not null;comment:客户ID"`
	RegisterSource string    `json:"register_source" gorm:"type:varchar(50);comment:注册来源"`
	AuthCode       string    `json:"auth_code" gorm:"type:varchar(100);comment:使用的授权码"`
	BindAt         time.Time `json:"bind_at" gorm:"not null;comment:绑定时间"`
	CreatedAt      time.Time `json:"created_at"`

	// 关联
	Agent    *Agent    `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
	Customer *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
}

func (AgentCustomer) TableName() string {
	return "agent_customers"
}

// Commission 佣金记录
type Commission struct {
	ID         uint    `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID    uint    `json:"agent_id" gorm:"not null;index;comment:代理商ID"`
	OrderID    uint    `json:"order_id" gorm:"not null;index;comment:订单ID"`
	CustomerID uint    `json:"customer_id" gorm:"not null;index;comment:客户ID"`

	// 佣金信息
	OrderAmount      float64 `json:"order_amount" gorm:"type:decimal(15,2);not null;comment:订单金额"`
	CommissionRate   float64 `json:"commission_rate" gorm:"type:decimal(5,2);not null;comment:分润比例"`
	CommissionAmount float64 `json:"commission_amount" gorm:"type:decimal(15,2);not null;comment:佣金金额"`

	// 层级信息
	AgentLevel     AgentLevel `json:"agent_level" gorm:"type:tinyint;not null;comment:代理等级"`
	CommissionType int        `json:"commission_type" gorm:"type:tinyint;not null;comment:佣金类型"`

	// 结算信息
	Status     int        `json:"status" gorm:"type:tinyint;not null;default:0;comment:状态"`
	SettledAt  *time.Time `json:"settled_at" gorm:"type:datetime;comment:结算时间"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// 关联
	Agent    *Agent    `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
	Customer *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
}

func (Commission) TableName() string {
	return "commissions"
}

// Settle 结算佣金
func (c *Commission) Settle() {
	c.Status = 1
	now := time.Now()
	c.SettledAt = &now
}

// Cancel 取消佣金
func (c *Commission) Cancel() {
	c.Status = 2
}

// Withdrawal 提现记录
type Withdrawal struct {
	ID           uint    `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID      uint    `json:"agent_id" gorm:"not null;index;comment:代理商ID"`
	Amount       float64 `json:"amount" gorm:"type:decimal(15,2);not null;comment:提现金额"`
	Fee          float64 `json:"fee" gorm:"type:decimal(15,2);not null;default:0.00;comment:手续费"`
	ActualAmount float64 `json:"actual_amount" gorm:"type:decimal(15,2);not null;comment:实际到账金额"`

	// 收款信息
	WithdrawalMethod WithdrawalMethod `json:"withdrawal_method" gorm:"type:tinyint;not null;comment:提现方式"`
	AccountName      string           `json:"account_name" gorm:"type:varchar(100);not null;comment:账户名"`
	AccountNumber    string           `json:"account_number" gorm:"type:varchar(100);not null;comment:账户号码"`
	BankName         string           `json:"bank_name" gorm:"type:varchar(100);comment:银行名称"`

	// 状态管理
	Status       int        `json:"status" gorm:"type:tinyint;not null;default:0;comment:状态"`
	RejectReason string     `json:"reject_reason" gorm:"type:varchar(500);comment:拒绝原因"`
	ApprovedBy   *uint      `json:"approved_by" gorm:"comment:审核人ID"`
	ApprovedAt   *time.Time `json:"approved_at" gorm:"type:datetime;comment:审核时间"`
	CompletedAt  *time.Time `json:"completed_at" gorm:"type:datetime;comment:完成时间"`

	Notes     string    `json:"notes" gorm:"type:varchar(500);comment:备注"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Agent    *Agent `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
	Approver *Admin `json:"approver,omitempty" gorm:"foreignKey:ApprovedBy"`
}

func (Withdrawal) TableName() string {
	return "withdrawals"
}

// Approve 审核通过
func (w *Withdrawal) Approve(approverID uint) {
	w.Status = 1
	w.ApprovedBy = &approverID
	now := time.Now()
	w.ApprovedAt = &now
}

// Reject 审核拒绝
func (w *Withdrawal) Reject(reason string, approverID uint) {
	w.Status = 3
	w.RejectReason = reason
	w.ApprovedBy = &approverID
	now := time.Now()
	w.ApprovedAt = &now
}

// Complete 完成提现
func (w *Withdrawal) Complete() {
	w.Status = 2
	now := time.Now()
	w.CompletedAt = &now
}

// GetStatusString 获取状态字符串
func (w *Withdrawal) GetStatusString() string {
	switch w.Status {
	case 0:
		return "待审核"
	case 1:
		return "处理中"
	case 2:
		return "已完成"
	case 3:
		return "已拒绝"
	default:
		return "未知"
	}
}

// GetMethodString 获取提现方式字符串
func (w *Withdrawal) GetMethodString() string {
	switch w.WithdrawalMethod {
	case WithdrawalMethodBank:
		return "银行卡"
	case WithdrawalMethodAlipay:
		return "支付宝"
	case WithdrawalMethodWechat:
		return "微信"
	default:
		return "未知"
	}
}

// AgentAuthCode 代理商专属授权码
type AgentAuthCode struct {
	ID           uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID      uint       `json:"agent_id" gorm:"not null;index;comment:代理商ID"`
	Code         string     `json:"code" gorm:"type:varchar(100);uniqueIndex;not null;comment:授权码"`
	Type         int        `json:"type" gorm:"type:tinyint;not null;default:1;comment:类型"`
	MaxUses      *int       `json:"max_uses" gorm:"comment:最大使用次数"`
	UsedCount    int        `json:"used_count" gorm:"not null;default:0;comment:已使用次数"`
	DiscountRate *float64   `json:"discount_rate" gorm:"type:decimal(5,2);comment:优惠比例"`
	Status       int        `json:"status" gorm:"type:tinyint;not null;default:1;comment:状态"`
	ExpiredAt    *time.Time `json:"expired_at" gorm:"type:datetime;comment:过期时间"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 关联
	Agent *Agent `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
}

func (AgentAuthCode) TableName() string {
	return "agent_auth_codes"
}

// IsValid 检查授权码是否有效
func (a *AgentAuthCode) IsValid() bool {
	if a.Status != 1 {
		return false
	}

	// 检查是否过期
	if a.ExpiredAt != nil && a.ExpiredAt.Before(time.Now()) {
		return false
	}

	// 检查使用次数
	if a.MaxUses != nil && a.UsedCount >= *a.MaxUses {
		return false
	}

	return true
}

// Use 使用授权码
func (a *AgentAuthCode) Use() error {
	if !a.IsValid() {
		return fmt.Errorf("授权码无效或已过期")
	}

	a.UsedCount++

	// 如果达到最大使用次数，更新状态
	if a.MaxUses != nil && a.UsedCount >= *a.MaxUses {
		a.Status = 2 // 已用完
	}

	return nil
}

// Disable 禁用授权码
func (a *AgentAuthCode) Disable() {
	a.Status = 0
}

// Enable 启用授权码
func (a *AgentAuthCode) Enable() {
	a.Status = 1
}
