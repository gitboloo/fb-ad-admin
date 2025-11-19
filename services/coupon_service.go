package services

import (
	"fmt"
	"time"

	"backend/models"
	"backend/repositories"
	"backend/types"
)

// CouponService 优惠券服务
type CouponService struct {
	couponRepo     *repositories.CouponRepository
	userCouponRepo *repositories.UserCouponRepository
	customerRepo   *repositories.CustomerRepository
}

// NewCouponService 创建优惠券服务
func NewCouponService() *CouponService {
	return &CouponService{
		couponRepo:     repositories.NewCouponRepository(),
		userCouponRepo: repositories.NewUserCouponRepository(),
		customerRepo:   repositories.NewCustomerRepository(),
	}
}

// List 获取优惠券列表
func (cs *CouponService) List(req *types.FilterRequest) ([]*models.Coupon, int64, error) {
	return cs.couponRepo.List(req)
}

// GetByID 根据ID获取优惠券
func (cs *CouponService) GetByID(id uint) (*models.Coupon, error) {
	return cs.couponRepo.GetByID(id)
}

// Create 创建优惠券
func (cs *CouponService) Create(coupon *models.Coupon) error {
	// 验证优惠券配置
	if err := cs.validateCoupon(coupon); err != nil {
		return err
	}

	return cs.couponRepo.Create(coupon)
}

// Update 更新优惠券
func (cs *CouponService) Update(coupon *models.Coupon) error {
	// 验证优惠券配置
	if err := cs.validateCoupon(coupon); err != nil {
		return err
	}

	return cs.couponRepo.Update(coupon)
}

// Delete 删除优惠券
func (cs *CouponService) Delete(id uint) error {
	// 检查是否有用户已领取此优惠券
	userCoupons, err := cs.userCouponRepo.GetByCouponID(id)
	if err != nil {
		return err
	}

	if len(userCoupons) > 0 {
		return &ServiceError{
			Code:    400,
			Message: "该优惠券已被用户领取，无法删除",
		}
	}

	return cs.couponRepo.Delete(id)
}

// DistributeCoupon 分发优惠券给指定用户
func (cs *CouponService) DistributeCoupon(couponID uint, userIDs []uint) (map[string]interface{}, error) {
	// 检查优惠券是否存在且可用
	coupon, err := cs.couponRepo.GetByID(couponID)
	if err != nil {
		return nil, err
	}

	if !coupon.IsAvailable() {
		return nil, &ServiceError{
			Code:    400,
			Message: "优惠券不可用",
		}
	}

	var successCount int
	var failedUsers []uint
	
	for _, userID := range userIDs {
		// 检查用户是否存在
		customer, err := cs.customerRepo.GetByID(userID)
		if err != nil {
			failedUsers = append(failedUsers, userID)
			continue
		}

		// 检查用户是否已经领取过此优惠券
		existingCoupon, _ := cs.userCouponRepo.GetByUserAndCoupon(userID, couponID)
		if existingCoupon != nil {
			failedUsers = append(failedUsers, userID)
			continue
		}

		// 检查新用户限制
		if !coupon.CanUseForNewUser(cs.isNewUser(customer)) {
			failedUsers = append(failedUsers, userID)
			continue
		}

		// 创建用户优惠券
		userCoupon := &models.UserCoupon{
			UserID:    userID,
			CouponID:  couponID,
			Status:    models.UserCouponStatusUnused,
		}

		// 计算过期时间
		userCoupon.CalculateExpiredAt(coupon)

		if err := cs.userCouponRepo.Create(userCoupon); err != nil {
			failedUsers = append(failedUsers, userID)
			continue
		}

		successCount++
	}

	// 更新优惠券使用数量
	if successCount > 0 {
		coupon.UsedCount += successCount
		cs.couponRepo.Update(coupon)
	}

	return map[string]interface{}{
		"success_count": successCount,
		"failed_users":  failedUsers,
		"total_users":   len(userIDs),
	}, nil
}

// ClaimCoupon 用户领取优惠券
func (cs *CouponService) ClaimCoupon(userID uint, couponID uint) (*models.UserCoupon, error) {
	// 检查优惠券是否存在且可用
	coupon, err := cs.couponRepo.GetByID(couponID)
	if err != nil {
		return nil, err
	}

	if !coupon.IsAvailable() {
		return nil, &ServiceError{
			Code:    400,
			Message: "优惠券不可用或已过期",
		}
	}

	// 检查用户是否存在
	customer, err := cs.customerRepo.GetByID(userID)
	if err != nil {
		return nil, &ServiceError{
			Code:    400,
			Message: "用户不存在",
		}
	}

	// 检查用户是否已经领取过此优惠券
	existingCoupon, _ := cs.userCouponRepo.GetByUserAndCoupon(userID, couponID)
	if existingCoupon != nil {
		return nil, &ServiceError{
			Code:    400,
			Message: "您已经领取过此优惠券",
		}
	}

	// 检查新用户限制
	if !coupon.CanUseForNewUser(cs.isNewUser(customer)) {
		return nil, &ServiceError{
			Code:    400,
			Message: "此优惠券仅限新用户领取",
		}
	}

	// 检查数量限制
	if coupon.TotalCount > 0 && coupon.UsedCount >= coupon.TotalCount {
		return nil, &ServiceError{
			Code:    400,
			Message: "优惠券已被领完",
		}
	}

	// 创建用户优惠券
	userCoupon := &models.UserCoupon{
		UserID:    userID,
		CouponID:  couponID,
		Status:    models.UserCouponStatusUnused,
	}

	// 计算过期时间
	userCoupon.CalculateExpiredAt(coupon)

	if err := cs.userCouponRepo.Create(userCoupon); err != nil {
		return nil, err
	}

	// 更新优惠券使用数量
	coupon.UsedCount++
	cs.couponRepo.Update(coupon)

	return userCoupon, nil
}

