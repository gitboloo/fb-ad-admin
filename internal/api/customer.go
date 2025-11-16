package api

import (
	"strconv"

	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/services"
	"github.com/ad-platform/backend/internal/types"
	"github.com/ad-platform/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// CustomerRequest 客户请求结构
type CustomerRequest struct {
	Name    string                 `json:"name" binding:"required,max=255"`
	Email   string                 `json:"email" binding:"required,email,max=255"`
	Phone   string                 `json:"phone" binding:"max=20"`
	Company string                 `json:"company" binding:"max=255"`
	Status  models.CustomerStatus  `json:"status" binding:"min=0,max=2"`
	Address string                 `json:"address"`
	Notes   string                 `json:"notes"`
	Balance float64                `json:"balance" binding:"min=0"`
}

// CustomerController 客户控制器
type CustomerController struct {
	customerService *services.CustomerService
}

// NewCustomerController 创建客户控制器
func NewCustomerController() *CustomerController {
	return &CustomerController{
		customerService: services.NewCustomerService(),
	}
}

// List 获取客户列表
func (cc *CustomerController) List(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	customers, total, err := cc.customerService.List(&req)
	if err != nil {
		utils.InternalServerError(c, "获取客户列表失败")
		return
	}

	utils.PagedSuccess(c, customers, total, req.GetPage(), req.GetSize())
}

// GetByID 获取客户详情
func (cc *CustomerController) GetByID(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	customer, err := cc.customerService.GetByID(req.ID)
	if err != nil {
		if err.Error() == "客户不存在" {
			utils.NotFound(c, "客户不存在")
		} else {
			utils.InternalServerError(c, "获取客户详情失败")
		}
		return
	}

	utils.Success(c, customer)
}

// Create 创建客户
func (cc *CustomerController) Create(c *gin.Context) {
	var req CustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	customer := &models.Customer{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Company: req.Company,
		Status:  req.Status,
		Address: req.Address,
		Notes:   req.Notes,
		Balance: req.Balance,
	}

	if err := cc.customerService.Create(customer); err != nil {
		if err.Error() == "邮箱已存在" {
			utils.BadRequest(c, "邮箱已存在")
		} else {
			utils.InternalServerError(c, "创建客户失败")
		}
		return
	}

	utils.Created(c, customer)
}

// Update 更新客户
func (cc *CustomerController) Update(c *gin.Context) {
	var uriReq types.IDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.ValidateError(c, err)
		return
	}

	var req CustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 检查客户是否存在
	customer, err := cc.customerService.GetByID(uriReq.ID)
	if err != nil {
		if err.Error() == "客户不存在" {
			utils.NotFound(c, "客户不存在")
		} else {
			utils.InternalServerError(c, "获取客户信息失败")
		}
		return
	}

	// 更新字段
	customer.Name = req.Name
	customer.Email = req.Email
	customer.Phone = req.Phone
	customer.Company = req.Company
	customer.Status = req.Status
	customer.Address = req.Address
	customer.Notes = req.Notes
	customer.Balance = req.Balance

	if err := cc.customerService.Update(customer); err != nil {
		if err.Error() == "邮箱已存在" {
			utils.BadRequest(c, "邮箱已存在")
		} else {
			utils.InternalServerError(c, "更新客户失败")
		}
		return
	}

	utils.Updated(c, customer)
}

// Delete 删除客户
func (cc *CustomerController) Delete(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := cc.customerService.Delete(req.ID); err != nil {
		if err.Error() == "客户不存在" {
			utils.NotFound(c, "客户不存在")
		} else {
			utils.InternalServerError(c, "删除客户失败")
		}
		return
	}

	utils.Deleted(c)
}

// UpdateStatus 更新客户状态
func (cc *CustomerController) UpdateStatus(c *gin.Context) {
	var uriReq types.IDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.ValidateError(c, err)
		return
	}

	var req types.StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := cc.customerService.UpdateStatus(uriReq.ID, models.CustomerStatus(req.Status)); err != nil {
		if err.Error() == "客户不存在" {
			utils.NotFound(c, "客户不存在")
		} else {
			utils.InternalServerError(c, "更新客户状态失败")
		}
		return
	}

	utils.SuccessWithMessage(c, "状态更新成功", nil)
}

// Block 阻止客户
func (cc *CustomerController) Block(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := cc.customerService.UpdateStatus(req.ID, models.CustomerStatusBlocked); err != nil {
		if err.Error() == "客户不存在" {
			utils.NotFound(c, "客户不存在")
		} else {
			utils.InternalServerError(c, "阻止客户失败")
		}
		return
	}

	utils.SuccessWithMessage(c, "客户已被阻止", nil)
}

