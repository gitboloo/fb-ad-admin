package admin

import (
	"backend/services"
	"backend/utils"
	"fmt"

	"github.com/gin-gonic/gin"
)

// DashboardHandler 仪表盘管理
type DashboardHandler struct {
	dashboardService *services.DashboardService
}

// NewDashboardHandler 创建仪表盘管理handler
func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{
		dashboardService: services.NewDashboardService(),
	}
}

// GetStats 获取仪表盘统计数据
// GET /api/dashboard/stats
func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats := gin.H{
		"new_customers_today":  h.dashboardService.GetNewCustomersToday(),
		"pending_orders":       h.dashboardService.GetPendingOrdersCount(),
		"conversion_rate":      h.dashboardService.GetConversionRate(),
		"average_order_value":  h.dashboardService.GetAverageOrderValue(),
		"active_users":         h.dashboardService.GetActiveUsersCount(),
		"returning_users_rate": h.dashboardService.GetReturningUsersRate(),
	}

	utils.Success(c, stats)
}

// GetActivities 获取最近活动
// GET /api/dashboard/activities
func (h *DashboardHandler) GetActivities(c *gin.Context) {
	limit := 10
	if l, ok := c.GetQuery("limit"); ok {
		if parsedLimit, err := parseLimit(l); err == nil {
			limit = parsedLimit
		}
	}

	activities := h.dashboardService.GetRecentActivities(limit)
	utils.Success(c, activities)
}

// GetRevenueTrend 获取收入趋势
// GET /api/dashboard/revenue-trend
func (h *DashboardHandler) GetRevenueTrend(c *gin.Context) {
	days := 7 // 默认7天
	if d := c.Query("days"); d != "" {
		if parsedDays, err := parseLimit(d); err == nil {
			days = parsedDays
		}
	}

	trend := h.dashboardService.GetRevenueTrend(days)
	utils.Success(c, trend)
}

// GetUserGrowth 获取用户增长
// GET /api/dashboard/user-growth
func (h *DashboardHandler) GetUserGrowth(c *gin.Context) {
	period := c.DefaultQuery("period", "week")
	growth := h.dashboardService.GetUserGrowth(period)
	utils.Success(c, growth)
}

// GetCampaignPerformance 获取广告计划表现
// GET /api/dashboard/campaign-performance
func (h *DashboardHandler) GetCampaignPerformance(c *gin.Context) {
	limit := 10
	if l, ok := c.GetQuery("limit"); ok {
		if parsedLimit, err := parseLimit(l); err == nil {
			limit = parsedLimit
		}
	}

	performance := h.dashboardService.GetCampaignPerformance(limit)
	utils.Success(c, performance)
}

// GetRealtime 获取实时数据
// GET /api/dashboard/realtime
func (h *DashboardHandler) GetRealtime(c *gin.Context) {
	data := h.dashboardService.GetRealtimeData()
	utils.Success(c, data)
}

// parseLimit 解析limit参数
func parseLimit(s string) (int, error) {
	var limit int
	_, err := fmt.Sscanf(s, "%d", &limit)
	if err != nil || limit <= 0 {
		return 0, fmt.Errorf("invalid limit")
	}
	if limit > 100 {
		limit = 100 // 最大限制
	}
	return limit, nil
}
