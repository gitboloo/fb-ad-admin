package services

import (
	"fmt"
	"backend/database"
	"backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

// DashboardService 仪表盘服务
type DashboardService struct {
	db *gorm.DB
}

// NewDashboardService 创建仪表盘服务实例
func NewDashboardService() *DashboardService {
	return &DashboardService{
		db: database.GetDB(),
	}
}

// GetNewCustomersToday 获取今日新增客户数
func (s *DashboardService) GetNewCustomersToday() int64 {
	var count int64
	today := time.Now().Format("2006-01-02")
	s.db.Model(&models.Customer{}).
		Where("DATE(created_at) = ?", today).
		Count(&count)
	return count
}

// GetPendingOrdersCount 获取待处理订单数
func (s *DashboardService) GetPendingOrdersCount() int64 {
	var count int64
	s.db.Model(&models.Transaction{}).
		Where("status = ?", "pending").
		Count(&count)
	return count
}

// GetConversionRate 获取转化率
func (s *DashboardService) GetConversionRate() float64 {
	// 简化实现：活跃用户数/总用户数
	var total, active int64
	s.db.Model(&models.Customer{}).Count(&total)
	s.db.Model(&models.Customer{}).Where("status = ?", "active").Count(&active)
	
	if total == 0 {
		return 0
	}
	return float64(active) / float64(total) * 100
}

// GetAverageOrderValue 获取平均订单价值
func (s *DashboardService) GetAverageOrderValue() float64 {
	var result struct {
		AvgAmount float64
	}
	
	s.db.Model(&models.Transaction{}).
		Select("AVG(amount) as avg_amount").
		Where("type = ? AND status = ?", "recharge", "completed").
		Scan(&result)
		
	return result.AvgAmount
}

// GetActiveUsersCount 获取活跃用户数
func (s *DashboardService) GetActiveUsersCount() int64 {
	var count int64
	// 最近7天有交易的用户
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	
	s.db.Model(&models.Transaction{}).
		Where("created_at > ?", sevenDaysAgo).
		Group("user_id").
		Count(&count)
		
	return count
}

// GetReturningUsersRate 获取回访用户率
func (s *DashboardService) GetReturningUsersRate() float64 {
	// 简化实现：有多次交易的用户占比
	var total, returning int64
	
	// 总用户数
	s.db.Model(&models.Customer{}).Count(&total)
	
	// 有多次交易的用户
	s.db.Model(&models.Transaction{}).
		Select("user_id").
		Group("user_id").
		Having("COUNT(*) > 1").
		Count(&returning)
	
	if total == 0 {
		return 0
	}
	return float64(returning) / float64(total) * 100
}

// GetRecentActivities 获取最近活动记录
func (s *DashboardService) GetRecentActivities(limit int) []gin.H {
	activities := []gin.H{}
	
	// 获取最近的交易活动
	var transactions []models.Transaction
	s.db.Preload("Customer").
		Order("created_at DESC").
		Limit(limit/2).
		Find(&transactions)
		
	for _, t := range transactions {
		customerName := "Unknown"
		if t.Customer.ID != 0 {
			customerName = t.Customer.Name
		}
		
		activities = append(activities, gin.H{
			"type":        "transaction",
			"title":       getTransactionTitle(int(t.Type)),
			"description": customerName + " " + getTransactionAction(int(t.Type)) + " ￥" + formatFloat(t.Amount),
			"time":        t.CreatedAt,
			"status":      t.Status,
		})
	}
	
	// 获取最近的新用户注册
	var customers []models.Customer
	s.db.Order("created_at DESC").
		Limit(limit/2).
		Find(&customers)
		
	for _, c := range customers {
		activities = append(activities, gin.H{
			"type":        "user_registration",
			"title":       "新用户注册",
			"description": c.Name + " 完成注册",
			"time":        c.CreatedAt,
			"status":      "success",
		})
	}
	
	return activities
}

// GetRevenueTrend 获取收入趋势
func (s *DashboardService) GetRevenueTrend(days int) []gin.H {
	trend := []gin.H{}
	
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		
		var total float64
		s.db.Model(&models.Transaction{}).
			Where("DATE(created_at) = ? AND type = ? AND status = ?", date, "recharge", "completed").
			Pluck("COALESCE(SUM(amount), 0)", &total)
			
		trend = append(trend, gin.H{
			"date":   date,
			"amount": total,
		})
	}
	
	return trend
}