// UseCoupon 使用优惠券
func (cs *CouponService) UseCoupon(userID uint, userCouponID uint, orderAmount float64) (float64, error) {
	// 获取用户优惠券
	userCoupon, err := cs.userCouponRepo.GetByID(userCouponID)
	if err != nil {
		return 0, err
	}

	// 验证用户是否拥有此优惠券
	if userCoupon.UserID != userID {
		return 0, &ServiceError{
			Code:    403,
			Message: "无权使用此优惠券",
		}
	}

	// 检查优惠券是否可用
	if !userCoupon.IsUsable() {
		return 0, &ServiceError{
			Code:    400,
			Message: "优惠券不可用或已过期",
		}
	}

	// 获取优惠券详情
	coupon, err := cs.couponRepo.GetByID(userCoupon.CouponID)
	if err != nil {
		return 0, err
	}

	// 计算折扣金额
	discountAmount := coupon.GetDiscountAmount(orderAmount)
	if discountAmount <= 0 {
		return 0, &ServiceError{
			Code:    400,
			Message: fmt.Sprintf("订单金额不满足优惠券使用条件，最低消费金额：%.2f", coupon.MinAmount),
		}
	}

	// 标记优惠券为已使用
	userCoupon.Use()
	
	if err := cs.userCouponRepo.Update(userCoupon); err != nil {
		return 0, err
	}

	return discountAmount, nil
}

// GetUserCoupons 获取用户优惠券列表
func (cs *CouponService) GetUserCoupons(userID uint, req *types.FilterRequest, status *models.CouponStatus) ([]*models.UserCoupon, int64, error) {
	return cs.userCouponRepo.GetByUserID(userID, req, status)
}

// GetAvailableCoupons 获取用户可用的优惠券
func (cs *CouponService) GetAvailableCoupons(userID uint, orderAmount float64) ([]*models.UserCoupon, error) {
	userCoupons, err := cs.userCouponRepo.GetAvailableByUserID(userID)
	if err != nil {
		return nil, err
	}

	var availableCoupons []*models.UserCoupon
	for _, userCoupon := range userCoupons {
		// 获取优惠券详情
		coupon, err := cs.couponRepo.GetByID(userCoupon.CouponID)
		if err != nil {
			continue
		}

		// 检查订单金额是否满足使用条件
		if orderAmount > 0 && orderAmount < coupon.MinAmount {
			continue
		}

		// 关联优惠券详情
		userCoupon.Coupon = coupon
		availableCoupons = append(availableCoupons, userCoupon)
	}

	return availableCoupons, nil
}

// GetStatistics 获取优惠券统计
func (cs *CouponService) GetStatistics() (*types.StatisticsResponse, error) {
	return cs.couponRepo.GetStatistics()
}

// validateCoupon 验证优惠券配置
func (cs *CouponService) validateCoupon(coupon *models.Coupon) error {
	// 验证折扣配置
	switch coupon.Type {
	case models.CouponTypeDiscount:
		if coupon.DiscountPercent <= 0 || coupon.DiscountPercent > 100 {
			return &ServiceError{
				Code:    400,
				Message: "折扣百分比必须在0-100之间",
			}
		}
	case models.CouponTypeFixed, models.CouponTypeValueAdded:
		if coupon.DiscountPercent <= 0 {
			return &ServiceError{
				Code:    400,
				Message: "固定金额必须大于0",
			}
		}
	}

	// 验证最大优惠金额
	if coupon.MaxAmount > 0 && coupon.MaxAmount < coupon.DiscountPercent && coupon.Type == models.CouponTypeFixed {
		return &ServiceError{
			Code:    400,
			Message: "最大优惠金额不能小于固定优惠金额",
		}
	}

	// 验证有效期配置
	if coupon.ValidityType == models.ValidityTypeRange {
		if coupon.DateRange == nil {
			return &ServiceError{
				Code:    400,
				Message: "日期范围类型必须设置有效日期范围",
			}
		}
		if coupon.DateRange.StartDate.After(coupon.DateRange.EndDate) {
			return &ServiceError{
				Code:    400,
				Message: "开始日期不能晚于结束日期",
			}
		}
	} else if coupon.ValidityType == models.ValidityTypeDays {
		if coupon.ValidityDays <= 0 {
			return &ServiceError{
				Code:    400,
				Message: "有效天数必须大于0",
			}
		}
	}

	return nil
}

// isNewUser 判断是否为新用户
func (cs *CouponService) isNewUser(customer *models.Customer) bool {
	// 可以根据注册时间、交易记录等判断是否为新用户
	// 这里简单判断注册时间在30天内的为新用户
	return time.Since(customer.CreatedAt) <= 30*24*time.Hour
}