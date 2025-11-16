package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============================================
// 代理商认证接口
// ============================================

// AgentLoginRequest 代理商登录请求
type AgentLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AgentLogin 代理商登录
// POST /api/agent/auth/login
func AgentLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AgentLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 查找代理商
		var agent models.Agent
		if err := db.Where("username = ? OR account = ?", req.Username, req.Username).
			First(&agent).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}

		// 验证密码
		if !agent.CheckPassword(req.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}

		// 检查账户状态
		if !agent.IsActive() {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "账户未激活或已被冻结",
				"status": agent.GetStatusString(),
			})
			return
		}

		// 记录登录时间
		agent.RecordLogin()
		db.Save(&agent)

		// 生成JWT Token
		token, err := utils.GenerateJWT(agent.ID, agent.Username, int(agent.AgentLevel))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "登录成功",
			"token":   token,
			"agent": gin.H{
				"id":          agent.ID,
				"username":    agent.Username,
				"real_name":   agent.RealName,
				"agent_level": agent.AgentLevel,
				"agent_code":  agent.AgentCode,
				"balance":     agent.Balance,
				"status":      agent.Status,
			},
		})
	}
}

// GetAgentProfile 获取代理商个人信息
// GET /api/agent/profile
func GetAgentProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, _ := c.Get("user_id")

		var agent models.Agent
		if err := db.Preload("Parent").First(&agent, agentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": agent,
		})
	}
}