// GetUserGrowth 获取用户增长数据
func (s *DashboardService) GetUserGrowth(period string) gin.H {
	var groupFormat string
	var days int
	
	switch period {
	case "day":
		groupFormat = "%Y-%m-%d"
		days = 30
	case "week":
		groupFormat = "%Y-%u"
		days = 84 // 12 weeks
	case "month":
		groupFormat = "%Y-%m"
		days = 365
	case "year":
		groupFormat = "%Y"
		days = 1825 // 5 years
	default:
		groupFormat = "%Y-%m-%d"
		days = 30
	}
	
	startDate := time.Now().AddDate(0, 0, -days)
	
	var results []struct {
		Period string
		Count  int64
	}
	
	s.db.Model(&models.Customer{}).
		Select("DATE_FORMAT(created_at, ?) as period, COUNT(*) as count", groupFormat).
		Where("created_at >= ?", startDate).
		Group("period").
		Order("period").
		Scan(&results)
		
	data := []gin.H{}
	for _, r := range results {
		data = append(data, gin.H{
			"period": r.Period,
			"count":  r.Count,
		})
	}
	
	// 计算增长率
	var growthRate float64
	if len(results) >= 2 {
		last := results[len(results)-1].Count
		previous := results[len(results)-2].Count
		if previous > 0 {
			growthRate = float64(last-previous) / float64(previous) * 100
		}
	}
	
	return gin.H{
		"data":       data,
		"growthRate": growthRate,
		"period":     period,
	}
}

// GetCampaignPerformance 获取营销计划表现
func (s *DashboardService) GetCampaignPerformance(limit int) []gin.H {
	performance := []gin.H{}
	
	var campaigns []models.Campaign
	s.db.Preload("Product").
		Where("status = ?", "running").
		Order("created_at DESC").
		Limit(limit).
		Find(&campaigns)
		
	for _, c := range campaigns {
		productName := ""
		if c.Product.ID != 0 {
			productName = c.Product.Name
		}
		
		// 这里简化处理，实际应该有更详细的统计数据
		performance = append(performance, gin.H{
			"id":          c.ID,
			"name":        c.Name,
			"product":     productName,
			"status":      c.Status,
			"impressions": 0, // 需要实际统计
			"clicks":      0, // 需要实际统计
			"conversions": 0, // 需要实际统计
			"ctr":         0.0, // 点击率
			"cvr":         0.0, // 转化率
		})
	}
	
	return performance
}

// GetRealtimeData 获取实时数据
func (s *DashboardService) GetRealtimeData() gin.H {
	// 当前在线用户数（简化实现）
	var onlineUsers int64 = 0
	
	// 今日订单数
	var todayOrders int64
	today := time.Now().Format("2006-01-02")
	s.db.Model(&models.Transaction{}).
		Where("DATE(created_at) = ?", today).
		Count(&todayOrders)
	
	// 今日收入
	var todayRevenue float64
	s.db.Model(&models.Transaction{}).
		Where("DATE(created_at) = ? AND type = ? AND status = ?", today, "recharge", "completed").
		Pluck("COALESCE(SUM(amount), 0)", &todayRevenue)
	
	// 待处理事项
	var pendingItems int64
	s.db.Model(&models.Transaction{}).
		Where("status = ?", "pending").
		Count(&pendingItems)
	
	return gin.H{
		"onlineUsers":   onlineUsers,
		"todayOrders":   todayOrders,
		"todayRevenue":  todayRevenue,
		"pendingItems":  pendingItems,
		"serverTime":    time.Now(),
	}
}

// 辅助函数
func getTransactionTitle(transType int) string {
	switch models.TransactionType(transType) {
	case models.TransactionTypeRecharge:
		return "充值"
	case models.TransactionTypeWithdraw:
		return "提现"
	case models.TransactionTypeConsume:
		return "消费"
	case models.TransactionTypeRefund:
		return "退款"
	case models.TransactionTypeReward:
		return "奖励"
	default:
		return "交易"
	}
}

func getTransactionAction(transType int) string {
	switch models.TransactionType(transType) {
	case models.TransactionTypeRecharge:
		return "充值了"
	case models.TransactionTypeWithdraw:
		return "提现了"
	case models.TransactionTypeConsume:
		return "消费了"
	case models.TransactionTypeRefund:
		return "退款了"
	case models.TransactionTypeReward:
		return "获得奖励"
	default:
		return "交易了"
	}
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}