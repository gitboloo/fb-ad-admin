package repositories

import (
	"fmt"

	"backend/database"
	"backend/models"
	"backend/types"
	"gorm.io/gorm"
)

// UserCouponRepository 用户优惠券仓库
type UserCouponRepository struct {
	db *gorm.DB
}

// NewUserCouponRepository 创建用户优惠券仓库
func NewUserCouponRepository() *UserCouponRepository {
	return &UserCouponRepository{
		db: database.DB,
	}
}

// Create 创建用户优惠券
func (ucr *UserCouponRepository) Create(userCoupon *models.UserCoupon) error {
	return ucr.db.Create(userCoupon).Error
}

// Update 更新用户优惠券
func (ucr *UserCouponRepository) Update(userCoupon *models.UserCoupon) error {
	return ucr.db.Save(userCoupon).Error
}

// GetByID 根据ID获取用户优惠券
func (ucr *UserCouponRepository) GetByID(id uint) (*models.UserCoupon, error) {
	var userCoupon models.UserCoupon
	if err := ucr.db.Preload("Coupon").First(&userCoupon, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户优惠券不存在")
		}
		return nil, err
	}
	return &userCoupon, nil
}

// GetByUserID 根据用户ID获取优惠券列表
func (ucr *UserCouponRepository) GetByUserID(userID uint, req *types.FilterRequest, status *models.CouponStatus) ([]*models.UserCoupon, int64, error) {
	var userCoupons []*models.UserCoupon
	var total int64

	query := ucr.db.Model(&models.UserCoupon{}).Preload("Coupon").Where("user_id = ?", userID)

	// 状态筛选
	if status != nil {
		// 这里筛选优惠券本身的状态，不是用户优惠券的状态
		query = query.Joins("JOIN coupons ON coupons.id = user_coupons.coupon_id").
			Where("coupons.status = ?", *status)
	}

	// 搜索条件
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Joins("JOIN coupons ON coupons.id = user_coupons.coupon_id").
			Where("coupons.name LIKE ? OR coupons.description LIKE ?", searchPattern, searchPattern)
	}

	// 日期范围筛选
	if req.StartDate != nil {
		query = query.Where("user_coupons.created_at >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("user_coupons.created_at <= ?", req.EndDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序和分页
	orderClause := fmt.Sprintf("user_coupons.%s %s", req.GetSort(), req.GetOrder())
	if err := query.Order(orderClause).
		Offset(req.GetOffset()).
		Limit(req.GetSize()).
		Find(&userCoupons).Error; err != nil {
		return nil, 0, err
	}

	return userCoupons, total, nil
}

// GetByCouponID 根据优惠券ID获取用户优惠券列表
func (ucr *UserCouponRepository) GetByCouponID(couponID uint) ([]*models.UserCoupon, error) {
	var userCoupons []*models.UserCoupon
	if err := ucr.db.Where("coupon_id = ?", couponID).Find(&userCoupons).Error; err != nil {
		return nil, err
	}
	return userCoupons, nil
}

// GetByUserAndCoupon 根据用户ID和优惠券ID获取用户优惠券
func (ucr *UserCouponRepository) GetByUserAndCoupon(userID uint, couponID uint) (*models.UserCoupon, error) {
	var userCoupon models.UserCoupon
	if err := ucr.db.Where("user_id = ? AND coupon_id = ?", userID, couponID).First(&userCoupon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &userCoupon, nil
}

// GetAvailableByUserID 获取用户可用的优惠券
func (ucr *UserCouponRepository) GetAvailableByUserID(userID uint) ([]*models.UserCoupon, error) {
	var userCoupons []*models.UserCoupon
	if err := ucr.db.Preload("Coupon").
		Where("user_id = ?", userID).
		Where("status = ?", models.UserCouponStatusUnused).
		Where("expired_at > NOW()").
		Find(&userCoupons).Error; err != nil {
		return nil, err
	}
	return userCoupons, nil
}

// GetUsedByUserID 获取用户已使用的优惠券
func (ucr *UserCouponRepository) GetUsedByUserID(userID uint) ([]*models.UserCoupon, error) {
	var userCoupons []*models.UserCoupon
	if err := ucr.db.Preload("Coupon").
		Where("user_id = ?", userID).
		Where("status = ?", models.UserCouponStatusUsed).
		Find(&userCoupons).Error; err != nil {
		return nil, err
	}
	return userCoupons, nil
}

// GetExpiredByUserID 获取用户已过期的优惠券
func (ucr *UserCouponRepository) GetExpiredByUserID(userID uint) ([]*models.UserCoupon, error) {
	var userCoupons []*models.UserCoupon
	if err := ucr.db.Preload("Coupon").
		Where("user_id = ?", userID).
		Where("status = ? OR expired_at <= NOW()", models.UserCouponStatusExpired).
		Find(&userCoupons).Error; err != nil {
		return nil, err
	}
	return userCoupons, nil
}

// BatchExpire 批量过期优惠券
func (ucr *UserCouponRepository) BatchExpire(ids []uint) error {
	return ucr.db.Model(&models.UserCoupon{}).
		Where("id IN ?", ids).
		Update("status", models.UserCouponStatusExpired).Error
}

// GetExpiredUserCoupons 获取已过期但状态未更新的用户优惠券
func (ucr *UserCouponRepository) GetExpiredUserCoupons() ([]*models.UserCoupon, error) {
	var userCoupons []*models.UserCoupon
	if err := ucr.db.Where("status = ?", models.UserCouponStatusUnused).
		Where("expired_at <= NOW()").
		Find(&userCoupons).Error; err != nil {
		return nil, err
	}
	return userCoupons, nil
}

// GetStatsByUser 获取用户优惠券统计
func (ucr *UserCouponRepository) GetStatsByUser(userID uint) (map[string]interface{}, error) {
	var total int64
	var available int64
	var used int64
	var expired int64

	// 总数
	ucr.db.Model(&models.UserCoupon{}).Where("user_id = ?", userID).Count(&total)

	// 可用数量
	ucr.db.Model(&models.UserCoupon{}).
		Where("user_id = ?", userID).
		Where("status = ?", models.UserCouponStatusUnused).
		Where("expired_at > NOW()").Count(&available)

	// 已使用数量
	ucr.db.Model(&models.UserCoupon{}).
		Where("user_id = ?", userID).
		Where("status = ?", models.UserCouponStatusUsed).Count(&used)

	// 已过期数量
	ucr.db.Model(&models.UserCoupon{}).
		Where("user_id = ?", userID).
		Where("status = ? OR expired_at <= NOW()", models.UserCouponStatusExpired).Count(&expired)

	return map[string]interface{}{
		"total":     total,
		"available": available,
		"used":      used,
		"expired":   expired,
	}, nil
}

// Delete 删除用户优惠券
func (ucr *UserCouponRepository) Delete(id uint) error {
	result := ucr.db.Delete(&models.UserCoupon{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("用户优惠券不存在")
	}
	return nil
}

// GetCountByCoupon 获取优惠券的领取数量
func (ucr *UserCouponRepository) GetCountByCoupon(couponID uint) (int64, error) {
	var count int64
	if err := ucr.db.Model(&models.UserCoupon{}).
		Where("coupon_id = ?", couponID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetUsageStatsByCoupon 获取优惠券的使用统计
func (ucr *UserCouponRepository) GetUsageStatsByCoupon(couponID uint) (map[string]interface{}, error) {
	var total int64
	var used int64
	var available int64
	var expired int64

	// 总领取数量
	ucr.db.Model(&models.UserCoupon{}).Where("coupon_id = ?", couponID).Count(&total)

	// 已使用数量
	ucr.db.Model(&models.UserCoupon{}).
		Where("coupon_id = ?", couponID).
		Where("status = ?", models.UserCouponStatusUsed).Count(&used)

	// 可用数量
	ucr.db.Model(&models.UserCoupon{}).
		Where("coupon_id = ?", couponID).
		Where("status = ?", models.UserCouponStatusUnused).
		Where("expired_at > NOW()").Count(&available)

	// 过期数量
	ucr.db.Model(&models.UserCoupon{}).
		Where("coupon_id = ?", couponID).
		Where("status = ? OR expired_at <= NOW()", models.UserCouponStatusExpired).Count(&expired)

	return map[string]interface{}{
		"total":     total,
		"used":      used,
		"available": available,
		"expired":   expired,
		"usage_rate": func() float64 {
			if total > 0 {
				return float64(used) / float64(total) * 100
			}
			return 0
		}(),
	}, nil
}