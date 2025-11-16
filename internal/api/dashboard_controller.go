package api

import (
	"github.com/ad-platform/backend/internal/service"
	"github.com/ad-platform/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"time"
)

// DashboardController 仪表盘控制器
type DashboardController struct {
	dashboardService  *service.DashboardService
	statisticsService *service.StatisticsService
}

// NewDashboardController 创建仪表盘控制器实例
func NewDashboardController() *DashboardController {
	return &DashboardController{
		dashboardService:  service.NewDashboardService(),
		statisticsService: service.NewStatisticsService(),
	}
}

// GetStats 获取仪表盘统计数据
func (ctrl *DashboardController) GetStats(c *gin.Context) {
	// 获取时间范围参数
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	
	// 如果没有提供日期范围，默认使用最近30天
	if startDate == "" || endDate == "" {
		endDate = time.Now().Format("2006-01-02")
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}

	stats := gin.H{
		"overview": gin.H{
			"totalProducts":    ctrl.statisticsService.GetTotalProducts(),
			"activeCampaigns":  ctrl.statisticsService.GetActiveCampaigns(),
			"totalCustomers":   ctrl.statisticsService.GetTotalCustomers(),
			"totalRevenue":     ctrl.statisticsService.GetTotalRevenue(startDate, endDate),
			"todayRevenue":     ctrl.statisticsService.GetTodayRevenue(),
			"monthlyGrowth":    ctrl.statisticsService.GetMonthlyGrowth(),
			"newCustomersToday": ctrl.dashboardService.GetNewCustomersToday(),
			"pendingOrders":    ctrl.dashboardService.GetPendingOrdersCount(),
		},
		"charts": gin.H{
			"revenueChart":   ctrl.statisticsService.GetRevenueChart(startDate, endDate),
			"userChart":      ctrl.statisticsService.GetUserChart(startDate, endDate),
			"productChart":   ctrl.statisticsService.GetProductTypeChart(),
			"campaignChart":  ctrl.statisticsService.GetCampaignStatusChart(),
		},
		"recent": gin.H{
			"recentTransactions": ctrl.statisticsService.GetRecentTransactions(5),
			"recentCustomers":    ctrl.statisticsService.GetRecentCustomers(5),
			"topProducts":        ctrl.statisticsService.GetTopProducts(5),
			"topCampaigns":       ctrl.statisticsService.GetTopCampaigns(5),
		},
		"quickStats": gin.H{
			"conversionRate": ctrl.dashboardService.GetConversionRate(),
			"avgOrderValue":  ctrl.dashboardService.GetAverageOrderValue(),
			"activeUsers":    ctrl.dashboardService.GetActiveUsersCount(),
			"returningUsers": ctrl.dashboardService.GetReturningUsersRate(),
		},
	}

	utils.Success(c, stats)
}

// GetActivities 获取最近活动
func (ctrl *DashboardController) GetActivities(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); l != "" {
		// 可以解析limit参数
	}
	
	activities := ctrl.dashboardService.GetRecentActivities(limit)
	utils.Success(c, activities)
}

// GetRevenueTrend 获取收入趋势
func (ctrl *DashboardController) GetRevenueTrend(c *gin.Context) {
	days := 30
	if d := c.Query("days"); d != "" {
		// 可以解析days参数
	}
	
	trend := ctrl.dashboardService.GetRevenueTrend(days)
	utils.Success(c, trend)
}

// GetUserGrowth 获取用户增长数据
func (ctrl *DashboardController) GetUserGrowth(c *gin.Context) {
	period := c.DefaultQuery("period", "month") // day, week, month, year
	
	growth := ctrl.dashboardService.GetUserGrowth(period)
	utils.Success(c, growth)
}

// GetCampaignPerformance 获取营销计划表现
func (ctrl *DashboardController) GetCampaignPerformance(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		// 可以解析limit参数
	}
	
	performance := ctrl.dashboardService.GetCampaignPerformance(limit)
	utils.Success(c, performance)
}

// GetRealtime 获取实时数据
func (ctrl *DashboardController) GetRealtime(c *gin.Context) {
	realtime := ctrl.dashboardService.GetRealtimeData()
	utils.Success(c, realtime)
}