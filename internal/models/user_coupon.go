package models

import (
	"time"

	"gorm.io/gorm"
)

type UserCouponStatus int

const (
	UserCouponStatusUnused  UserCouponStatus = 1
	UserCouponStatusUsed    UserCouponStatus = 2
	UserCouponStatusExpired UserCouponStatus = 3
)

type UserCoupon struct {
	ID        uint             `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint             `json:"user_id" gorm:"not null;index"`
	CouponID  uint             `json:"coupon_id" gorm:"not null;index"`
	Status    UserCouponStatus `json:"status" gorm:"type:tinyint;not null;default:1"`
	UsedAt    *time.Time       `json:"used_at"`
	ExpiredAt time.Time        `json:"expired_at"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	DeletedAt gorm.DeletedAt   `json:"-" gorm:"index"`
	
	// 关联
	Coupon *Coupon `json:"coupon,omitempty" gorm:"foreignKey:CouponID"`
}

func (UserCoupon) TableName() string {
	return "user_coupons"
}

// IsUsable 检查用户优惠券是否可用
func (uc *UserCoupon) IsUsable() bool {
	if uc.Status != UserCouponStatusUnused {
		return false
	}
	
	// 检查是否过期
	now := time.Now()
	if now.After(uc.ExpiredAt) {
		return false
	}
	
	return true
}

// Use 使用优惠券
func (uc *UserCoupon) Use() {
	now := time.Now()
	uc.Status = UserCouponStatusUsed
	uc.UsedAt = &now
}

// Expire 使优惠券过期
func (uc *UserCoupon) Expire() {
	uc.Status = UserCouponStatusExpired
}

// CalculateExpiredAt 计算过期时间
func (uc *UserCoupon) CalculateExpiredAt(coupon *Coupon) {
	if coupon.ValidityType == ValidityTypeDays {
		uc.ExpiredAt = time.Now().AddDate(0, 0, coupon.ValidityDays)
	} else if coupon.ValidityType == ValidityTypeRange && coupon.DateRange != nil {
		uc.ExpiredAt = coupon.DateRange.EndDate
	} else {
		// 默认30天有效期
		uc.ExpiredAt = time.Now().AddDate(0, 0, 30)
	}
}