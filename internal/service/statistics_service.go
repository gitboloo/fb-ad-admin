package service

import (
	"github.com/ad-platform/backend/internal/database"
	"github.com/ad-platform/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

// StatisticsService 统计服务
type StatisticsService struct {
	db *gorm.DB
}

// NewStatisticsService 创建统计服务实例
func NewStatisticsService() *StatisticsService {
	return &StatisticsService{
		db: database.GetDB(),
	}
}

// GetTotalProducts 获取产品总数
func (s *StatisticsService) GetTotalProducts() int64 {
	var count int64
	s.db.Model(&models.Product{}).Count(&count)
	return count
}

// GetActiveCampaigns 获取活跃计划数
func (s *StatisticsService) GetActiveCampaigns() int64 {
	var count int64
	s.db.Model(&models.Campaign{}).Where("status = ?", "running").Count(&count)
	return count
}

// GetTotalCustomers 获取客户总数
func (s *StatisticsService) GetTotalCustomers() int64 {
	var count int64
	s.db.Model(&models.Customer{}).Count(&count)
	return count
}

// GetTotalRevenue 获取总收入
func (s *StatisticsService) GetTotalRevenue(startDate, endDate string) float64 {
	var total float64
	query := s.db.Model(&models.Transaction{}).Where("status = ?", models.TransactionStatusSuccess)
	
	if startDate != "" && endDate != "" {
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate+" 23:59:59")
	}
	
	query.Where("type = ?", models.TransactionTypeRecharge).Pluck("SUM(amount)", &total)
	return total
}

// GetTodayRevenue 获取今日收入
func (s *StatisticsService) GetTodayRevenue() float64 {
	today := time.Now().Format("2006-01-02")
	return s.GetTotalRevenue(today, today)
}

// GetMonthlyGrowth 获取月度增长率
func (s *StatisticsService) GetMonthlyGrowth() float64 {
	// 本月收入
	now := time.Now()
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)
	thisMonth := s.GetTotalRevenue(firstDayOfMonth.Format("2006-01-02"), lastDayOfMonth.Format("2006-01-02"))
	
	// 上月收入
	firstDayOfLastMonth := firstDayOfMonth.AddDate(0, -1, 0)
	lastDayOfLastMonth := firstDayOfMonth.AddDate(0, 0, -1)
	lastMonth := s.GetTotalRevenue(firstDayOfLastMonth.Format("2006-01-02"), lastDayOfLastMonth.Format("2006-01-02"))
	
	if lastMonth == 0 {
		return 0
	}
	
	return ((thisMonth - lastMonth) / lastMonth) * 100
}

// GetRevenueChart 获取收入图表数据
func (s *StatisticsService) GetRevenueChart(startDate, endDate string) []gin.H {
	var results []gin.H
	
	query := s.db.Model(&models.Transaction{}).
		Select("DATE(created_at) as date, SUM(amount) as total").
		Where("status = ? AND type = ?", models.TransactionStatusSuccess, models.TransactionTypeRecharge)
	
	if startDate != "" && endDate != "" {
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate+" 23:59:59")
	}
	
	rows, _ := query.Group("DATE(created_at)").Order("date ASC").Rows()
	defer rows.Close()
	
	for rows.Next() {
		var date string
		var total float64
		rows.Scan(&date, &total)
		results = append(results, gin.H{"date": date, "total": total})
	}
	
	return results
}

// GetUserChart 获取用户增长图表数据
func (s *StatisticsService) GetUserChart(startDate, endDate string) []gin.H {
	var results []gin.H
	
	query := s.db.Model(&models.Customer{}).
		Select("DATE(created_at) as date, COUNT(*) as count")
	
	if startDate != "" && endDate != "" {
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate+" 23:59:59")
	}
	
	rows, _ := query.Group("DATE(created_at)").Order("date ASC").Rows()
	defer rows.Close()
	
	for rows.Next() {
		var date string
		var count int
		rows.Scan(&date, &count)
		results = append(results, gin.H{"date": date, "count": count})
	}
	
	return results
}

