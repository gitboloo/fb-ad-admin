package admin

import (
	"backend/api"

	"github.com/gin-gonic/gin"
)

// ProductHandler 产品管理（包装旧的ProductController）
type ProductHandler struct {
	controller *api.ProductController
}

// NewProductHandler 创建产品管理handler
func NewProductHandler() *ProductHandler {
	return &ProductHandler{
		controller: api.NewProductController(),
	}
}

// List 产品列表
func (h *ProductHandler) List(c *gin.Context) {
	h.controller.List(c)
}

// GetByID 获取产品详情
func (h *ProductHandler) GetByID(c *gin.Context) {
	h.controller.GetByID(c)
}

// Create 创建产品
func (h *ProductHandler) Create(c *gin.Context) {
	h.controller.Create(c)
}

// Update 更新产品
func (h *ProductHandler) Update(c *gin.Context) {
	h.controller.Update(c)
}

// Delete 删除产品
func (h *ProductHandler) Delete(c *gin.Context) {
	h.controller.Delete(c)
}

// UpdateStatus 更新产品状态
func (h *ProductHandler) UpdateStatus(c *gin.Context) {
	h.controller.UpdateStatus(c)
}

// UploadLogo 上传产品Logo
func (h *ProductHandler) UploadLogo(c *gin.Context) {
	h.controller.UploadLogo(c)
}

// UploadImages 上传产品图片
func (h *ProductHandler) UploadImages(c *gin.Context) {
	h.controller.UploadImages(c)
}

// UploadFile 通用文件上传（不需要产品ID）
func (h *ProductHandler) UploadFile(c *gin.Context) {
	h.controller.UploadFile(c)
}

// UploadFiles 通用多文件上传（不需要产品ID）
func (h *ProductHandler) UploadFiles(c *gin.Context) {
	h.controller.UploadFiles(c)
}

// GetStatistics 获取产品统计
func (h *ProductHandler) GetStatistics(c *gin.Context) {
	h.controller.GetStatistics(c)
}

// CampaignHandler 广告计划管理（包装旧的CampaignController）
type CampaignHandler struct {
	controller *api.CampaignController
}

// NewCampaignHandler 创建广告计划管理handler
func NewCampaignHandler() *CampaignHandler {
	return &CampaignHandler{
		controller: api.NewCampaignController(),
	}
}

// List 计划列表
func (h *CampaignHandler) List(c *gin.Context) {
	h.controller.List(c)
}

// GetByID 获取计划详情
func (h *CampaignHandler) GetByID(c *gin.Context) {
	h.controller.GetByID(c)
}

// Create 创建计划
func (h *CampaignHandler) Create(c *gin.Context) {
	h.controller.Create(c)
}

// Update 更新计划
func (h *CampaignHandler) Update(c *gin.Context) {
	h.controller.Update(c)
}

// Delete 删除计划
func (h *CampaignHandler) Delete(c *gin.Context) {
	h.controller.Delete(c)
}

// UploadMainImage 上传主图
func (h *CampaignHandler) UploadMainImage(c *gin.Context) {
	h.controller.UploadMainImage(c)
}

// UploadVideo 上传视频
func (h *CampaignHandler) UploadVideo(c *gin.Context) {
	h.controller.UploadVideo(c)
}

// GetStatistics 获取统计
func (h *CampaignHandler) GetStatistics(c *gin.Context) {
	h.controller.GetStatistics(c)
}

// UpdateStatus 更新状态
func (h *CampaignHandler) UpdateStatus(c *gin.Context) {
	h.controller.UpdateStatus(c)
}

// Pause 暂停
func (h *CampaignHandler) Pause(c *gin.Context) {
	h.controller.Pause(c)
}

// Resume 恢复
func (h *CampaignHandler) Resume(c *gin.Context) {
	h.controller.Resume(c)
}

// CustomerHandler 客户管理（包装旧的CustomerController）
type CustomerHandler struct {
	controller *api.CustomerController
}

// NewCustomerHandler 创建客户管理handler
func NewCustomerHandler() *CustomerHandler {
	return &CustomerHandler{
		controller: api.NewCustomerController(),
	}
}

// List 客户列表
func (h *CustomerHandler) List(c *gin.Context) {
	h.controller.List(c)
}

// GetByID 获取客户详情
func (h *CustomerHandler) GetByID(c *gin.Context) {
	h.controller.GetByID(c)
}

// Create 创建客户
func (h *CustomerHandler) Create(c *gin.Context) {
	h.controller.Create(c)
}

// Update 更新客户
func (h *CustomerHandler) Update(c *gin.Context) {
	h.controller.Update(c)
}

// Delete 删除客户
func (h *CustomerHandler) Delete(c *gin.Context) {
	h.controller.Delete(c)
}

// UpdateStatus 更新状态
func (h *CustomerHandler) UpdateStatus(c *gin.Context) {
	h.controller.UpdateStatus(c)
}

