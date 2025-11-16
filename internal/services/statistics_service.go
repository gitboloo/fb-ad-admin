package services

import (
	"fmt"
	"time"

	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/repositories"
	"github.com/ad-platform/backend/internal/types"
)

// StatisticsService 统计服务
type StatisticsService struct {
	productRepo     *repositories.ProductRepository
	campaignRepo    *repositories.CampaignRepository
	customerRepo    *repositories.CustomerRepository
	transactionRepo *repositories.TransactionRepository
	couponRepo      *repositories.CouponRepository
	authCodeRepo    *repositories.AuthCodeRepository
}

// NewStatisticsService 创建统计服务
func NewStatisticsService() *StatisticsService {
	return &StatisticsService{
		productRepo:     repositories.NewProductRepository(),
		campaignRepo:    repositories.NewCampaignRepository(),
		customerRepo:    repositories.NewCustomerRepository(),
		transactionRepo: repositories.NewTransactionRepository(),
		couponRepo:      repositories.NewCouponRepository(),
		authCodeRepo:    repositories.NewAuthCodeRepository(),
	}
}

// GetOverviewStats 获取总览统计
func (ss *StatisticsService) GetOverviewStats() (map[string]interface{}, error) {
	// 获取各模块的基本统计
	productStats, _ := ss.productRepo.GetStatistics()
	customerStats, _ := ss.customerRepo.GetStatistics()
	transactionStats, _ := ss.transactionRepo.GetStatistics()
	campaignStats, _ := ss.campaignRepo.GetStatistics()

	// 计算关键指标
	totalRevenue, _ := ss.transactionRepo.GetAmountByType(models.TransactionTypeRecharge)
	totalBalance, _ := ss.customerRepo.GetTotalBalance()
	activeCustomers, _ := ss.customerRepo.GetActiveCustomerCount()

	return map[string]interface{}{
		"products": map[string]interface{}{
			"total":  productStats.Total,
			"active": productStats.Active,
			"growth": productStats.Growth,
		},
		"customers": map[string]interface{}{
			"total":  customerStats.Total,
			"active": activeCustomers,
			"growth": customerStats.Growth,
		},
		"transactions": map[string]interface{}{
			"total":   transactionStats.Total,
			"revenue": totalRevenue,
			"growth":  transactionStats.Growth,
		},
		"campaigns": map[string]interface{}{
			"total":  campaignStats.Total,
			"active": campaignStats.Active,
			"growth": campaignStats.Growth,
		},
		"financial": map[string]interface{}{
			"total_revenue": totalRevenue,
			"total_balance": totalBalance,
		},
		"updated_at": time.Now(),
	}, nil
}

// GetProductStats 获取产品统计
func (ss *StatisticsService) GetProductStats(startTime, endTime *time.Time, groupBy string) (map[string]interface{}, error) {
	stats, err := ss.productRepo.GetStatistics()
	if err != nil {
		return nil, err
	}

	// 按类型统计
	typeStats := make(map[string]interface{})
	for i := models.ProductTypeApp; i <= models.ProductTypeOther; i++ {
		products, _ := ss.productRepo.GetByType(i)
		typeName := ss.getProductTypeName(i)
		typeStats[typeName] = len(products)
	}

	// 按状态统计
	statusStats := make(map[string]interface{})
	for i := models.ProductStatusInactive; i <= models.ProductStatusSuspended; i++ {
		products, _ := ss.productRepo.GetByStatus(i)
		statusName := ss.getProductStatusName(i)
		statusStats[statusName] = len(products)
	}

	return map[string]interface{}{
		"overview":     stats,
		"by_type":      typeStats,
		"by_status":    statusStats,
		"trend_data":   stats.TrendData,
		"categories":   stats.Categories,
		"last_updated": time.Now(),
	}, nil
}