// GetProductTypeChart 获取产品类型分布
func (s *StatisticsService) GetProductTypeChart() []gin.H {
	var results []gin.H
	
	rows, _ := s.db.Model(&models.Product{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Rows()
	defer rows.Close()
	
	for rows.Next() {
		var productType string
		var count int
		rows.Scan(&productType, &count)
		results = append(results, gin.H{"type": productType, "count": count})
	}
	
	return results
}

// GetCampaignStatusChart 获取计划状态分布
func (s *StatisticsService) GetCampaignStatusChart() []gin.H {
	var results []gin.H
	
	rows, _ := s.db.Model(&models.Campaign{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Rows()
	defer rows.Close()
	
	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		results = append(results, gin.H{"status": status, "count": count})
	}
	
	return results
}

// GetRecentTransactions 获取最近交易
func (s *StatisticsService) GetRecentTransactions(limit int) []gin.H {
	var transactions []models.Transaction
	s.db.Preload("Customer").Order("created_at DESC").Limit(limit).Find(&transactions)
	
	var results []gin.H
	for _, t := range transactions {
		customerName := ""
		if t.Customer.ID != 0 {
			customerName = t.Customer.Name
		}
		
		results = append(results, gin.H{
			"id": t.ID,
			"customer": customerName,
			"type": t.Type,
			"amount": t.Amount,
			"status": t.Status,
			"created_at": t.CreatedAt,
		})
	}
	
	return results
}

// GetRecentCustomers 获取最近注册客户
func (s *StatisticsService) GetRecentCustomers(limit int) []gin.H {
	var customers []models.Customer
	s.db.Order("created_at DESC").Limit(limit).Find(&customers)
	
	var results []gin.H
	for _, c := range customers {
		results = append(results, gin.H{
			"id": c.ID,
			"name": c.Name,
			"email": c.Email,
			"phone": c.Phone,
			"status": c.Status,
			"created_at": c.CreatedAt,
		})
	}
	
	return results
}

// GetTopProducts 获取热门产品
func (s *StatisticsService) GetTopProducts(limit int) []gin.H {
	var products []models.Product
	
	// 这里简单按创建时间排序，实际应该按销售量或其他指标排序
	s.db.Where("status = ?", "active").Order("created_at DESC").Limit(limit).Find(&products)
	
	var results []gin.H
	for _, p := range products {
		results = append(results, gin.H{
			"id": p.ID,
			"name": p.Name,
			"type": p.Type,
			"company": p.Company,
			"status": p.Status,
		})
	}
	
	return results
}

// GetTopCampaigns 获取热门计划
func (s *StatisticsService) GetTopCampaigns(limit int) []gin.H {
	var campaigns []models.Campaign
	
	s.db.Preload("Product").Where("status = ?", "running").Order("created_at DESC").Limit(limit).Find(&campaigns)
	
	var results []gin.H
	for _, c := range campaigns {
		productName := ""
		if c.Product.ID != 0 {
			productName = c.Product.Name
		}
		
		results = append(results, gin.H{
			"id": c.ID,
			"name": c.Name,
			"product": productName,
			"status": c.Status,
		})
	}
	
	return results
}

// GetTopCustomers 获取优质客户
func (s *StatisticsService) GetTopCustomers(limit int) []gin.H {
	var results []gin.H
	
	// 按充值金额排序获取优质客户
	rows, _ := s.db.Model(&models.Transaction{}).
		Select("user_id, SUM(amount) as total_amount").
		Where("type = ? AND status = ?", models.TransactionTypeRecharge, models.TransactionStatusSuccess).
		Group("user_id").
		Order("total_amount DESC").
		Limit(limit).
		Rows()
	defer rows.Close()
	
	for rows.Next() {
		var userID uint
		var totalAmount float64
		rows.Scan(&userID, &totalAmount)
		
		var customer models.Customer
		if err := s.db.First(&customer, userID).Error; err == nil {
			results = append(results, gin.H{
				"id": customer.ID,
				"name": customer.Name,
				"email": customer.Email,
				"total_amount": totalAmount,
			})
		}
	}
	
	return results
}

// 以下是其他统计方法的简化实现
func (s *StatisticsService) GetProductsOverview() gin.H {
	var total, active, inactive int64
	s.db.Model(&models.Product{}).Count(&total)
	s.db.Model(&models.Product{}).Where("status = ?", "active").Count(&active)
	s.db.Model(&models.Product{}).Where("status = ?", "inactive").Count(&inactive)
	
	return gin.H{
		"total": total,
		"active": active,
		"inactive": inactive,
	}
}

func (s *StatisticsService) GetCampaignsOverview() gin.H {
	var total, running, paused, completed int64
	s.db.Model(&models.Campaign{}).Count(&total)
	s.db.Model(&models.Campaign{}).Where("status = ?", "running").Count(&running)
	s.db.Model(&models.Campaign{}).Where("status = ?", "paused").Count(&paused)
	s.db.Model(&models.Campaign{}).Where("status = ?", "completed").Count(&completed)
	
	return gin.H{
		"total": total,
		"running": running,
		"paused": paused,
		"completed": completed,
	}
}

func (s *StatisticsService) GetCustomersOverview() gin.H {
	var total, active, banned int64
	s.db.Model(&models.Customer{}).Count(&total)
	s.db.Model(&models.Customer{}).Where("status = ?", "active").Count(&active)
	s.db.Model(&models.Customer{}).Where("status = ?", "banned").Count(&banned)
	
	return gin.H{
		"total": total,
		"active": active,
		"banned": banned,
	}
}

func (s *StatisticsService) GetFinanceOverview(startDate, endDate string) gin.H {
	return gin.H{
		"total_revenue": s.GetTotalRevenue(startDate, endDate),
		"today_revenue": s.GetTodayRevenue(),
		"monthly_growth": s.GetMonthlyGrowth(),
	}
}

func (s *StatisticsService) GetCouponsOverview() gin.H {
	var total, used, expired int64
	s.db.Model(&models.Coupon{}).Count(&total)
	s.db.Model(&models.UserCoupon{}).Where("status = ?", "used").Count(&used)
	s.db.Model(&models.UserCoupon{}).Where("status = ?", "expired").Count(&expired)
	
	return gin.H{
		"total": total,
		"used": used,
		"expired": expired,
	}
}

func (s *StatisticsService) GetAuthCodesOverview() gin.H {
	var total, used, expired int64
	s.db.Model(&models.AuthCode{}).Count(&total)
	s.db.Model(&models.AuthCode{}).Where("status = ?", "used").Count(&used)
	s.db.Model(&models.AuthCode{}).Where("expired_at < ?", time.Now()).Count(&expired)
	
	return gin.H{
		"total": total,
		"used": used,
		"expired": expired,
	}
}

// 以下方法返回简化的数据，实际应根据业务需求实现
func (s *StatisticsService) GetProductStatistics(page, pageSize int, productType, status string) gin.H {
	return gin.H{"data": []gin.H{}, "total": 0}
}

func (s *StatisticsService) GetCampaignStatistics(page, pageSize int, productID, status string) gin.H {
	return gin.H{"data": []gin.H{}, "total": 0}
}

func (s *StatisticsService) GetCouponStatistics() gin.H {
	return gin.H{"data": []gin.H{}}
}

func (s *StatisticsService) GetRevenueStatistics(startDate, endDate, groupBy string) gin.H {
	return gin.H{"data": []gin.H{}}
}

func (s *StatisticsService) GetUserStatistics(startDate, endDate string) gin.H {
	return gin.H{"data": []gin.H{}}
}

func (s *StatisticsService) GetTrendsStatistics(days int) gin.H {
	return gin.H{"data": []gin.H{}}
}

func (s *StatisticsService) GetRealtimeStatistics() gin.H {
	return gin.H{
		"online_users": 0,
		"active_campaigns": s.GetActiveCampaigns(),
		"pending_transactions": 0,
	}
}

func (s *StatisticsService) GetComparisonStatistics(compType string, ids []uint, startDate, endDate string) gin.H {
	return gin.H{"data": []gin.H{}}
}

func (s *StatisticsService) GetForecastStatistics(days int) gin.H {
	return gin.H{"data": []gin.H{}}
}

func (s *StatisticsService) GenerateCustomReport(metrics, dimensions []string, filters map[string]interface{}, startDate, endDate string) gin.H {
	return gin.H{"data": []gin.H{}}
}

func (s *StatisticsService) ExportReport(exportType, report, startDate, endDate string) (string, error) {
	// 实际应生成文件并返回路径
	return "/exports/report_" + time.Now().Format("20060102150405") + "." + exportType, nil
}