// Block 封禁客户
func (h *CustomerHandler) Block(c *gin.Context) {
	h.controller.Block(c)
}

// Unblock 解封客户
func (h *CustomerHandler) Unblock(c *gin.Context) {
	h.controller.Unblock(c)
}

// GetTransactions 获取交易记录
func (h *CustomerHandler) GetTransactions(c *gin.Context) {
	h.controller.GetTransactions(c)
}

// GetCoupons 获取优惠券
func (h *CustomerHandler) GetCoupons(c *gin.Context) {
	h.controller.GetCoupons(c)
}

// UpdateBalance 更新余额
func (h *CustomerHandler) UpdateBalance(c *gin.Context) {
	h.controller.UpdateBalance(c)
}

// GetStatistics 获取统计
func (h *CustomerHandler) GetStatistics(c *gin.Context) {
	h.controller.GetStatistics(c)
}

// Export 导出客户
func (h *CustomerHandler) Export(c *gin.Context) {
	h.controller.Export(c)
}

// BatchUpdateStatus 批量更新状态
func (h *CustomerHandler) BatchUpdateStatus(c *gin.Context) {
	h.controller.BatchUpdateStatus(c)
}

// CouponHandler 优惠券管理（包装旧的CouponController）
type CouponHandler struct {
	controller *api.CouponController
}

// NewCouponHandler 创建优惠券管理handler
func NewCouponHandler() *CouponHandler {
	return &CouponHandler{
		controller: api.NewCouponController(),
	}
}

// List 优惠券列表
func (h *CouponHandler) List(c *gin.Context) {
	h.controller.List(c)
}

// GetByID 获取优惠券详情
func (h *CouponHandler) GetByID(c *gin.Context) {
	h.controller.GetByID(c)
}

// Create 创建优惠券
func (h *CouponHandler) Create(c *gin.Context) {
	h.controller.Create(c)
}

// Update 更新优惠券
func (h *CouponHandler) Update(c *gin.Context) {
	h.controller.Update(c)
}

// Delete 删除优惠券
func (h *CouponHandler) Delete(c *gin.Context) {
	h.controller.Delete(c)
}

// Distribute 分发优惠券
func (h *CouponHandler) Distribute(c *gin.Context) {
	h.controller.Distribute(c)
}

// GetStatistics 获取统计
func (h *CouponHandler) GetStatistics(c *gin.Context) {
	h.controller.GetStatistics(c)
}

// AuthCodeHandler 授权码管理（包装旧的AuthCodeController）
type AuthCodeHandler struct {
	controller *api.AuthCodeController
}

// NewAuthCodeHandler 创建授权码管理handler
func NewAuthCodeHandler() *AuthCodeHandler {
	return &AuthCodeHandler{
		controller: api.NewAuthCodeController(),
	}
}

// List 授权码列表
func (h *AuthCodeHandler) List(c *gin.Context) {
	h.controller.List(c)
}

// GetByID 获取授权码详情
func (h *AuthCodeHandler) GetByID(c *gin.Context) {
	h.controller.GetByID(c)
}

// Generate 生成授权码
func (h *AuthCodeHandler) Generate(c *gin.Context) {
	h.controller.Generate(c)
}

// Revoke 撤销授权码
func (h *AuthCodeHandler) Revoke(c *gin.Context) {
	h.controller.Revoke(c)
}

// Verify 验证授权码
func (h *AuthCodeHandler) Verify(c *gin.Context) {
	h.controller.Verify(c)
}

// BatchRevoke 批量撤销
func (h *AuthCodeHandler) BatchRevoke(c *gin.Context) {
	h.controller.BatchRevoke(c)
}

// Export 导出授权码
func (h *AuthCodeHandler) Export(c *gin.Context) {
	h.controller.Export(c)
}

// GetStatistics 获取统计
func (h *AuthCodeHandler) GetStatistics(c *gin.Context) {
	h.controller.GetStatistics(c)
}

// GetUsageHistory 获取使用历史
func (h *AuthCodeHandler) GetUsageHistory(c *gin.Context) {
	h.controller.GetUsageHistory(c)
}

// GetExpiredCodes 获取过期授权码
func (h *AuthCodeHandler) GetExpiredCodes(c *gin.Context) {
	h.controller.GetExpiredCodes(c)
}

// CleanExpired 清理过期授权码
func (h *AuthCodeHandler) CleanExpired(c *gin.Context) {
	h.controller.CleanExpired(c)
}

// GetCodeByCode 通过code查询
func (h *AuthCodeHandler) GetCodeByCode(c *gin.Context) {
	h.controller.GetCodeByCode(c)
}

// FinanceHandler 财务管理（包装旧的FinanceController）
type FinanceHandler struct {
	controller *api.FinanceController
}

// NewFinanceHandler 创建财务管理handler
func NewFinanceHandler() *FinanceHandler {
	return &FinanceHandler{
		controller: api.NewFinanceController(),
	}
}