// GetCampaignStats 获取计划统计
func (ss *StatisticsService) GetCampaignStats(startTime, endTime *time.Time, groupBy string, productID *uint) (map[string]interface{}, error) {
	stats, err := ss.campaignRepo.GetStatistics()
	if err != nil {
		return nil, err
	}

	// 按状态统计
	statusStats := make(map[string]interface{})
	for i := models.CampaignStatusInactive; i <= models.CampaignStatusEnded; i++ {
		campaigns, _ := ss.campaignRepo.GetByStatus(i)
		statusName := ss.getCampaignStatusName(i)
		statusStats[statusName] = len(campaigns)
	}

	// 如果指定了产品ID，获取该产品的计划统计
	var productCampaigns interface{}
	if productID != nil {
		campaigns, _ := ss.campaignRepo.GetByProductID(*productID)
		productCampaigns = map[string]interface{}{
			"total":        len(campaigns),
			"campaigns":    campaigns,
			"product_id":   *productID,
		}
	}

	// 获取正在运行的计划
	runningCampaigns, _ := ss.campaignRepo.GetRunningCampaigns()

	return map[string]interface{}{
		"overview":          stats,
		"by_status":         statusStats,
		"running_campaigns": len(runningCampaigns),
		"product_campaigns": productCampaigns,
		"trend_data":        stats.TrendData,
		"categories":        stats.Categories,
		"last_updated":      time.Now(),
	}, nil
}

// GetCouponStats 获取优惠券统计
func (ss *StatisticsService) GetCouponStats(startTime, endTime *time.Time, groupBy string) (map[string]interface{}, error) {
	stats, err := ss.couponRepo.GetStatistics()
	if err != nil {
		return nil, err
	}

	// 按类型统计
	typeStats := make(map[string]interface{})
	for i := models.CouponTypeValueAdded; i <= models.CouponTypeFixed; i++ {
		coupons, _ := ss.couponRepo.GetByType(i)
		typeName := ss.getCouponTypeName(i)
		typeStats[typeName] = len(coupons)
	}

	// 按状态统计
	statusStats := make(map[string]interface{})
	for i := models.CouponStatusInactive; i <= models.CouponStatusUsedUp; i++ {
		coupons, _ := ss.couponRepo.GetByStatus(i)
		statusName := ss.getCouponStatusName(i)
		statusStats[statusName] = len(coupons)
	}

	return map[string]interface{}{
		"overview":     stats,
		"by_type":      typeStats,
		"by_status":    statusStats,
		"trend_data":   stats.TrendData,
		"categories":   stats.Categories,
		"last_updated": time.Now(),
	}, nil
}

// GetRevenueStats 获取收入统计
func (ss *StatisticsService) GetRevenueStats(startTime, endTime *time.Time, groupBy string) (map[string]interface{}, error) {
	// 获取各类型交易金额
	rechargeAmount, _ := ss.transactionRepo.GetAmountByType(models.TransactionTypeRecharge)
	withdrawAmount, _ := ss.transactionRepo.GetAmountByType(models.TransactionTypeWithdraw)
	consumeAmount, _ := ss.transactionRepo.GetAmountByType(models.TransactionTypeConsume)
	refundAmount, _ := ss.transactionRepo.GetAmountByType(models.TransactionTypeRefund)

	// 计算净收入和利润
	netRevenue := rechargeAmount - withdrawAmount
	profit := consumeAmount - refundAmount

	// 获取交易趋势
	transactionStats, _ := ss.transactionRepo.GetStatistics()

	return map[string]interface{}{
		"total_revenue": rechargeAmount,
		"total_withdraw": withdrawAmount,
		"net_revenue":   netRevenue,
		"total_consume": consumeAmount,
		"total_refund":  refundAmount,
		"profit":        profit,
		"trend_data":    transactionStats.TrendData,
		"growth":        transactionStats.Growth,
		"by_type": map[string]interface{}{
			"recharge": rechargeAmount,
			"withdraw": withdrawAmount,
			"consume":  consumeAmount,
			"refund":   refundAmount,
		},
		"last_updated": time.Now(),
	}, nil
}

