package repositories

import (
	"fmt"
	"strings"

	"backend/database"
	"backend/models"
	"backend/types"
	"gorm.io/gorm"
)

// TransactionRepository 交易仓库
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository 创建交易仓库
func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{
		db: database.DB,
	}
}

// List 获取交易列表
func (tr *TransactionRepository) List(req *types.FilterRequest) ([]*models.Transaction, int64, error) {
	var transactions []*models.Transaction
	var total int64

	query := tr.db.Model(&models.Transaction{})

	// 搜索条件
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("description LIKE ? OR order_no LIKE ? OR payment_id LIKE ?", 
			searchPattern, searchPattern, searchPattern)
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 分类筛选（交易类型）
	if req.Category != "" {
		query = query.Where("type = ?", req.Category)
	}

	// 日期范围筛选
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", req.EndDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序和分页
	orderClause := fmt.Sprintf("%s %s", req.GetSort(), req.GetOrder())
	if err := query.Order(orderClause).
		Offset(req.GetOffset()).
		Limit(req.GetSize()).
		Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetByID 根据ID获取交易
func (tr *TransactionRepository) GetByID(id uint) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := tr.db.First(&transaction, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("交易不存在")
		}
		return nil, err
	}
	return &transaction, nil
}

// GetByOrderNo 根据订单号获取交易
func (tr *TransactionRepository) GetByOrderNo(orderNo string) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := tr.db.Where("order_no = ?", orderNo).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("交易不存在")
		}
		return nil, err
	}
	return &transaction, nil
}

// Create 创建交易
func (tr *TransactionRepository) Create(transaction *models.Transaction) error {
	return tr.db.Create(transaction).Error
}

// Update 更新交易
func (tr *TransactionRepository) Update(transaction *models.Transaction) error {
	return tr.db.Save(transaction).Error
}

// GetByUserID 根据用户ID获取交易列表
func (tr *TransactionRepository) GetByUserID(userID uint, req *types.FilterRequest) ([]*models.Transaction, int64, error) {
	var transactions []*models.Transaction
	var total int64

	query := tr.db.Model(&models.Transaction{}).Where("user_id = ?", userID)

	// 搜索条件
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("description LIKE ? OR order_no LIKE ?", searchPattern, searchPattern)
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 分类筛选（交易类型）
	if req.Category != "" {
		query = query.Where("type = ?", req.Category)
	}

	// 日期范围筛选
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", req.EndDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序和分页
	orderClause := fmt.Sprintf("%s %s", req.GetSort(), req.GetOrder())
	if err := query.Order(orderClause).
		Offset(req.GetOffset()).
		Limit(req.GetSize()).
		Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetByType 根据类型获取交易
func (tr *TransactionRepository) GetByType(transactionType models.TransactionType) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	if err := tr.db.Where("type = ?", transactionType).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetByStatus 根据状态获取交易
func (tr *TransactionRepository) GetByStatus(status models.TransactionStatus) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	if err := tr.db.Where("status = ?", status).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetPendingByUserID 获取用户待处理的交易
func (tr *TransactionRepository) GetPendingByUserID(userID uint) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	if err := tr.db.Where("user_id = ?", userID).
		Where("status IN ?", []models.TransactionStatus{
			models.TransactionStatusPending,
			models.TransactionStatusProcessing,
		}).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// Search 搜索交易
func (tr *TransactionRepository) Search(keyword string, limit int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	searchPattern := "%" + strings.TrimSpace(keyword) + "%"
	
	query := tr.db.Where("description LIKE ? OR order_no LIKE ? OR payment_id LIKE ?", 
		searchPattern, searchPattern, searchPattern)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetStatsByUserID 获取用户交易统计
func (tr *TransactionRepository) GetStatsByUserID(userID uint) (map[string]interface{}, error) {
	var totalTransactions int64
	var totalRecharge, totalWithdraw, totalConsume float64

	// 总交易数
	tr.db.Model(&models.Transaction{}).Where("user_id = ?", userID).Count(&totalTransactions)

	// 充值总额
	tr.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND status = ?", userID, models.TransactionTypeRecharge, models.TransactionStatusSuccess).
		Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalRecharge)

	// 提现总额
	tr.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND status = ?", userID, models.TransactionTypeWithdraw, models.TransactionStatusSuccess).
		Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalWithdraw)

	// 消费总额
	tr.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND status = ?", userID, models.TransactionTypeConsume, models.TransactionStatusSuccess).
		Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalConsume)

	return map[string]interface{}{
		"total_transactions": totalTransactions,
		"total_recharge":     totalRecharge,
		"total_withdraw":     totalWithdraw,
		"total_consume":      totalConsume,
		"net_amount":         totalRecharge - totalWithdraw - totalConsume,
	}, nil
}

