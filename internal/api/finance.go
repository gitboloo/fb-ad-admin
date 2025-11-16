package api

import (
	"github.com/ad-platform/backend/internal/services"
	"github.com/ad-platform/backend/internal/types"
	"github.com/ad-platform/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// RechargeRequest 充值请求结构
type RechargeRequest struct {
	Amount        float64 `json:"amount" binding:"required,min=0.01"`
	PaymentMethod string  `json:"payment_method" binding:"required,max=50"`
	PaymentID     string  `json:"payment_id" binding:"max=100"`
	Description   string  `json:"description" binding:"max=500"`
}

// WithdrawRequest 提现请求结构
type WithdrawRequest struct {
	Amount      float64 `json:"amount" binding:"required,min=0.01"`
	BankAccount string  `json:"bank_account" binding:"required,max=100"`
	BankName    string  `json:"bank_name" binding:"required,max=100"`
	AccountName string  `json:"account_name" binding:"required,max=100"`
	Description string  `json:"description" binding:"max=500"`
}

// FinanceController 财务控制器
type FinanceController struct {
	financeService *services.FinanceService
}

// NewFinanceController 创建财务控制器
func NewFinanceController() *FinanceController {
	return &FinanceController{
		financeService: services.NewFinanceService(),
	}
}

// Recharge 充值
func (fc *FinanceController) Recharge(c *gin.Context) {
	var req RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	transaction, err := fc.financeService.Recharge(userID.(uint), req.Amount, req.PaymentMethod, req.PaymentID, req.Description)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "充值申请已提交", transaction)
}

// Withdraw 提现
func (fc *FinanceController) Withdraw(c *gin.Context) {
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	// 构建银行信息描述
	bankInfo := map[string]interface{}{
		"bank_account": req.BankAccount,
		"bank_name":    req.BankName,
		"account_name": req.AccountName,
	}

	transaction, err := fc.financeService.Withdraw(userID.(uint), req.Amount, bankInfo, req.Description)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "提现申请已提交", transaction)
}

// GetTransactions 获取交易记录
func (fc *FinanceController) GetTransactions(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	transactions, total, err := fc.financeService.GetTransactions(userID.(uint), &req)
	if err != nil {
		utils.InternalServerError(c, "获取交易记录失败")
		return
	}

	utils.PagedSuccess(c, transactions, total, req.GetPage(), req.GetSize())
}

// GetBalance 获取余额
func (fc *FinanceController) GetBalance(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	balance, err := fc.financeService.GetBalance(userID.(uint))
	if err != nil {
		utils.InternalServerError(c, "获取余额失败")
		return
	}

	utils.Success(c, map[string]interface{}{
		"balance": balance,
		"user_id": userID,
	})
}

// GetStatistics 财务统计
func (fc *FinanceController) GetStatistics(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	stats, err := fc.financeService.GetUserStatistics(userID.(uint))
	if err != nil {
		utils.InternalServerError(c, "获取财务统计失败")
		return
	}

	utils.Success(c, stats)
}

// AdminGetAllTransactions 管理员获取所有交易记录
func (fc *FinanceController) AdminGetAllTransactions(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	transactions, total, err := fc.financeService.GetAllTransactions(&req)
	if err != nil {
		utils.InternalServerError(c, "获取交易记录失败")
		return
	}

	utils.PagedSuccess(c, transactions, total, req.GetPage(), req.GetSize())
}

// AdminGetStatistics 管理员获取财务统计
func (fc *FinanceController) AdminGetStatistics(c *gin.Context) {
	stats, err := fc.financeService.GetAdminStatistics()
	if err != nil {
		utils.InternalServerError(c, "获取财务统计失败")
		return
	}

	utils.Success(c, stats)
}

// ProcessTransaction 处理交易（管理员）
func (fc *FinanceController) ProcessTransaction(c *gin.Context) {
	var uriReq types.IDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.ValidateError(c, err)
		return
	}

	var req struct {
		Action string `json:"action" binding:"required,oneof=approve reject"`
		Reason string `json:"reason" binding:"max=500"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if req.Action == "approve" {
		err := fc.financeService.ApproveTransaction(uriReq.ID, req.Reason)
		if err != nil {
			utils.BadRequest(c, err.Error())
			return
		}
		utils.SuccessWithMessage(c, "交易已批准", nil)
	} else {
		err := fc.financeService.RejectTransaction(uriReq.ID, req.Reason)
		if err != nil {
			utils.BadRequest(c, err.Error())
			return
		}
		utils.SuccessWithMessage(c, "交易已拒绝", nil)
	}
}

// GetPendingTransactions 获取待处理的交易
func (fc *FinanceController) GetPendingTransactions(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	transactions, total, err := fc.financeService.GetPendingTransactions(&req)
	if err != nil {
		utils.InternalServerError(c, "获取待处理交易失败")
		return
	}

	utils.PagedSuccess(c, transactions, total, req.GetPage(), req.GetSize())
}

// GetTransactionsByType 根据类型获取交易
func (fc *FinanceController) GetTransactionsByType(c *gin.Context) {
	transactionType := c.Param("type")
	
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	transactions, total, err := fc.financeService.GetTransactionsByType(transactionType, &req)
	if err != nil {
		utils.InternalServerError(c, "获取交易记录失败")
		return
	}

	utils.PagedSuccess(c, transactions, total, req.GetPage(), req.GetSize())
}

// GetDashboardStats 获取仪表板统计
func (fc *FinanceController) GetDashboardStats(c *gin.Context) {
	stats, err := fc.financeService.GetDashboardStats()
	if err != nil {
		utils.InternalServerError(c, "获取仪表板统计失败")
		return
	}

	utils.Success(c, stats)
}

// ExportTransactions 导出交易记录
func (fc *FinanceController) ExportTransactions(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 设置导出数量限制
	req.Size = 10000

	transactions, _, err := fc.financeService.GetAllTransactions(&req)
	if err != nil {
		utils.InternalServerError(c, "导出交易记录失败")
		return
	}

	// 这里可以生成Excel或CSV文件
	// 暂时返回JSON数据
	utils.Success(c, map[string]interface{}{
		"transactions": transactions,
		"count":        len(transactions),
		"exported_at":  "now",
	})
}

// BatchProcessTransactions 批量处理交易
func (fc *FinanceController) BatchProcessTransactions(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required,min=1"`
		Action string `json:"action" binding:"required,oneof=approve reject"`
		Reason string `json:"reason" binding:"max=500"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	var successCount int
	var failedCount int

	for _, id := range req.IDs {
		var err error
		if req.Action == "approve" {
			err = fc.financeService.ApproveTransaction(id, req.Reason)
		} else {
			err = fc.financeService.RejectTransaction(id, req.Reason)
		}

		if err != nil {
			failedCount++
		} else {
			successCount++
		}
	}

	utils.Success(c, map[string]interface{}{
		"success_count": successCount,
		"failed_count":  failedCount,
		"total_count":   len(req.IDs),
	})
}