// GetUserStats 获取用户统计
func (ss *StatisticsService) GetUserStats(startTime, endTime *time.Time, groupBy string) (map[string]interface{}, error) {
	stats, err := ss.customerRepo.GetStatistics()
	if err != nil {
		return nil, err
	}

	// 获取用户余额统计
	totalBalance, _ := ss.customerRepo.GetTotalBalance()
	withBalanceUsers, _ := ss.customerRepo.GetCustomersWithBalance()

	// 按状态统计
	statusStats := make(map[string]interface{})
	for i := models.CustomerStatusInactive; i <= models.CustomerStatusBlocked; i++ {
		customers, _ := ss.customerRepo.GetByStatus(i)
		statusName := ss.getCustomerStatusName(i)
		statusStats[statusName] = len(customers)
	}

	return map[string]interface{}{
		"overview":              stats,
		"by_status":             statusStats,
		"total_balance":         totalBalance,
		"users_with_balance":    len(withBalanceUsers),
		"trend_data":            stats.TrendData,
		"categories":            stats.Categories,
		"last_updated":          time.Now(),
	}, nil
}

// GetTrendAnalysis 获取趋势分析
func (ss *StatisticsService) GetTrendAnalysis(metric, period, compare string) (map[string]interface{}, error) {
	// var days int  // 暂时注释，后续使用
	switch period {
	case "7d":
		// days = 7
	case "30d":
		// days = 30
	case "90d":
		// days = 90
	case "1y":
		// days = 365
	default:
		// days = 30
	}

	var currentData, compareData []types.TrendData
	var currentTotal, compareTotal int64

	// 根据指标获取相应数据
	switch metric {
	case "products":
		stats, _ := ss.productRepo.GetStatistics()
		currentData = stats.TrendData
		currentTotal = stats.Total
	case "campaigns":
		stats, _ := ss.campaignRepo.GetStatistics()
		currentData = stats.TrendData
		currentTotal = stats.Total
	case "users":
		stats, _ := ss.customerRepo.GetStatistics()
		currentData = stats.TrendData
		currentTotal = stats.Total
	case "revenue":
		stats, _ := ss.transactionRepo.GetStatistics()
		currentData = stats.TrendData
		currentTotal = stats.Total
	default:
		// 总览数据
		// overview, _ := ss.GetOverviewStats()  // 暂时注释
		currentTotal = 1000 // 示例数据
	}

	// 计算变化率
	var changeRate float64
	if compareTotal > 0 {
		changeRate = float64(currentTotal-compareTotal) / float64(compareTotal) * 100
	}

	return map[string]interface{}{
		"metric":        metric,
		"period":        period,
		"compare":       compare,
		"current_data":  currentData,
		"compare_data":  compareData,
		"current_total": currentTotal,
		"compare_total": compareTotal,
		"change_rate":   changeRate,
		"trend":         ss.getTrendDirection(changeRate),
		"last_updated":  time.Now(),
	}, nil
}

// GetRealtimeStats 获取实时统计
func (ss *StatisticsService) GetRealtimeStats() (map[string]interface{}, error) {
	// 获取最近的活动数据
	recentTransactions, _ := ss.transactionRepo.GetRecentTransactions(10)
	recentCustomers, _ := ss.customerRepo.GetRecentCustomers(10)

	// 获取待处理的交易
	pendingTransactions, _ := ss.transactionRepo.GetByStatus(models.TransactionStatusPending)

	// 计算今日统计
	// todayStart := time.Now().Truncate(24 * time.Hour)  // 暂时注释
	todayStats := map[string]interface{}{
		"new_customers":    0,
		"new_transactions": 0,
		"revenue":          0.0,
	}

	return map[string]interface{}{
		"recent_transactions":  recentTransactions,
		"recent_customers":     recentCustomers,
		"pending_transactions": len(pendingTransactions),
		"today_stats":          todayStats,
		"online_users":         42, // 示例数据
		"system_load":          0.65,
		"last_updated":         time.Now(),
	}, nil
}