// GetStatistics 获取交易统计
func (tr *TransactionRepository) GetStatistics() (*types.StatisticsResponse, error) {
	var total int64
	var success int64
	var pending int64

	// 总交易数
	if err := tr.db.Model(&models.Transaction{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// 成功交易数
	if err := tr.db.Model(&models.Transaction{}).
		Where("status = ?", models.TransactionStatusSuccess).Count(&success).Error; err != nil {
		return nil, err
	}

	// 待处理交易数
	if err := tr.db.Model(&models.Transaction{}).
		Where("status IN ?", []models.TransactionStatus{
			models.TransactionStatusPending,
			models.TransactionStatusProcessing,
		}).Count(&pending).Error; err != nil {
		return nil, err
	}

	// 按类型统计
	var typeStats []struct {
		Type  models.TransactionType `json:"type"`
		Count int64                  `json:"count"`
	}
	if err := tr.db.Model(&models.Transaction{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Find(&typeStats).Error; err != nil {
		return nil, err
	}

	categories := make(map[string]interface{})
	for _, stat := range typeStats {
		var typeName string
		switch stat.Type {
		case models.TransactionTypeRecharge:
			typeName = "充值"
		case models.TransactionTypeWithdraw:
			typeName = "提现"
		case models.TransactionTypeConsume:
			typeName = "消费"
		case models.TransactionTypeRefund:
			typeName = "退款"
		case models.TransactionTypeReward:
			typeName = "奖励"
		default:
			typeName = "其他"
		}
		categories[typeName] = stat.Count
	}

	// 趋势数据（最近7天交易额）
	var trendData []types.TrendData
	if err := tr.db.Raw(`
		SELECT DATE(created_at) as date, COALESCE(SUM(amount), 0) as value 
		FROM transactions 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) 
		AND status = ?
		GROUP BY DATE(created_at) 
		ORDER BY date ASC
	`, models.TransactionStatusSuccess).Scan(&trendData).Error; err != nil {
		return nil, err
	}

	// 计算增长率（与上周同期比较）
	var currentWeek int64
	var lastWeek int64
	
	tr.db.Model(&models.Transaction{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").Count(&currentWeek)
	tr.db.Model(&models.Transaction{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 14 DAY) AND created_at < DATE_SUB(NOW(), INTERVAL 7 DAY)").
		Count(&lastWeek)

	var growth float64
	if lastWeek > 0 {
		growth = float64(currentWeek-lastWeek) / float64(lastWeek) * 100
	}

	return &types.StatisticsResponse{
		Total:      total,
		Active:     success,
		Inactive:   pending,
		Growth:     growth,
		TrendData:  trendData,
		Categories: categories,
	}, nil
}

// GetTotalAmount 获取总交易金额
func (tr *TransactionRepository) GetTotalAmount() (float64, error) {
	var totalAmount float64
	if err := tr.db.Model(&models.Transaction{}).
		Where("status = ?", models.TransactionStatusSuccess).
		Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalAmount); err != nil {
		return 0, err
	}
	return totalAmount, nil
}

// GetAmountByType 根据类型获取交易金额
func (tr *TransactionRepository) GetAmountByType(transactionType models.TransactionType) (float64, error) {
	var amount float64
	if err := tr.db.Model(&models.Transaction{}).
		Where("type = ? AND status = ?", transactionType, models.TransactionStatusSuccess).
		Select("COALESCE(SUM(amount), 0)").Row().Scan(&amount); err != nil {
		return 0, err
	}
	return amount, nil
}

// GetRecentTransactions 获取最近的交易
func (tr *TransactionRepository) GetRecentTransactions(limit int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	
	query := tr.db.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetLargeTransactions 获取大额交易
func (tr *TransactionRepository) GetLargeTransactions(minAmount float64, limit int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	
	query := tr.db.Where("amount >= ?", minAmount).Order("amount DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}