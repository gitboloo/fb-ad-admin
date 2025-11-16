package repositories

import (
	"fmt"
	"strings"

	"github.com/ad-platform/backend/internal/database"
	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/types"
	"gorm.io/gorm"
)

// CustomerRepository 客户仓库
type CustomerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository 创建客户仓库
func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{
		db: database.DB,
	}
}

// List 获取客户列表
func (cr *CustomerRepository) List(req *types.FilterRequest) ([]*models.Customer, int64, error) {
	var customers []*models.Customer
	var total int64

	query := cr.db.Model(&models.Customer{})

	// 搜索条件
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("name LIKE ? OR email LIKE ? OR phone LIKE ? OR company LIKE ?", 
			searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
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
		Find(&customers).Error; err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

// GetByID 根据ID获取客户
func (cr *CustomerRepository) GetByID(id uint) (*models.Customer, error) {
	var customer models.Customer
	if err := cr.db.First(&customer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("客户不存在")
		}
		return nil, err
	}
	return &customer, nil
}

// GetByEmail 根据邮箱获取客户
func (cr *CustomerRepository) GetByEmail(email string) (*models.Customer, error) {
	var customer models.Customer
	if err := cr.db.Where("email = ?", email).First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &customer, nil
}

// Create 创建客户
func (cr *CustomerRepository) Create(customer *models.Customer) error {
	return cr.db.Create(customer).Error
}

// Update 更新客户
func (cr *CustomerRepository) Update(customer *models.Customer) error {
	return cr.db.Save(customer).Error
}

// Delete 删除客户
func (cr *CustomerRepository) Delete(id uint) error {
	result := cr.db.Delete(&models.Customer{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("客户不存在")
	}
	return nil
}

// GetByStatus 根据状态获取客户
func (cr *CustomerRepository) GetByStatus(status models.CustomerStatus) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := cr.db.Where("status = ?", status).Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}

// Search 搜索客户
func (cr *CustomerRepository) Search(keyword string, limit int) ([]*models.Customer, error) {
	var customers []*models.Customer
	searchPattern := "%" + strings.TrimSpace(keyword) + "%"
	
	query := cr.db.Where("name LIKE ? OR email LIKE ? OR phone LIKE ? OR company LIKE ?", 
		searchPattern, searchPattern, searchPattern, searchPattern)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}

// BatchUpdateStatus 批量更新状态
func (cr *CustomerRepository) BatchUpdateStatus(ids []uint, status models.CustomerStatus) error {
	return cr.db.Model(&models.Customer{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// GetStatistics 获取客户统计
func (cr *CustomerRepository) GetStatistics() (*types.StatisticsResponse, error) {
	var total int64
	var active int64
	var inactive int64

	// 总数统计
	if err := cr.db.Model(&models.Customer{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// 活跃客户数
	if err := cr.db.Model(&models.Customer{}).
		Where("status = ?", models.CustomerStatusActive).Count(&active).Error; err != nil {
		return nil, err
	}

	// 非活跃客户数
	if err := cr.db.Model(&models.Customer{}).
		Where("status != ?", models.CustomerStatusActive).Count(&inactive).Error; err != nil {
		return nil, err
	}

	// 按状态统计
	var statusStats []struct {
		Status models.CustomerStatus `json:"status"`
		Count  int64                 `json:"count"`
	}
	if err := cr.db.Model(&models.Customer{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusStats).Error; err != nil {
		return nil, err
	}

	categories := make(map[string]interface{})
	for _, stat := range statusStats {
		var statusName string
		switch stat.Status {
		case models.CustomerStatusActive:
			statusName = "活跃"
		case models.CustomerStatusInactive:
			statusName = "未激活"
		case models.CustomerStatusBlocked:
			statusName = "已阻止"
		default:
			statusName = "未知"
		}
		categories[statusName] = stat.Count
	}

	// 趋势数据（最近7天）
	var trendData []types.TrendData
	if err := cr.db.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as value 
		FROM customers 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) 
		GROUP BY DATE(created_at) 
		ORDER BY date ASC
	`).Scan(&trendData).Error; err != nil {
		return nil, err
	}

	// 计算增长率（与上周同期比较）
	var currentWeek int64
	var lastWeek int64
	
	cr.db.Model(&models.Customer{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").Count(&currentWeek)
	cr.db.Model(&models.Customer{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 14 DAY) AND created_at < DATE_SUB(NOW(), INTERVAL 7 DAY)").
		Count(&lastWeek)

	var growth float64
	if lastWeek > 0 {
		growth = float64(currentWeek-lastWeek) / float64(lastWeek) * 100
	}

	return &types.StatisticsResponse{
		Total:      total,
		Active:     active,
		Inactive:   inactive,
		Growth:     growth,
		TrendData:  trendData,
		Categories: categories,
	}, nil
}

// GetTopCustomers 获取顶级客户
func (cr *CustomerRepository) GetTopCustomers(limit int, orderBy string) ([]*models.Customer, error) {
	var customers []*models.Customer
	
	query := cr.db.Where("status = ?", models.CustomerStatusActive)
	
	switch orderBy {
	case "balance":
		query = query.Order("balance DESC")
	case "created_at":
		query = query.Order("created_at ASC")
	default:
		query = query.Order("balance DESC")
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}

// GetRecentCustomers 获取最近注册的客户
func (cr *CustomerRepository) GetRecentCustomers(limit int) ([]*models.Customer, error) {
	var customers []*models.Customer
	
	query := cr.db.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}

// GetCustomersByDateRange 根据注册日期范围获取客户
func (cr *CustomerRepository) GetCustomersByDateRange(startDate, endDate string) ([]*models.Customer, error) {
	var customers []*models.Customer
	
	if err := cr.db.Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Find(&customers).Error; err != nil {
		return nil, err
	}
	
	return customers, nil
}

// GetCustomersWithBalance 获取有余额的客户
func (cr *CustomerRepository) GetCustomersWithBalance() ([]*models.Customer, error) {
	var customers []*models.Customer
	
	if err := cr.db.Where("balance > 0").
		Order("balance DESC").
		Find(&customers).Error; err != nil {
		return nil, err
	}
	
	return customers, nil
}

// GetCustomerCount 获取客户总数
func (cr *CustomerRepository) GetCustomerCount() (int64, error) {
	var count int64
	if err := cr.db.Model(&models.Customer{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetActiveCustomerCount 获取活跃客户数
func (cr *CustomerRepository) GetActiveCustomerCount() (int64, error) {
	var count int64
	if err := cr.db.Model(&models.Customer{}).
		Where("status = ?", models.CustomerStatusActive).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetTotalBalance 获取所有客户的总余额
func (cr *CustomerRepository) GetTotalBalance() (float64, error) {
	var totalBalance float64
	if err := cr.db.Model(&models.Customer{}).
		Select("COALESCE(SUM(balance), 0)").Row().Scan(&totalBalance); err != nil {
		return 0, err
	}
	return totalBalance, nil
}