// GetTopPerformers 获取顶级表现者
func (ss *StatisticsService) GetTopPerformers(category, metric string, limit int) (map[string]interface{}, error) {
	var performers interface{}

	switch category {
	case "products":
		// 根据指标获取顶级产品
		performers = ss.getTopProducts(metric, limit)
	case "campaigns":
		// 根据指标获取顶级计划
		performers = ss.getTopCampaigns(metric, limit)
	case "users":
		// 根据指标获取顶级用户
		performers = ss.getTopUsers(metric, limit)
	default:
		return nil, &ServiceError{
			Code:    400,
			Message: "不支持的分类",
		}
	}

	return map[string]interface{}{
		"category":     category,
		"metric":       metric,
		"limit":        limit,
		"performers":   performers,
		"last_updated": time.Now(),
	}, nil
}

// GetComparisonAnalysis 获取对比分析
func (ss *StatisticsService) GetComparisonAnalysis(itemType string, ids []uint, metric string, startTime, endTime *time.Time) (map[string]interface{}, error) {
	var items []map[string]interface{}

	switch itemType {
	case "products":
		for _, id := range ids {
			product, err := ss.productRepo.GetByID(id)
			if err != nil {
				continue
			}
			items = append(items, map[string]interface{}{
				"id":     id,
				"name":   product.Name,
				"type":   product.Type,
				"status": product.Status,
				"metric_value": ss.getProductMetricValue(id, metric),
			})
		}
	case "campaigns":
		for _, id := range ids {
			campaign, err := ss.campaignRepo.GetByID(id)
			if err != nil {
				continue
			}
			items = append(items, map[string]interface{}{
				"id":     id,
				"name":   campaign.Name,
				"status": campaign.Status,
				"metric_value": ss.getCampaignMetricValue(id, metric),
			})
		}
	case "users":
		for _, id := range ids {
			customer, err := ss.customerRepo.GetByID(id)
			if err != nil {
				continue
			}
			items = append(items, map[string]interface{}{
				"id":     id,
				"name":   customer.Name,
				"email":  customer.Email,
				"status": customer.Status,
				"metric_value": ss.getUserMetricValue(id, metric),
			})
		}
	}

	return map[string]interface{}{
		"type":         itemType,
		"metric":       metric,
		"items":        items,
		"start_date":   startTime,
		"end_date":     endTime,
		"last_updated": time.Now(),
	}, nil
}

// GetForecastAnalysis 获取预测分析
func (ss *StatisticsService) GetForecastAnalysis(metric, period, algorithm string) (map[string]interface{}, error) {
	// 这里应该实现实际的预测算法
	// 暂时返回示例数据
	
	var forecastData []map[string]interface{}
	days := ss.getPeriodDays(period)
	
	for i := 1; i <= days; i++ {
		date := time.Now().AddDate(0, 0, i)
		value := ss.calculateForecastValue(metric, i, algorithm)
		
		forecastData = append(forecastData, map[string]interface{}{
			"date":  date.Format("2006-01-02"),
			"value": value,
			"confidence": 0.85, // 置信度
		})
	}

	return map[string]interface{}{
		"metric":        metric,
		"period":        period,
		"algorithm":     algorithm,
		"forecast_data": forecastData,
		"accuracy":      0.85,
		"last_updated":  time.Now(),
	}, nil
}

// GenerateCustomReport 生成自定义报告
func (ss *StatisticsService) GenerateCustomReport(name, description string, filters map[string]interface{}, metrics []string, groupBy string, startTime, endTime *time.Time) (map[string]interface{}, error) {
	report := map[string]interface{}{
		"name":        name,
		"description": description,
		"filters":     filters,
		"metrics":     metrics,
		"group_by":    groupBy,
		"start_date":  startTime,
		"end_date":    endTime,
		"generated_at": time.Now(),
		"data":        make(map[string]interface{}),
	}

	// 根据指标生成报告数据
	data := report["data"].(map[string]interface{})
	for _, metric := range metrics {
		switch metric {
		case "products":
			data[metric], _ = ss.GetProductStats(startTime, endTime, groupBy)
		case "campaigns":
			data[metric], _ = ss.GetCampaignStats(startTime, endTime, groupBy, nil)
		case "users":
			data[metric], _ = ss.GetUserStats(startTime, endTime, groupBy)
		case "revenue":
			data[metric], _ = ss.GetRevenueStats(startTime, endTime, groupBy)
		}
	}

	return report, nil
}

