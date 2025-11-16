package services

import (
	"encoding/json"
	"fmt"

	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/repositories"
	"github.com/ad-platform/backend/internal/types"
)

// FinanceService 财务服务
type FinanceService struct {
	transactionRepo *repositories.TransactionRepository
	customerRepo    *repositories.CustomerRepository
}

// NewFinanceService 创建财务服务
func NewFinanceService() *FinanceService {
	return &FinanceService{
		transactionRepo: repositories.NewTransactionRepository(),
		customerRepo:    repositories.NewCustomerRepository(),
	}
}

// Recharge 充值
func (fs *FinanceService) Recharge(userID uint, amount float64, paymentMethod, paymentID, description string) (*models.Transaction, error) {
	// 验证用户是否存在且可用
	customer, err := fs.customerRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if !customer.IsActive() {
		return nil, &ServiceError{
			Code:    400,
			Message: "账户未激活，无法充值",
		}
	}

	if customer.IsBlocked() {
		return nil, &ServiceError{
			Code:    403,
			Message: "账户已被阻止，无法充值",
		}
	}

	// 验证充值金额
	if amount <= 0 {
		return nil, &ServiceError{
			Code:    400,
			Message: "充值金额必须大于0",
		}
	}

	// 创建充值交易记录
	transaction := &models.Transaction{
		UserID:        userID,
		Type:          models.TransactionTypeRecharge,
		Amount:        amount,
		Status:        models.TransactionStatusPending,
		Description:   description,
		PaymentMethod: paymentMethod,
		PaymentID:     paymentID,
		BalanceBefore: customer.Balance,
	}

	if err := fs.transactionRepo.Create(transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

// Withdraw 提现
func (fs *FinanceService) Withdraw(userID uint, amount float64, bankInfo map[string]interface{}, description string) (*models.Transaction, error) {
	// 验证用户是否存在且可用
	customer, err := fs.customerRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if !customer.IsActive() {
		return nil, &ServiceError{
			Code:    400,
			Message: "账户未激活，无法提现",
		}
	}

	if customer.IsBlocked() {
		return nil, &ServiceError{
			Code:    403,
			Message: "账户已被阻止，无法提现",
		}
	}

	// 验证提现金额
	if amount <= 0 {
		return nil, &ServiceError{
			Code:    400,
			Message: "提现金额必须大于0",
		}
	}

	// 检查余额是否足够
	if customer.Balance < amount {
		return nil, &ServiceError{
			Code:    400,
			Message: fmt.Sprintf("余额不足，当前余额：%.2f", customer.Balance),
		}
	}

	// 构建银行信息描述
	bankInfoJSON, _ := json.Marshal(bankInfo)
	bankInfoDesc := string(bankInfoJSON)
	if description != "" {
		description = description + " | 银行信息：" + bankInfoDesc
	} else {
		description = "银行信息：" + bankInfoDesc
	}

	// 创建提现交易记录
	transaction := &models.Transaction{
		UserID:        userID,
		Type:          models.TransactionTypeWithdraw,
		Amount:        amount,
		Status:        models.TransactionStatusPending,
		Description:   description,
		BalanceBefore: customer.Balance,
	}

	if err := fs.transactionRepo.Create(transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetTransactions 获取用户交易记录
func (fs *FinanceService) GetTransactions(userID uint, req *types.FilterRequest) ([]*models.Transaction, int64, error) {
	return fs.transactionRepo.GetByUserID(userID, req)
}

// GetAllTransactions 获取所有交易记录（管理员）
func (fs *FinanceService) GetAllTransactions(req *types.FilterRequest) ([]*models.Transaction, int64, error) {
	return fs.transactionRepo.List(req)
}

// GetBalance 获取用户余额
func (fs *FinanceService) GetBalance(userID uint) (float64, error) {
	customer, err := fs.customerRepo.GetByID(userID)
	if err != nil {
		return 0, err
	}
	return customer.Balance, nil
}

// GetUserStatistics 获取用户财务统计
func (fs *FinanceService) GetUserStatistics(userID uint) (map[string]interface{}, error) {
	customer, err := fs.customerRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	stats, err := fs.transactionRepo.GetStatsByUserID(userID)
	if err != nil {
		return nil, err
	}

	stats["current_balance"] = customer.Balance
	return stats, nil
}

// GetAdminStatistics 获取管理员财务统计
func (fs *FinanceService) GetAdminStatistics() (map[string]interface{}, error) {
	// 获取总交易金额
	totalAmount, err := fs.transactionRepo.GetTotalAmount()
	if err != nil {
		return nil, err
	}

	// 获取各类型交易金额
	rechargeAmount, _ := fs.transactionRepo.GetAmountByType(models.TransactionTypeRecharge)
	withdrawAmount, _ := fs.transactionRepo.GetAmountByType(models.TransactionTypeWithdraw)
	consumeAmount, _ := fs.transactionRepo.GetAmountByType(models.TransactionTypeConsume)
	refundAmount, _ := fs.transactionRepo.GetAmountByType(models.TransactionTypeRefund)

	// 获取客户总余额
	totalBalance, _ := fs.customerRepo.GetTotalBalance()

	// 获取交易统计
	transactionStats, err := fs.transactionRepo.GetStatistics()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_amount":        totalAmount,
		"recharge_amount":     rechargeAmount,
		"withdraw_amount":     withdrawAmount,
		"consume_amount":      consumeAmount,
		"refund_amount":       refundAmount,
		"total_balance":       totalBalance,
		"transaction_stats":   transactionStats,
		"net_cash_flow":       rechargeAmount - withdrawAmount,
		"platform_revenue":    consumeAmount - refundAmount,
	}, nil
}

// ApproveTransaction 批准交易
func (fs *FinanceService) ApproveTransaction(transactionID uint, reason string) error {
	transaction, err := fs.transactionRepo.GetByID(transactionID)
	if err != nil {
		return err
	}

	if transaction.Status != models.TransactionStatusPending {
		return &ServiceError{
			Code:    400,
			Message: "只能批准待处理的交易",
		}
	}

	customer, err := fs.customerRepo.GetByID(transaction.UserID)
	if err != nil {
		return err
	}

	// 根据交易类型处理余额
	switch transaction.Type {
	case models.TransactionTypeRecharge:
		// 充值：增加余额
		customer.UpdateBalance(transaction.Amount)
		transaction.BalanceAfter = customer.Balance
		
	case models.TransactionTypeWithdraw:
		// 提现：减少余额
		if customer.Balance < transaction.Amount {
			return &ServiceError{
				Code:    400,
				Message: "用户余额不足，无法完成提现",
			}
		}
		customer.UpdateBalance(-transaction.Amount)
		transaction.BalanceAfter = customer.Balance
		
	case models.TransactionTypeRefund:
		// 退款：增加余额
		customer.UpdateBalance(transaction.Amount)
		transaction.BalanceAfter = customer.Balance
	}

	// 更新交易状态
	transaction.Complete(customer.Balance)
	if reason != "" {
		transaction.Description += " | 处理备注：" + reason
	}

	// 保存更新
	if err := fs.customerRepo.Update(customer); err != nil {
		return err
	}

	return fs.transactionRepo.Update(transaction)
}

// RejectTransaction 拒绝交易
func (fs *FinanceService) RejectTransaction(transactionID uint, reason string) error {
	transaction, err := fs.transactionRepo.GetByID(transactionID)
	if err != nil {
		return err
	}

	if transaction.Status != models.TransactionStatusPending {
		return &ServiceError{
			Code:    400,
			Message: "只能拒绝待处理的交易",
		}
	}

	// 更新交易状态为失败
	transaction.Fail()
	if reason != "" {
		transaction.Description += " | 拒绝原因：" + reason
	}

	return fs.transactionRepo.Update(transaction)
}

// GetPendingTransactions 获取待处理的交易
func (fs *FinanceService) GetPendingTransactions(req *types.FilterRequest) ([]*models.Transaction, int64, error) {
	// 修改请求以只获取待处理的交易
	status := int(models.TransactionStatusPending)
	req.Status = &status
	
	return fs.transactionRepo.List(req)
}

// GetTransactionsByType 根据类型获取交易
func (fs *FinanceService) GetTransactionsByType(transactionType string, req *types.FilterRequest) ([]*models.Transaction, int64, error) {
	// 设置分类筛选
	req.Category = transactionType
	
	return fs.transactionRepo.List(req)
}

// GetDashboardStats 获取仪表板统计
func (fs *FinanceService) GetDashboardStats() (map[string]interface{}, error) {
	// 今日交易统计
	todayStats := map[string]interface{}{
		"total_transactions": 0,
		"total_amount":       0.0,
		"recharge_amount":    0.0,
		"withdraw_amount":    0.0,
	}

	// 待处理交易数量
	var pendingCount int64
	pendingTransactions, err := fs.transactionRepo.GetByStatus(models.TransactionStatusPending)
	if err != nil {
		return nil, err
	}
	pendingCount = int64(len(pendingTransactions))

	// 总体统计
	totalAmount, _ := fs.transactionRepo.GetTotalAmount()
	totalBalance, _ := fs.customerRepo.GetTotalBalance()

	return map[string]interface{}{
		"today_stats":        todayStats,
		"pending_count":      pendingCount,
		"total_amount":       totalAmount,
		"total_balance":      totalBalance,
		"pending_transactions": pendingTransactions[:min(len(pendingTransactions), 10)], // 最新10条
	}, nil
}

// CreateConsumptionTransaction 创建消费交易
func (fs *FinanceService) CreateConsumptionTransaction(userID uint, amount float64, description string) error {
	customer, err := fs.customerRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if !customer.CanMakeTransaction(amount) {
		return &ServiceError{
			Code:    400,
			Message: "余额不足",
		}
	}

	// 创建消费交易
	transaction := &models.Transaction{
		UserID:        userID,
		Type:          models.TransactionTypeConsume,
		Amount:        amount,
		Status:        models.TransactionStatusSuccess,
		Description:   description,
		BalanceBefore: customer.Balance,
	}

	// 扣减余额
	customer.UpdateBalance(-amount)
	transaction.BalanceAfter = customer.Balance
	transaction.Complete(customer.Balance)

	// 保存更新
	if err := fs.customerRepo.Update(customer); err != nil {
		return err
	}

	return fs.transactionRepo.Create(transaction)
}

// CreateRefundTransaction 创建退款交易
func (fs *FinanceService) CreateRefundTransaction(userID uint, amount float64, description string) error {
	customer, err := fs.customerRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// 创建退款交易
	transaction := &models.Transaction{
		UserID:        userID,
		Type:          models.TransactionTypeRefund,
		Amount:        amount,
		Status:        models.TransactionStatusSuccess,
		Description:   description,
		BalanceBefore: customer.Balance,
	}

	// 增加余额
	customer.UpdateBalance(amount)
	transaction.BalanceAfter = customer.Balance
	transaction.Complete(customer.Balance)

	// 保存更新
	if err := fs.customerRepo.Update(customer); err != nil {
		return err
	}

	return fs.transactionRepo.Create(transaction)
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}