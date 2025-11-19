package services

import (
	"backend/models"
	"backend/repositories"
	"backend/types"
)

// CustomerService 客户服务
type CustomerService struct {
	customerRepo    *repositories.CustomerRepository
	transactionRepo *repositories.TransactionRepository
	userCouponRepo  *repositories.UserCouponRepository
}

// NewCustomerService 创建客户服务
func NewCustomerService() *CustomerService {
	return &CustomerService{
		customerRepo:    repositories.NewCustomerRepository(),
		transactionRepo: repositories.NewTransactionRepository(),
		userCouponRepo:  repositories.NewUserCouponRepository(),
	}
}

// List 获取客户列表
func (cs *CustomerService) List(req *types.FilterRequest) ([]*models.Customer, int64, error) {
	return cs.customerRepo.List(req)
}

// GetByID 根据ID获取客户
func (cs *CustomerService) GetByID(id uint) (*models.Customer, error) {
	return cs.customerRepo.GetByID(id)
}

// GetByEmail 根据邮箱获取客户
func (cs *CustomerService) GetByEmail(email string) (*models.Customer, error) {
	return cs.customerRepo.GetByEmail(email)
}

// Create 创建客户
func (cs *CustomerService) Create(customer *models.Customer) error {
	// 检查邮箱是否已存在
	existingCustomer, _ := cs.customerRepo.GetByEmail(customer.Email)
	if existingCustomer != nil {
		return &ServiceError{
			Code:    400,
			Message: "邮箱已存在",
		}
	}

	return cs.customerRepo.Create(customer)
}

// Update 更新客户
func (cs *CustomerService) Update(customer *models.Customer) error {
	// 检查邮箱是否被其他客户使用
	existingCustomer, _ := cs.customerRepo.GetByEmail(customer.Email)
	if existingCustomer != nil && existingCustomer.ID != customer.ID {
		return &ServiceError{
			Code:    400,
			Message: "邮箱已存在",
		}
	}

	return cs.customerRepo.Update(customer)
}

// Delete 删除客户
func (cs *CustomerService) Delete(id uint) error {
	// 检查客户是否存在未完成的交易
	pendingTransactions, err := cs.transactionRepo.GetPendingByUserID(id)
	if err != nil {
		return err
	}

	if len(pendingTransactions) > 0 {
		return &ServiceError{
			Code:    400,
			Message: "客户存在未完成的交易，无法删除",
		}
	}

	// 检查客户余额
	customer, err := cs.customerRepo.GetByID(id)
	if err != nil {
		return err
	}

	if customer.Balance > 0 {
		return &ServiceError{
			Code:    400,
			Message: "客户账户余额不为零，无法删除",
		}
	}

	return cs.customerRepo.Delete(id)
}

// UpdateStatus 更新客户状态
func (cs *CustomerService) UpdateStatus(id uint, status models.CustomerStatus) error {
	customer, err := cs.customerRepo.GetByID(id)
	if err != nil {
		return err
	}

	customer.Status = status
	return cs.customerRepo.Update(customer)
}

// UpdateBalance 更新客户余额
func (cs *CustomerService) UpdateBalance(id uint, amount float64, reason string) error {
	customer, err := cs.customerRepo.GetByID(id)
	if err != nil {
		return err
	}

	if !customer.IsActive() {
		return &ServiceError{
			Code:    400,
			Message: "客户账户未激活，无法更新余额",
		}
	}

	// 检查余额是否足够（如果是扣款）
	if amount < 0 && customer.Balance < -amount {
		return &ServiceError{
			Code:    400,
			Message: "账户余额不足",
		}
	}

	// 创建交易记录
	transactionType := models.TransactionTypeRecharge
	if amount < 0 {
		transactionType = models.TransactionTypeWithdraw
		amount = -amount // 转为正数存储
	}

	balanceBefore := customer.Balance
	customer.UpdateBalance(amount)
	
	transaction := &models.Transaction{
		UserID:        id,
		Type:          transactionType,
		Amount:        amount,
		Status:        models.TransactionStatusSuccess,
		Description:   reason,
		BalanceBefore: balanceBefore,
		BalanceAfter:  customer.Balance,
	}

	// 更新客户余额
	if err := cs.customerRepo.Update(customer); err != nil {
		return err
	}

	// 创建交易记录
	return cs.transactionRepo.Create(transaction)
}