// ExportReport 导出报告
func (ss *StatisticsService) ExportReport(reportType, format string, config map[string]interface{}) (map[string]interface{}, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_report_%s.%s", reportType, timestamp, format)
	
	// 这里应该实现实际的报告导出逻辑
	return map[string]interface{}{
		"filename":    filename,
		"format":      format,
		"type":        reportType,
		"generated_at": time.Now(),
		"download_url": fmt.Sprintf("/api/downloads/%s", filename),
	}, nil
}

// GetAvailableMetrics 获取可用指标
func (ss *StatisticsService) GetAvailableMetrics(category string) ([]map[string]interface{}, error) {
	var metrics []map[string]interface{}

	switch category {
	case "products":
		metrics = []map[string]interface{}{
			{"key": "total_count", "name": "产品总数", "type": "number"},
			{"key": "active_count", "name": "活跃产品数", "type": "number"},
			{"key": "by_type", "name": "按类型分布", "type": "category"},
			{"key": "growth_rate", "name": "增长率", "type": "percentage"},
		}
	case "campaigns":
		metrics = []map[string]interface{}{
			{"key": "total_count", "name": "计划总数", "type": "number"},
			{"key": "running_count", "name": "运行中计划数", "type": "number"},
			{"key": "budget_total", "name": "总预算", "type": "currency"},
			{"key": "performance", "name": "表现指标", "type": "score"},
		}
	case "users":
		metrics = []map[string]interface{}{
			{"key": "total_count", "name": "用户总数", "type": "number"},
			{"key": "active_count", "name": "活跃用户数", "type": "number"},
			{"key": "total_balance", "name": "总余额", "type": "currency"},
			{"key": "registration_rate", "name": "注册率", "type": "rate"},
		}
	case "revenue":
		metrics = []map[string]interface{}{
			{"key": "total_revenue", "name": "总收入", "type": "currency"},
			{"key": "net_revenue", "name": "净收入", "type": "currency"},
			{"key": "profit", "name": "利润", "type": "currency"},
			{"key": "growth_rate", "name": "收入增长率", "type": "percentage"},
		}
	}

	return metrics, nil
}

// GetAvailableDimensions 获取可用维度
func (ss *StatisticsService) GetAvailableDimensions(category string) ([]map[string]interface{}, error) {
	var dimensions []map[string]interface{}

	switch category {
	case "time":
		dimensions = []map[string]interface{}{
			{"key": "day", "name": "按日", "type": "time"},
			{"key": "week", "name": "按周", "type": "time"},
			{"key": "month", "name": "按月", "type": "time"},
			{"key": "year", "name": "按年", "type": "time"},
		}
	case "products":
		dimensions = []map[string]interface{}{
			{"key": "type", "name": "产品类型", "type": "category"},
			{"key": "status", "name": "产品状态", "type": "status"},
			{"key": "company", "name": "公司", "type": "text"},
		}
	case "geography":
		dimensions = []map[string]interface{}{
			{"key": "country", "name": "国家", "type": "geography"},
			{"key": "region", "name": "地区", "type": "geography"},
			{"key": "city", "name": "城市", "type": "geography"},
		}
	}

	return dimensions, nil
}

// 辅助方法

func (ss *StatisticsService) getProductTypeName(productType models.ProductType) string {
	switch productType {
	case models.ProductTypeApp:
		return "应用"
	case models.ProductTypeGame:
		return "游戏"
	case models.ProductTypeWeb:
		return "网站"
	case models.ProductTypeOther:
		return "其他"
	default:
		return "未知"
	}
}

func (ss *StatisticsService) getProductStatusName(status models.ProductStatus) string {
	switch status {
	case models.ProductStatusActive:
		return "活跃"
	case models.ProductStatusInactive:
		return "未激活"
	case models.ProductStatusSuspended:
		return "已暂停"
	default:
		return "未知"
	}
}