// AdminGetAllTransactions 获取所有交易
func (h *FinanceHandler) AdminGetAllTransactions(c *gin.Context) {
	h.controller.AdminGetAllTransactions(c)
}

// AdminGetStatistics 获取统计
func (h *FinanceHandler) AdminGetStatistics(c *gin.Context) {
	h.controller.AdminGetStatistics(c)
}

// ProcessTransaction 处理交易
func (h *FinanceHandler) ProcessTransaction(c *gin.Context) {
	h.controller.ProcessTransaction(c)
}

// GetPendingTransactions 获取待处理交易
func (h *FinanceHandler) GetPendingTransactions(c *gin.Context) {
	h.controller.GetPendingTransactions(c)
}

// GetTransactionsByType 按类型获取交易
func (h *FinanceHandler) GetTransactionsByType(c *gin.Context) {
	h.controller.GetTransactionsByType(c)
}

// GetDashboardStats 获取仪表盘统计
func (h *FinanceHandler) GetDashboardStats(c *gin.Context) {
	h.controller.GetDashboardStats(c)
}

// ExportTransactions 导出交易
func (h *FinanceHandler) ExportTransactions(c *gin.Context) {
	h.controller.ExportTransactions(c)
}

// BatchProcessTransactions 批量处理交易
func (h *FinanceHandler) BatchProcessTransactions(c *gin.Context) {
	h.controller.BatchProcessTransactions(c)
}

// PermissionHandler 权限管理（包装旧的PermissionController）
type PermissionHandler struct {
	controller *api.PermissionController
}

// NewPermissionHandler 创建权限管理handler
func NewPermissionHandler() *PermissionHandler {
	return &PermissionHandler{
		controller: api.NewPermissionController(),
	}
}

// GetPermissionTree 获取权限树
func (h *PermissionHandler) GetPermissionTree(c *gin.Context) {
	h.controller.GetPermissionTree(c)
}

// GetAllPermissions 获取所有权限
func (h *PermissionHandler) GetAllPermissions(c *gin.Context) {
	h.controller.GetAllPermissions(c)
}

// GetPermissionByID 获取权限详情
func (h *PermissionHandler) GetPermissionByID(c *gin.Context) {
	h.controller.GetPermissionByID(c)
}

// CreatePermission 创建权限
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	h.controller.CreatePermission(c)
}

// UpdatePermission 更新权限
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	h.controller.UpdatePermission(c)
}

// DeletePermission 删除权限
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	h.controller.DeletePermission(c)
}

// StatisticsHandler 统计分析（包装旧的StatisticsController）
type StatisticsHandler struct {
	controller *api.StatisticsController
}

// NewStatisticsHandler 创建统计分析handler
func NewStatisticsHandler() *StatisticsHandler {
	return &StatisticsHandler{
		controller: api.NewStatisticsController(),
	}
}

// GetOverview 获取概览
func (h *StatisticsHandler) GetOverview(c *gin.Context) {
	h.controller.GetOverview(c)
}

// GetProducts 获取产品统计
func (h *StatisticsHandler) GetProducts(c *gin.Context) {
	h.controller.GetProducts(c)
}

// SystemHandler 系统管理（包装旧的SystemController）
type SystemHandler struct {
	controller *api.SystemController
}

// NewSystemHandler 创建系统管理handler
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		controller: api.NewSystemController(),
	}
}

// GetConfigs 获取配置
func (h *SystemHandler) GetConfigs(c *gin.Context) {
	h.controller.GetConfigs(c)
}

// UpdateConfigs 更新配置
func (h *SystemHandler) UpdateConfigs(c *gin.Context) {
	h.controller.UpdateConfigs(c)
}

// GetConfig 获取单个配置
func (h *SystemHandler) GetConfig(c *gin.Context) {
	h.controller.GetConfig(c)
}

// UpdateConfig 更新单个配置
func (h *SystemHandler) UpdateConfig(c *gin.Context) {
	h.controller.UpdateConfig(c)
}

// GetStats 获取统计
func (h *SystemHandler) GetStats(c *gin.Context) {
	h.controller.GetStats(c)
}

// GetDashboard 获取仪表盘
func (h *SystemHandler) GetDashboard(c *gin.Context) {
	h.controller.GetDashboard(c)
}

// GetSystemInfo 获取系统信息
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	h.controller.GetSystemInfo(c)
}

// UpdateSystemInfo 更新系统信息
func (h *SystemHandler) UpdateSystemInfo(c *gin.Context) {
	h.controller.UpdateSystemInfo(c)
}

// GetMaintenanceMode 获取维护模式
func (h *SystemHandler) GetMaintenanceMode(c *gin.Context) {
	h.controller.GetMaintenanceMode(c)
}

// SetMaintenanceMode 设置维护模式
func (h *SystemHandler) SetMaintenanceMode(c *gin.Context) {
	h.controller.SetMaintenanceMode(c)
}

// GetHealth 健康检查
func (h *SystemHandler) GetHealth(c *gin.Context) {
	h.controller.GetHealth(c)
}
