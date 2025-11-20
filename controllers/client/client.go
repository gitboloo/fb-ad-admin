package client

import (
	"backend/api"

	"github.com/gin-gonic/gin"
)

// AuthHandler 客户端认证
type AuthHandler struct {
	controller *api.AdminController
}

// NewAuthHandler 创建客户端认证handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		controller: api.NewAdminController(),
	}
}

// Login 客户登录
func (h *AuthHandler) Login(c *gin.Context) {
	h.controller.Login(c)
}

// Logout 客户退出
func (h *AuthHandler) Logout(c *gin.Context) {
	h.controller.Logout(c)
}

// GetProfile 获取客户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	h.controller.GetProfile(c)
}

// UpdatePassword 修改密码
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	h.controller.UpdatePassword(c)
}

// CustomerHandler 客户个人资料
type CustomerHandler struct {
	controller *api.CustomerController
}

// NewCustomerHandler 创建客户handler
func NewCustomerHandler() *CustomerHandler {
	return &CustomerHandler{
		controller: api.NewCustomerController(),
	}
}

// GetProfile 获取客户资料
func (h *CustomerHandler) GetProfile(c *gin.Context) {
	h.controller.GetProfile(c)
}

// UpdateProfile 更新客户资料
func (h *CustomerHandler) UpdateProfile(c *gin.Context) {
	h.controller.UpdateProfile(c)
}

// FinanceHandler 客户端财务
type FinanceHandler struct {
	controller *api.FinanceController
}

// NewFinanceHandler 创建财务handler
func NewFinanceHandler() *FinanceHandler {
	return &FinanceHandler{
		controller: api.NewFinanceController(),
	}
}

// Recharge 充值
func (h *FinanceHandler) Recharge(c *gin.Context) {
	h.controller.Recharge(c)
}

// Withdraw 提现
func (h *FinanceHandler) Withdraw(c *gin.Context) {
	h.controller.Withdraw(c)
}

// GetTransactions 获取交易记录
func (h *FinanceHandler) GetTransactions(c *gin.Context) {
	h.controller.GetTransactions(c)
}

// GetBalance 获取余额
func (h *FinanceHandler) GetBalance(c *gin.Context) {
	h.controller.GetBalance(c)
}

// GetStatistics 获取统计
func (h *FinanceHandler) GetStatistics(c *gin.Context) {
	h.controller.GetStatistics(c)
}

// CouponHandler 客户端优惠券
type CouponHandler struct {
	controller *api.CouponController
}

// NewCouponHandler 创建优惠券handler
func NewCouponHandler() *CouponHandler {
	return &CouponHandler{
		controller: api.NewCouponController(),
	}
}

// GetUserCoupons 获取用户优惠券
func (h *CouponHandler) GetUserCoupons(c *gin.Context) {
	h.controller.GetUserCoupons(c)
}

// ClaimCoupon 领取优惠券
func (h *CouponHandler) ClaimCoupon(c *gin.Context) {
	h.controller.ClaimCoupon(c)
}

// UseCoupon 使用优惠券
func (h *CouponHandler) UseCoupon(c *gin.Context) {
	h.controller.UseCoupon(c)
}

// GetAvailableCoupons 获取可用优惠券
func (h *CouponHandler) GetAvailableCoupons(c *gin.Context) {
	h.controller.GetAvailableCoupons(c)
}

// AuthCodeHandler 客户端授权码
type AuthCodeHandler struct {
	controller *api.AuthCodeController
}

// NewAuthCodeHandler 创建授权码handler
func NewAuthCodeHandler() *AuthCodeHandler {
	return &AuthCodeHandler{
		controller: api.NewAuthCodeController(),
	}
}

// Verify 验证授权码
func (h *AuthCodeHandler) Verify(c *gin.Context) {
	h.controller.Verify(c)
}
