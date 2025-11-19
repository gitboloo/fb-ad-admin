package api

import (
	"strconv"
	"time"

	"backend/models"
	"backend/services"
	"backend/types"
	"backend/utils"
	"github.com/gin-gonic/gin"
)

// CouponRequest 优惠券请求结构
type CouponRequest struct {
	Name            string                `json:"name" binding:"required,max=255"`
	Description     string                `json:"description"`
	IsNewUser       bool                  `json:"is_new_user"`
	Type            models.CouponType     `json:"type" binding:"required,min=1,max=5"`
	DiscountPercent float64               `json:"discount_percent" binding:"required,min=0"`
	MinAmount       float64               `json:"min_amount" binding:"min=0"`
	MaxAmount       float64               `json:"max_amount" binding:"min=0"`
	ValidityType    models.ValidityType   `json:"validity_type" binding:"required,min=1,max=2"`
	ValidityDays    int                   `json:"validity_days" binding:"min=0"`
	DateRange       *models.DateRange     `json:"date_range"`
	Status          models.CouponStatus   `json:"status" binding:"min=0,max=3"`
	TotalCount      int                   `json:"total_count" binding:"min=0"`
}

// ClaimCouponRequest 领取优惠券请求
type ClaimCouponRequest struct {
	CouponID uint `json:"coupon_id" binding:"required,min=1"`
}

// UseCouponRequest 使用优惠券请求
type UseCouponRequest struct {
	OrderAmount float64 `json:"order_amount" binding:"required,min=0"`
}

// DistributeCouponRequest 分发优惠券请求
type DistributeCouponRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required,min=1"`
}

// CouponController 优惠券控制器
type CouponController struct {
	couponService *services.CouponService
}

// NewCouponController 创建优惠券控制器
func NewCouponController() *CouponController {
	return &CouponController{
		couponService: services.NewCouponService(),
	}
}

// List 获取优惠券列表
func (cc *CouponController) List(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	coupons, total, err := cc.couponService.List(&req)
	if err != nil {
		utils.InternalServerError(c, "获取优惠券列表失败")
		return
	}

	utils.PagedSuccess(c, coupons, total, req.GetPage(), req.GetSize())
}

// GetByID 获取优惠券详情
func (cc *CouponController) GetByID(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	coupon, err := cc.couponService.GetByID(req.ID)
	if err != nil {
		if err.Error() == "优惠券不存在" {
			utils.NotFound(c, "优惠券不存在")
		} else {
			utils.InternalServerError(c, "获取优惠券详情失败")
		}
		return
	}

	utils.Success(c, coupon)
}

// Create 创建优惠券
func (cc *CouponController) Create(c *gin.Context) {
	var req CouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 验证日期范围
	if req.ValidityType == models.ValidityTypeRange && req.DateRange == nil {
		utils.BadRequest(c, "选择日期范围类型时，必须提供有效日期范围")
		return
	}

	if req.ValidityType == models.ValidityTypeDays && req.ValidityDays <= 0 {
		utils.BadRequest(c, "选择天数类型时，有效天数必须大于0")
		return
	}

	coupon := &models.Coupon{
		Name:            req.Name,
		Description:     req.Description,
		IsNewUser:       req.IsNewUser,
		Type:            req.Type,
		DiscountPercent: req.DiscountPercent,
		MinAmount:       req.MinAmount,
		MaxAmount:       req.MaxAmount,
		ValidityType:    req.ValidityType,
		ValidityDays:    req.ValidityDays,
		DateRange:       req.DateRange,
		Status:          req.Status,
		TotalCount:      req.TotalCount,
	}

	if err := cc.couponService.Create(coupon); err != nil {
		utils.InternalServerError(c, "创建优惠券失败")
		return
	}

	utils.Created(c, coupon)
}