// Unblock 解除阻止
func (cc *CustomerController) Unblock(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := cc.customerService.UpdateStatus(req.ID, models.CustomerStatusActive); err != nil {
		if err.Error() == "客户不存在" {
			utils.NotFound(c, "客户不存在")
		} else {
			utils.InternalServerError(c, "解除阻止失败")
		}
		return
	}

	utils.SuccessWithMessage(c, "已解除阻止", nil)
}

// GetTransactions 获取客户交易记录
func (cc *CustomerController) GetTransactions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的客户ID")
		return
	}

	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	transactions, total, err := cc.customerService.GetTransactions(uint(id), &req)
	if err != nil {
		if err.Error() == "客户不存在" {
			utils.NotFound(c, "客户不存在")
		} else {
			utils.InternalServerError(c, "获取交易记录失败")
		}
		return
	}

	utils.PagedSuccess(c, transactions, total, req.GetPage(), req.GetSize())
}

// GetCoupons 获取客户优惠券
func (cc *CustomerController) GetCoupons(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的客户ID")
		return
	}

	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	coupons, total, err := cc.customerService.GetCoupons(uint(id), &req)
	if err != nil {
		if err.Error() == "客户不存在" {
			utils.NotFound(c, "客户不存在")
		} else {
			utils.InternalServerError(c, "获取优惠券失败")
		}
		return
	}

	utils.PagedSuccess(c, coupons, total, req.GetPage(), req.GetSize())
}

// UpdateBalance 更新客户余额
func (cc *CustomerController) UpdateBalance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的客户ID")
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required"`
		Reason string  `json:"reason" binding:"required,max=500"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := cc.customerService.UpdateBalance(uint(id), req.Amount, req.Reason); err != nil {
		if err.Error() == "客户不存在" {
			utils.NotFound(c, "客户不存在")
		} else {
			utils.BadRequest(c, err.Error())
		}
		return
	}

	utils.SuccessWithMessage(c, "余额更新成功", nil)
}

// GetStatistics 获取客户统计
func (cc *CustomerController) GetStatistics(c *gin.Context) {
	stats, err := cc.customerService.GetStatistics()
	if err != nil {
		utils.InternalServerError(c, "获取客户统计失败")
		return
	}

	utils.Success(c, stats)
}

// Export 导出客户数据
func (cc *CustomerController) Export(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 设置导出数量限制
	req.Size = 10000

	customers, _, err := cc.customerService.List(&req)
	if err != nil {
		utils.InternalServerError(c, "导出客户数据失败")
		return
	}

	// 这里可以生成Excel或CSV文件
	// 暂时返回JSON数据
	utils.Success(c, map[string]interface{}{
		"customers": customers,
		"count":     len(customers),
		"exported_at": "now",
	})
}

// BatchUpdateStatus 批量更新状态
func (cc *CustomerController) BatchUpdateStatus(c *gin.Context) {
	var req struct {
		IDs    []uint                `json:"ids" binding:"required,min=1"`
		Status models.CustomerStatus `json:"status" binding:"min=0,max=2"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := cc.customerService.BatchUpdateStatus(req.IDs, req.Status); err != nil {
		utils.InternalServerError(c, "批量更新状态失败")
		return
	}

	utils.SuccessWithMessage(c, "批量更新成功", nil)
}

// GetProfile 获取客户资料（供客户自己使用）
func (cc *CustomerController) GetProfile(c *gin.Context) {
	// 获取当前用户ID（从JWT token中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	customer, err := cc.customerService.GetByID(userID.(uint))
	if err != nil {
		utils.NotFound(c, "用户信息不存在")
		return
	}

	utils.Success(c, customer)
}

// UpdateProfile 更新客户资料（供客户自己使用）
func (cc *CustomerController) UpdateProfile(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	var req struct {
		Name    string `json:"name" binding:"required,max=255"`
		Phone   string `json:"phone" binding:"max=20"`
		Company string `json:"company" binding:"max=255"`
		Address string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	customer, err := cc.customerService.GetByID(userID.(uint))
	if err != nil {
		utils.NotFound(c, "用户信息不存在")
		return
	}

	// 更新允许客户自己修改的字段
	customer.Name = req.Name
	customer.Phone = req.Phone
	customer.Company = req.Company
	customer.Address = req.Address

	if err := cc.customerService.Update(customer); err != nil {
		utils.InternalServerError(c, "更新资料失败")
		return
	}

	utils.Updated(c, customer)
}