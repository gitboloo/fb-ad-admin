package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type CouponStatus int
type CouponType int
type ValidityType int

const (
	CouponStatusInactive CouponStatus = 0
	CouponStatusActive   CouponStatus = 1
	CouponStatusExpired  CouponStatus = 2
	CouponStatusUsedUp   CouponStatus = 3
)

const (
	CouponTypeValueAdded CouponType = 1 // 增值券
	CouponTypeDiscount   CouponType = 2 // 抵扣券
	CouponTypeTeam       CouponType = 3 // 团队券
	CouponTypeCustom     CouponType = 4 // 自定义
	CouponTypeFixed      CouponType = 5 // 固定金额
)

const (
	ValidityTypeDays  ValidityType = 1 // 按天数计算
	ValidityTypeRange ValidityType = 2 // 按日期范围
)

type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// Value 实现 driver.Valuer 接口
func (dr DateRange) Value() (driver.Value, error) {
	return json.Marshal(dr)
}

func (dr *DateRange) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}
	
	return json.Unmarshal(bytes, dr)
}

type Coupon struct {
	ID              uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name            string         `json:"name" gorm:"type:varchar(255);not null"`
	Description     string         `json:"description" gorm:"type:text"`
	IsNewUser       bool           `json:"is_new_user" gorm:"type:boolean;default:false"` // 是否仅新用户可用
	Type            CouponType     `json:"type" gorm:"type:tinyint;not null;default:1"`
	DiscountPercent float64        `json:"discount_percent" gorm:"type:decimal(5,2);default:0"` // 折扣百分比
	MinAmount       float64        `json:"min_amount" gorm:"type:decimal(10,2);default:0"`      // 最小使用金额
	MaxAmount       float64        `json:"max_amount" gorm:"type:decimal(10,2);default:0"`      // 最大优惠金额
	ValidityType    ValidityType   `json:"validity_type" gorm:"type:tinyint;not null;default:1"`
	ValidityDays    int            `json:"validity_days" gorm:"type:int;default:0"`              // 有效天数
	DateRange       *DateRange     `json:"date_range" gorm:"type:json"`                         // 有效日期范围
	Status          CouponStatus   `json:"status" gorm:"type:tinyint;not null;default:1"`
	TotalCount      int            `json:"total_count" gorm:"type:int;default:0"`               // 总发放数量，0表示无限制
	UsedCount       int            `json:"used_count" gorm:"type:int;default:0"`                // 已使用数量
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	UserCoupons []UserCoupon `json:"user_coupons,omitempty" gorm:"foreignKey:CouponID"`
}

func (Coupon) TableName() string {
	return "coupons"
}

// IsActive 检查优惠券是否激活
func (c *Coupon) IsActive() bool {
	return c.Status == CouponStatusActive
}

// IsAvailable 检查优惠券是否可用
func (c *Coupon) IsAvailable() bool {
	if !c.IsActive() {
		return false
	}
	
	// 检查数量限制
	if c.TotalCount > 0 && c.UsedCount >= c.TotalCount {
		return false
	}
	
	// 检查有效期
	now := time.Now()
	if c.ValidityType == ValidityTypeRange && c.DateRange != nil {
		return now.After(c.DateRange.StartDate) && now.Before(c.DateRange.EndDate)
	}
	
	return true
}

// GetDiscountAmount 计算折扣金额
func (c *Coupon) GetDiscountAmount(orderAmount float64) float64 {
	if orderAmount < c.MinAmount {
		return 0
	}
	
	var discount float64
	switch c.Type {
	case CouponTypeDiscount:
		discount = orderAmount * c.DiscountPercent / 100
	case CouponTypeFixed:
		discount = c.DiscountPercent // 固定金额时，DiscountPercent字段存储固定金额
	case CouponTypeValueAdded:
		discount = c.DiscountPercent // 增值券，直接增加金额
	default:
		discount = orderAmount * c.DiscountPercent / 100
	}
	
	// 检查最大优惠金额限制
	if c.MaxAmount > 0 && discount > c.MaxAmount {
		discount = c.MaxAmount
	}
	
	return discount
}

// CanUseForNewUser 检查是否可以给新用户使用
func (c *Coupon) CanUseForNewUser(isNewUser bool) bool {
	if c.IsNewUser && !isNewUser {
		return false
	}
	return true
}