// Update 更新优惠券
func (cc *CouponController) Update(c *gin.Context) {
	var uriReq types.IDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.ValidateError(c, err)
		return
	}

	var req CouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 检查优惠券是否存在
	coupon, err := cc.couponService.GetByID(uriReq.ID)
	if err != nil {
		if err.Error() == "优惠券不存在" {
			utils.NotFound(c, "优惠券不存在")
		} else {
			utils.InternalServerError(c, "获取优惠券信息失败")
		}
		return
	}

	// 验证日期范围
	if req.ValidityType == models.ValidityTypeRange && req.DateRange == nil {
		utils.BadRequest(c, "选择日期范围类型时，必须提供有效日期范围")
		return
	}

	// 更新字段
	coupon.Name = req.Name
	coupon.Description = req.Description
	coupon.IsNewUser = req.IsNewUser
	coupon.Type = req.Type
	coupon.DiscountPercent = req.DiscountPercent
	coupon.MinAmount = req.MinAmount
	coupon.MaxAmount = req.MaxAmount
	coupon.ValidityType = req.ValidityType
	coupon.ValidityDays = req.ValidityDays
	coupon.DateRange = req.DateRange
	coupon.Status = req.Status
	coupon.TotalCount = req.TotalCount

	if err := cc.couponService.Update(coupon); err != nil {
		utils.InternalServerError(c, "更新优惠券失败")
		return
	}

	utils.Updated(c, coupon)
}

// Delete 删除优惠券
func (cc *CouponController) Delete(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := cc.couponService.Delete(req.ID); err != nil {
		if err.Error() == "优惠券不存在" {
			utils.NotFound(c, "优惠券不存在")
		} else {
			utils.InternalServerError(c, "删除优惠券失败")
		}
		return
	}

	utils.Deleted(c)
}

// Distribute 分发优惠券
func (cc *CouponController) Distribute(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的优惠券ID")
		return
	}

	var req DistributeCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	result, err := cc.couponService.DistributeCoupon(uint(id), req.UserIDs)
	if err != nil {
		if err.Error() == "优惠券不存在" {
			utils.NotFound(c, "优惠券不存在")
		} else {
			utils.BadRequest(c, err.Error())
		}
		return
	}

	utils.SuccessWithMessage(c, "优惠券分发成功", result)
}

// GetUserCoupons 获取用户优惠券列表
func (cc *CouponController) GetUserCoupons(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 获取用户ID（通常从JWT token中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	// 获取状态筛选
	statusStr := c.Query("coupon_status")
	var couponStatus *models.CouponStatus
	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			cs := models.CouponStatus(status)
			couponStatus = &cs
		}
	}

	userCoupons, total, err := cc.couponService.GetUserCoupons(userID.(uint), &req, couponStatus)
	if err != nil {
		utils.InternalServerError(c, "获取用户优惠券失败")
		return
	}

	utils.PagedSuccess(c, userCoupons, total, req.GetPage(), req.GetSize())
}

// ClaimCoupon 领取优惠券
func (cc *CouponController) ClaimCoupon(c *gin.Context) {
	var req ClaimCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	userCoupon, err := cc.couponService.ClaimCoupon(userID.(uint), req.CouponID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "优惠券领取成功", userCoupon)
}

// UseCoupon 使用优惠券
func (cc *CouponController) UseCoupon(c *gin.Context) {
	var uriReq types.IDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.ValidateError(c, err)
		return
	}

	var req UseCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	discountAmount, err := cc.couponService.UseCoupon(userID.(uint), uriReq.ID, req.OrderAmount)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	result := map[string]interface{}{
		"discount_amount": discountAmount,
		"final_amount":    req.OrderAmount - discountAmount,
		"used_at":         time.Now(),
	}

	utils.SuccessWithMessage(c, "优惠券使用成功", result)
}

// GetAvailableCoupons 获取可用优惠券
func (cc *CouponController) GetAvailableCoupons(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	// 获取订单金额
	orderAmountStr := c.Query("order_amount")
	var orderAmount float64
	if orderAmountStr != "" {
		if amount, err := strconv.ParseFloat(orderAmountStr, 64); err == nil {
			orderAmount = amount
		}
	}

	coupons, err := cc.couponService.GetAvailableCoupons(userID.(uint), orderAmount)
	if err != nil {
		utils.InternalServerError(c, "获取可用优惠券失败")
		return
	}

	utils.Success(c, coupons)
}

// GetStatistics 获取优惠券统计
func (cc *CouponController) GetStatistics(c *gin.Context) {
	stats, err := cc.couponService.GetStatistics()
	if err != nil {
		utils.InternalServerError(c, "获取优惠券统计失败")
		return
	}

	utils.Success(c, stats)
}