// GetTransactions 获取客户交易记录
func (cs *CustomerService) GetTransactions(customerID uint, req *types.FilterRequest) ([]*models.Transaction, int64, error) {
	// 检查客户是否存在
	_, err := cs.customerRepo.GetByID(customerID)
	if err != nil {
		return nil, 0, err
	}

	return cs.transactionRepo.GetByUserID(customerID, req)
}

// GetCoupons 获取客户优惠券
func (cs *CustomerService) GetCoupons(customerID uint, req *types.FilterRequest) ([]*models.UserCoupon, int64, error) {
	// 检查客户是否存在
	_, err := cs.customerRepo.GetByID(customerID)
	if err != nil {
		return nil, 0, err
	}

	return cs.userCouponRepo.GetByUserID(customerID, req, nil)
}

// GetStatistics 获取客户统计
func (cs *CustomerService) GetStatistics() (*types.StatisticsResponse, error) {
	return cs.customerRepo.GetStatistics()
}

// BatchUpdateStatus 批量更新状态
func (cs *CustomerService) BatchUpdateStatus(ids []uint, status models.CustomerStatus) error {
	return cs.customerRepo.BatchUpdateStatus(ids, status)
}

// GetActiveCustomers 获取活动客户
func (cs *CustomerService) GetActiveCustomers() ([]*models.Customer, error) {
	return cs.customerRepo.GetByStatus(models.CustomerStatusActive)
}

// GetCustomersByStatus 根据状态获取客户
func (cs *CustomerService) GetCustomersByStatus(status models.CustomerStatus) ([]*models.Customer, error) {
	return cs.customerRepo.GetByStatus(status)
}

// SearchCustomers 搜索客户
func (cs *CustomerService) SearchCustomers(keyword string, limit int) ([]*models.Customer, error) {
	return cs.customerRepo.Search(keyword, limit)
}

// GetCustomerSummary 获取客户摘要信息
func (cs *CustomerService) GetCustomerSummary(id uint) (map[string]interface{}, error) {
	customer, err := cs.customerRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 获取交易统计
	transactionStats, err := cs.transactionRepo.GetStatsByUserID(id)
	if err != nil {
		return nil, err
	}

	// 获取优惠券统计
	couponStats, err := cs.userCouponRepo.GetStatsByUser(id)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"customer":         customer,
		"transaction_stats": transactionStats,
		"coupon_stats":     couponStats,
		"last_login":       customer.LastLoginAt,
		"member_days":      customer.CreatedAt,
	}, nil
}

// RecordLogin 记录客户登录
func (cs *CustomerService) RecordLogin(id uint) error {
	customer, err := cs.customerRepo.GetByID(id)
	if err != nil {
		return err
	}

	customer.RecordLogin()
	return cs.customerRepo.Update(customer)
}

// ValidateCustomerForTransaction 验证客户是否可以进行交易
func (cs *CustomerService) ValidateCustomerForTransaction(id uint, amount float64) error {
	customer, err := cs.customerRepo.GetByID(id)
	if err != nil {
		return err
	}

	if !customer.IsActive() {
		return &ServiceError{
			Code:    400,
			Message: "客户账户未激活",
		}
	}

	if customer.IsBlocked() {
		return &ServiceError{
			Code:    403,
			Message: "客户账户已被阻止",
		}
	}

	if !customer.CanMakeTransaction(amount) {
		return &ServiceError{
			Code:    400,
			Message: "账户余额不足",
		}
	}

	return nil
}

// GetTopCustomers 获取顶级客户（按余额或交易额排序）
func (cs *CustomerService) GetTopCustomers(limit int, orderBy string) ([]*models.Customer, error) {
	return cs.customerRepo.GetTopCustomers(limit, orderBy)
}

// GetRecentCustomers 获取最近注册的客户
func (cs *CustomerService) GetRecentCustomers(limit int) ([]*models.Customer, error) {
	return cs.customerRepo.GetRecentCustomers(limit)
}