package api

import (
	"backend/services"
	"backend/utils"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

// StatisticsController 统计控制器
type StatisticsController struct {
	statisticsService *services.StatisticsService
}

// NewStatisticsController 创建统计控制器实例
func NewStatisticsController() *StatisticsController {
	return &StatisticsController{
		statisticsService: services.NewStatisticsService(),
	}
}

// GetDashboardStats 获取仪表盘统计数据
func (ctrl *StatisticsController) GetDashboardStats(c *gin.Context) {
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
			"totalProducts": ctrl.statisticsService.GetTotalProducts(),
			"activeCampaigns": ctrl.statisticsService.GetActiveCampaigns(),
			"totalCustomers": ctrl.statisticsService.GetTotalCustomers(),
			"totalRevenue": ctrl.statisticsService.GetTotalRevenue(startDate, endDate),
			"todayRevenue": ctrl.statisticsService.GetTodayRevenue(),
			"monthlyGrowth": ctrl.statisticsService.GetMonthlyGrowth(),
		},
		"charts": gin.H{
			"revenueChart": ctrl.statisticsService.GetRevenueChart(startDate, endDate),
			"userChart": ctrl.statisticsService.GetUserChart(startDate, endDate),
			"productChart": ctrl.statisticsService.GetProductTypeChart(),
			"campaignChart": ctrl.statisticsService.GetCampaignStatusChart(),
		},
		"recent": gin.H{
			"recentTransactions": ctrl.statisticsService.GetRecentTransactions(5),
			"recentCustomers": ctrl.statisticsService.GetRecentCustomers(5),
			"topProducts": ctrl.statisticsService.GetTopProducts(5),
			"topCampaigns": ctrl.statisticsService.GetTopCampaigns(5),
		},
	}

	utils.Success(c, stats)
}

// GetOverview 获取总览统计
func (ctrl *StatisticsController) GetOverview(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	
	overview := gin.H{
		"products": ctrl.statisticsService.GetProductsOverview(),
		"campaigns": ctrl.statisticsService.GetCampaignsOverview(),
		"customers": ctrl.statisticsService.GetCustomersOverview(),
		"finance": ctrl.statisticsService.GetFinanceOverview(startDate, endDate),
		"coupons": ctrl.statisticsService.GetCouponsOverview(),
		"authCodes": ctrl.statisticsService.GetAuthCodesOverview(),
	}

	utils.Success(c, overview)
}

// GetProducts 获取产品统计
func (ctrl *StatisticsController) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	productType := c.Query("type")
	status := c.Query("status")
	
	result := ctrl.statisticsService.GetProductStatistics(page, pageSize, productType, status)
	utils.Success(c, result)
}

// GetCampaigns 获取计划统计
func (ctrl *StatisticsController) GetCampaigns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	productID := c.Query("product_id")
	status := c.Query("status")
	
	result := ctrl.statisticsService.GetCampaignStatistics(page, pageSize, productID, status)
	utils.Success(c, result)
}

// GetCoupons 获取优惠券统计
func (ctrl *StatisticsController) GetCoupons(c *gin.Context) {
	result := ctrl.statisticsService.GetCouponStatistics()
	utils.Success(c, result)
}

// GetRevenue 获取收入统计
func (ctrl *StatisticsController) GetRevenue(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "day") // day, week, month
	
	result := ctrl.statisticsService.GetRevenueStatistics(startDate, endDate, groupBy)
	utils.Success(c, result)
}

// GetUsers 获取用户统计
func (ctrl *StatisticsController) GetUsers(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	
	result := ctrl.statisticsService.GetUserStatistics(startDate, endDate)
	utils.Success(c, result)
}

// GetTrends 获取趋势统计
func (ctrl *StatisticsController) GetTrends(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	
	result := ctrl.statisticsService.GetTrendsStatistics(days)
	utils.Success(c, result)
}

// GetRealtime 获取实时统计
func (ctrl *StatisticsController) GetRealtime(c *gin.Context) {
	result := ctrl.statisticsService.GetRealtimeStatistics()
	utils.Success(c, result)
}

// GetTopPerformers 获取表现最佳的项目
func (ctrl *StatisticsController) GetTopPerformers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	result := gin.H{
		"products": ctrl.statisticsService.GetTopProducts(limit),
		"campaigns": ctrl.statisticsService.GetTopCampaigns(limit),
		"customers": ctrl.statisticsService.GetTopCustomers(limit),
	}
	
	utils.Success(c, result)
}

// GetComparison 获取对比统计
func (ctrl *StatisticsController) GetComparison(c *gin.Context) {
	var req struct {
		Type       string   `json:"type" binding:"required"` // product, campaign, revenue
		IDs        []uint   `json:"ids"`
		StartDate  string   `json:"start_date"`
		EndDate    string   `json:"end_date"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	
	result := ctrl.statisticsService.GetComparisonStatistics(req.Type, req.IDs, req.StartDate, req.EndDate)
	utils.Success(c, result)
}

// GetForecast 获取预测统计
func (ctrl *StatisticsController) GetForecast(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	
	result := ctrl.statisticsService.GetForecastStatistics(days)
	utils.Success(c, result)
}

// GetCustomReport 获取自定义报表
func (ctrl *StatisticsController) GetCustomReport(c *gin.Context) {
	var req struct {
		Metrics    []string `json:"metrics" binding:"required"`
		Dimensions []string `json:"dimensions"`
		Filters    map[string]interface{} `json:"filters"`
		StartDate  string   `json:"start_date"`
		EndDate    string   `json:"end_date"`
		GroupBy    string   `json:"group_by"`
		OrderBy    string   `json:"order_by"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	
	result := ctrl.statisticsService.GenerateCustomReport(req.Metrics, req.Dimensions, req.Filters, req.StartDate, req.EndDate)
	utils.Success(c, result)
}

// ExportReport 导出报表
func (ctrl *StatisticsController) ExportReport(c *gin.Context) {
	var req struct {
		Type      string `json:"type" binding:"required"` // pdf, excel, csv
		Report    string `json:"report" binding:"required"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	
	// 生成报表文件
	filePath, err := ctrl.statisticsService.ExportReport(req.Type, req.Report, req.StartDate, req.EndDate)
	if err != nil {
		utils.ServerError(c, "报表导出失败")
		return
	}
	
	utils.Success(c, gin.H{"file_path": filePath})
}

// GetMetrics 获取可用指标列表
func (ctrl *StatisticsController) GetMetrics(c *gin.Context) {
	metrics := []gin.H{
		{"key": "revenue", "name": "收入", "unit": "元"},
		{"key": "orders", "name": "订单数", "unit": "个"},
		{"key": "customers", "name": "客户数", "unit": "人"},
		{"key": "products", "name": "产品数", "unit": "个"},
		{"key": "campaigns", "name": "计划数", "unit": "个"},
		{"key": "coupons_used", "name": "优惠券使用", "unit": "张"},
		{"key": "conversion_rate", "name": "转化率", "unit": "%"},
		{"key": "avg_order_value", "name": "平均订单金额", "unit": "元"},
	}
	
	utils.Success(c, metrics)
}

// GetDimensions 获取可用维度列表
func (ctrl *StatisticsController) GetDimensions(c *gin.Context) {
	dimensions := []gin.H{
		{"key": "date", "name": "日期"},
		{"key": "product", "name": "产品"},
		{"key": "campaign", "name": "计划"},
		{"key": "customer_type", "name": "客户类型"},
		{"key": "region", "name": "地区"},
		{"key": "channel", "name": "渠道"},
	}
	
	utils.Success(c, dimensions)
}