func (ss *StatisticsService) getCampaignStatusName(status models.CampaignStatus) string {
	switch status {
	case models.CampaignStatusActive:
		return "活跃"
	case models.CampaignStatusInactive:
		return "未激活"
	case models.CampaignStatusPaused:
		return "已暂停"
	case models.CampaignStatusEnded:
		return "已结束"
	default:
		return "未知"
	}
}

func (ss *StatisticsService) getCouponTypeName(couponType models.CouponType) string {
	switch couponType {
	case models.CouponTypeValueAdded:
		return "增值券"
	case models.CouponTypeDiscount:
		return "折扣券"
	case models.CouponTypeTeam:
		return "团队券"
	case models.CouponTypeCustom:
		return "自定义券"
	case models.CouponTypeFixed:
		return "固定金额券"
	default:
		return "其他"
	}
}

func (ss *StatisticsService) getCouponStatusName(status models.CouponStatus) string {
	switch status {
	case models.CouponStatusActive:
		return "活跃"
	case models.CouponStatusInactive:
		return "未激活"
	case models.CouponStatusExpired:
		return "已过期"
	case models.CouponStatusUsedUp:
		return "已用完"
	default:
		return "未知"
	}
}

func (ss *StatisticsService) getCustomerStatusName(status models.CustomerStatus) string {
	switch status {
	case models.CustomerStatusActive:
		return "活跃"
	case models.CustomerStatusInactive:
		return "未激活"
	case models.CustomerStatusBlocked:
		return "已阻止"
	default:
		return "未知"
	}
}

func (ss *StatisticsService) getTrendDirection(changeRate float64) string {
	if changeRate > 5 {
		return "上升"
	} else if changeRate < -5 {
		return "下降"
	} else {
		return "稳定"
	}
}

func (ss *StatisticsService) getTopProducts(metric string, limit int) interface{} {
	// 实现获取顶级产品的逻辑
	return []map[string]interface{}{}
}

func (ss *StatisticsService) getTopCampaigns(metric string, limit int) interface{} {
	// 实现获取顶级计划的逻辑
	return []map[string]interface{}{}
}

func (ss *StatisticsService) getTopUsers(metric string, limit int) interface{} {
	// 实现获取顶级用户的逻辑
	users, _ := ss.customerRepo.GetTopCustomers(limit, "balance")
	return users
}

func (ss *StatisticsService) getProductMetricValue(id uint, metric string) float64 {
	// 根据指标计算产品的指标值
	return 0.0
}

func (ss *StatisticsService) getCampaignMetricValue(id uint, metric string) float64 {
	// 根据指标计算计划的指标值
	return 0.0
}

func (ss *StatisticsService) getUserMetricValue(id uint, metric string) float64 {
	// 根据指标计算用户的指标值
	customer, err := ss.customerRepo.GetByID(id)
	if err != nil {
		return 0.0
	}
	
	switch metric {
	case "balance":
		return customer.Balance
	case "transactions":
		transactions, _, _ := ss.transactionRepo.GetByUserID(id, &types.FilterRequest{
			PageRequest: types.PageRequest{
				Size: 1000,
			},
		})
		return float64(len(transactions))
	default:
		return 0.0
	}
}

func (ss *StatisticsService) getPeriodDays(period string) int {
	switch period {
	case "7d":
		return 7
	case "30d":
		return 30
	case "90d":
		return 90
	case "1y":
		return 365
	default:
		return 30
	}
}

func (ss *StatisticsService) calculateForecastValue(metric string, dayOffset int, algorithm string) float64 {
	// 这里应该实现实际的预测算法
	// 暂时返回示例数据
	base := 1000.0
	growth := 0.05 // 5% 增长率
	
	switch algorithm {
	case "linear":
		return base * (1 + growth*float64(dayOffset)/30)
	case "exponential":
		return base * (1 + growth)*float64(dayOffset)/30
	case "seasonal":
		seasonal := 1.0 + 0.1*float64(dayOffset%7)/7 // 周期性变化
		return base * (1 + growth*float64(dayOffset)/30) * seasonal
	default:
		return base * (1 + growth*float64(dayOffset)/30)
	}
}