// UpdateAgentProfile 更新代理商个人信息
// PUT /api/agent/profile
func UpdateAgentProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, _ := c.Get("user_id")

		var agent models.Agent
		if err := db.First(&agent, agentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}

		var input struct {
			RealName string `json:"real_name"`
			Phone    string `json:"phone"`
			Email    string `json:"email"`
			Company  string `json:"company"`
			Address  string `json:"address"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 更新字段
		if input.RealName != "" {
			agent.RealName = input.RealName
		}
		if input.Phone != "" {
			agent.Phone = input.Phone
		}
		if input.Email != "" {
			agent.Email = input.Email
		}
		if input.Company != "" {
			agent.Company = input.Company
		}
		if input.Address != "" {
			agent.Address = input.Address
		}

		if err := db.Save(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "更新成功",
			"data":    agent,
		})
	}
}

// ============================================
// 代理商客户管理
// ============================================

// GetAgentCustomers 获取代理商的客户列表
// GET /api/agent/customers
func GetAgentCustomers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, _ := c.Get("user_id")

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		offset := (page - 1) * pageSize

		var total int64
		var agentCustomers []models.AgentCustomer

		// 统计总数
		db.Model(&models.AgentCustomer{}).Where("agent_id = ?", agentID).Count(&total)

		// 查询客户列表（预加载客户信息）
		if err := db.Preload("Customer").
			Where("agent_id = ?", agentID).
			Offset(offset).
			Limit(pageSize).
			Find(&agentCustomers).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": agentCustomers,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	}
}

// ============================================
// 佣金管理
// ============================================

// GetAgentCommissions 获取代理商佣金列表
// GET /api/agent/commissions
func GetAgentCommissions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, _ := c.Get("user_id")

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		status := c.Query("status") // 0=待结算, 1=已结算

		offset := (page - 1) * pageSize

		query := db.Model(&models.Commission{}).Where("agent_id = ?", agentID)

		if status != "" {
			query = query.Where("status = ?", status)
		}

		var total int64
		query.Count(&total)

		var commissions []models.Commission
		if err := query.Preload("Customer").
			Order("created_at DESC").
			Offset(offset).
			Limit(pageSize).
			Find(&commissions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": commissions,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	}
}

// GetCommissionSummary 获取佣金汇总统计
// GET /api/agent/commissions/summary
func GetCommissionSummary(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, _ := c.Get("user_id")

		// 查询代理商信息
		var agent models.Agent
		db.First(&agent, agentID)

		// 统计待结算佣金
		var pendingCommission float64
		db.Model(&models.Commission{}).
			Where("agent_id = ? AND status = 0", agentID).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&pendingCommission)

		// 统计今日佣金
		today := time.Now().Format("2006-01-02")
		var todayCommission float64
		db.Model(&models.Commission{}).
			Where("agent_id = ? AND DATE(created_at) = ?", agentID, today).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&todayCommission)

		// 统计本月佣金
		thisMonth := time.Now().Format("2006-01")
		var monthCommission float64
		db.Model(&models.Commission{}).
			Where("agent_id = ? AND DATE_FORMAT(created_at, '%Y-%m') = ?", agentID, thisMonth).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&monthCommission)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"balance":            agent.Balance,
				"total_commission":   agent.TotalCommission,
				"pending_commission": pendingCommission,
				"today_commission":   todayCommission,
				"month_commission":   monthCommission,
				"frozen_balance":     agent.FrozenBalance,
			},
		})
	}
}

// ============================================
// 提现管理
// ============================================

// CreateWithdrawalRequest 申请提现请求
type CreateWithdrawalRequest struct {
	Amount           float64 `json:"amount" binding:"required,gt=0"`
	WithdrawalMethod int     `json:"withdrawal_method" binding:"required,min=1,max=3"`
	AccountName      string  `json:"account_name" binding:"required"`
	AccountNumber    string  `json:"account_number" binding:"required"`
	BankName         string  `json:"bank_name"`
}

// CreateWithdrawal 申请提现
// POST /api/agent/withdraw
func CreateWithdrawal(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, _ := c.Get("user_id")

		var req CreateWithdrawalRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 最低提现金额检查
		if req.Amount < 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "最低提现金额为100元"})
			return
		}

		// 查询代理商
		var agent models.Agent
		if err := db.First(&agent, agentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}

		// 检查余额
		if !agent.CanWithdraw(req.Amount) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "余额不足或账户状态异常",
				"balance": agent.GetAvailableBalance(),
			})
			return
		}

		// 计算手续费（2%）
		fee := req.Amount * 0.02
		actualAmount := req.Amount - fee

		// 开启事务
		err := db.Transaction(func(tx *gorm.DB) error {
			// 冻结余额
			if err := agent.FreezeAmount(req.Amount); err != nil {
				return err
			}
			if err := tx.Save(&agent).Error; err != nil {
				return err
			}

			// 创建提现记录
			withdrawal := models.Withdrawal{
				AgentID:          agent.ID,
				Amount:           req.Amount,
				Fee:              fee,
				ActualAmount:     actualAmount,
				WithdrawalMethod: models.WithdrawalMethod(req.WithdrawalMethod),
				AccountName:      req.AccountName,
				AccountNumber:    req.AccountNumber,
				BankName:         req.BankName,
				Status:           0, // 待审核
			}

			return tx.Create(&withdrawal).Error
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "提现申请失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "提现申请已提交，等待审核",
			"data": gin.H{
				"amount":        req.Amount,
				"fee":           fee,
				"actual_amount": actualAmount,
			},
		})
	}
}

// GetAgentWithdrawals 获取提现记录
// GET /api/agent/withdrawals
func GetAgentWithdrawals(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, _ := c.Get("user_id")

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		offset := (page - 1) * pageSize

		var total int64
		var withdrawals []models.Withdrawal

		db.Model(&models.Withdrawal{}).Where("agent_id = ?", agentID).Count(&total)

		if err := db.Where("agent_id = ?", agentID).
			Order("created_at DESC").
			Offset(offset).
			Limit(pageSize).
			Find(&withdrawals).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": withdrawals,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	}
}

// ============================================
// 代理商仪表板
// ============================================

// GetAgentDashboard 获取代理商仪表板数据
// GET /api/agent/dashboard
func GetAgentDashboard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, _ := c.Get("user_id")

		// 查询代理商信息
		var agent models.Agent
		db.First(&agent, agentID)

		// 统计数据
		var stats struct {
			TotalCustomers  int64
			ActiveCustomers int64
			TotalOrders     int64
			TotalSales      float64
			PendingCommission float64
			TodayCommission   float64
			MonthCommission   float64
		}

		// 客户统计
		db.Model(&models.AgentCustomer{}).Where("agent_id = ?", agentID).Count(&stats.TotalCustomers)

		// 待结算佣金
		db.Model(&models.Commission{}).
			Where("agent_id = ? AND status = 0", agentID).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&stats.PendingCommission)

		// 今日佣金
		today := time.Now().Format("2006-01-02")
		db.Model(&models.Commission{}).
			Where("agent_id = ? AND DATE(created_at) = ?", agentID, today).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&stats.TodayCommission)

		// 本月佣金
		thisMonth := time.Now().Format("2006-01")
		db.Model(&models.Commission{}).
			Where("agent_id = ? AND DATE_FORMAT(created_at, '%Y-%m') = ?", agentID, thisMonth).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&stats.MonthCommission)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"agent": gin.H{
					"username":    agent.Username,
					"real_name":   agent.RealName,
					"agent_code":  agent.AgentCode,
					"agent_level": agent.GetLevelString(),
					"balance":     agent.Balance,
					"total_commission": agent.TotalCommission,
				},
				"stats": stats,
			},
		})
	}
}
