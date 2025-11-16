package api

import (
	"net/http"
	"strconv"

	"github.com/ad-platform/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============================================
// 管理员 - 代理商管理接口
// ============================================

// GetAgentList 获取代理商列表（管理员）
// GET /api/admin/agents
func GetAgentList(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		status := c.Query("status")
		level := c.Query("level")
		keyword := c.Query("keyword")

		offset := (page - 1) * pageSize

		query := db.Model(&models.Agent{})

		// 过滤条件
		if status != "" {
			query = query.Where("status = ?", status)
		}
		if level != "" {
			query = query.Where("agent_level = ?", level)
		}
		if keyword != "" {
			query = query.Where("username LIKE ? OR real_name LIKE ? OR phone LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}

		var total int64
		query.Count(&total)

		var agents []models.Agent
		if err := query.Preload("Parent").
			Order("created_at DESC").
			Offset(offset).
			Limit(pageSize).
			Find(&agents).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": agents,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	}
}

// CreateAgentRequest 创建代理商请求
type CreateAgentRequest struct {
	Username           string  `json:"username" binding:"required"`
	Account            string  `json:"account" binding:"required,email"`
	Password           string  `json:"password" binding:"required,min=6"`
	RealName           string  `json:"real_name" binding:"required"`
	Phone              string  `json:"phone" binding:"required"`
	Email              string  `json:"email"`
	Company            string  `json:"company"`
	AgentLevel         int     `json:"agent_level" binding:"required,min=1,max=3"`
	ParentID           *uint   `json:"parent_id"`
	CommissionRate     float64 `json:"commission_rate" binding:"required,min=0,max=100"`
	SelfCommissionRate float64 `json:"self_commission_rate" binding:"min=0,max=100"`
}

// CreateAgent 创建代理商（管理员）
// POST /api/admin/agents
func CreateAgent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateAgentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 检查用户名是否已存在
		var existingAgent models.Agent
		if err := db.Where("username = ? OR account = ?", req.Username, req.Account).
			First(&existingAgent).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或账号已存在"})
			return
		}

		// 如果是二级/三级代理，检查上级是否存在
		if req.ParentID != nil {
			var parent models.Agent
			if err := db.First(&parent, req.ParentID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "上级代理商不存在"})
				return
			}
		}

		// 创建代理商
		agent := models.Agent{
			Username:           req.Username,
			Account:            req.Account,
			Password:           req.Password, // 会在BeforeCreate中自动加密
			RealName:           req.RealName,
			Phone:              req.Phone,
			Email:              req.Email,
			Company:            req.Company,
			AgentLevel:         models.AgentLevel(req.AgentLevel),
			ParentID:           req.ParentID,
			CommissionRate:     req.CommissionRate,
			SelfCommissionRate: req.SelfCommissionRate,
			Status:             models.AgentStatusPending, // 默认待审核
		}

		if err := db.Create(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "代理商创建成功",
			"data":    agent,
		})
	}
}

// GetAgentDetail 获取代理商详情（管理员）
// GET /api/admin/agents/:id
func GetAgentDetail(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var agent models.Agent
		if err := db.Preload("Parent").
			Preload("Children").
			First(&agent, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}

		// 统计客户数
		var customerCount int64
		db.Model(&models.AgentCustomer{}).Where("agent_id = ?", id).Count(&customerCount)

		// 统计佣金
		var totalCommission float64
		db.Model(&models.Commission{}).
			Where("agent_id = ?", id).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&totalCommission)

		c.JSON(http.StatusOK, gin.H{
			"data": agent,
			"stats": gin.H{
				"customer_count":   customerCount,
				"total_commission": totalCommission,
			},
		})
	}
}

// UpdateAgent 更新代理商信息（管理员）
// PUT /api/admin/agents/:id
func UpdateAgent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var agent models.Agent
		if err := db.First(&agent, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}

		var input struct {
			RealName           string   `json:"real_name"`
			Phone              string   `json:"phone"`
			Email              string   `json:"email"`
			Company            string   `json:"company"`
			CommissionRate     *float64 `json:"commission_rate"`
			SelfCommissionRate *float64 `json:"self_commission_rate"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 更新字段
		updates := make(map[string]interface{})
		if input.RealName != "" {
			updates["real_name"] = input.RealName
		}
		if input.Phone != "" {
			updates["phone"] = input.Phone
		}
		if input.Email != "" {
			updates["email"] = input.Email
		}
		if input.Company != "" {
			updates["company"] = input.Company
		}
		if input.CommissionRate != nil {
			updates["commission_rate"] = *input.CommissionRate
		}
		if input.SelfCommissionRate != nil {
			updates["self_commission_rate"] = *input.SelfCommissionRate
		}

		if err := db.Model(&agent).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "更新成功",
			"data":    agent,
		})
	}
}

// ApproveAgent 审核通过代理商（管理员）
// POST /api/admin/agents/:id/approve
func ApproveAgent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var agent models.Agent
		if err := db.First(&agent, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}

		agent.Approve()

		if err := db.Save(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "审核通过",
			"data":    agent,
		})
	}
}

// RejectAgent 审核拒绝代理商（管理员）
// POST /api/admin/agents/:id/reject
func RejectAgent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var input struct {
			Reason string `json:"reason" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var agent models.Agent
		if err := db.First(&agent, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}

		agent.Reject(input.Reason)

		if err := db.Save(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "已拒绝",
			"data":    agent,
		})
	}
}

// FreezeAgent 冻结代理商（管理员）
// POST /api/admin/agents/:id/freeze
func FreezeAgent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var agent models.Agent
		if err := db.First(&agent, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}

		agent.Freeze()

		if err := db.Save(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "代理商已冻结",
			"data":    agent,
		})
	}
}

// UnfreezeAgent 解冻代理商（管理员）
// POST /api/admin/agents/:id/unfreeze
func UnfreezeAgent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var agent models.Agent
		if err := db.First(&agent, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}

		agent.Unfreeze()

		if err := db.Save(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "代理商已解冻",
			"data":    agent,
		})
	}
}

// ============================================
// 提现审核（管理员）
// ============================================

// GetWithdrawalList 获取提现申请列表（管理员）
// GET /api/admin/withdrawals
func GetWithdrawalList(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		status := c.Query("status")

		offset := (page - 1) * pageSize

		query := db.Model(&models.Withdrawal{})

		if status != "" {
			query = query.Where("status = ?", status)
		}

		var total int64
		query.Count(&total)

		var withdrawals []models.Withdrawal
		if err := query.Preload("Agent").
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

// ApproveWithdrawal 审核通过提现（管理员）
// POST /api/admin/withdrawals/:id/approve
func ApproveWithdrawal(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		adminID, _ := c.Get("user_id")

		var withdrawal models.Withdrawal
		if err := db.First(&withdrawal, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "提现记录不存在"})
			return
		}

		if withdrawal.Status != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该提现申请已处理"})
			return
		}

		// 审核通过
		withdrawal.Approve(adminID.(uint))

		if err := db.Save(&withdrawal).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "审核通过",
			"data":    withdrawal,
		})
	}
}

// RejectWithdrawal 审核拒绝提现（管理员）
// POST /api/admin/withdrawals/:id/reject
func RejectWithdrawal(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		adminID, _ := c.Get("user_id")

		var input struct {
			Reason string `json:"reason" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var withdrawal models.Withdrawal
		if err := db.Preload("Agent").First(&withdrawal, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "提现记录不存在"})
			return
		}

		if withdrawal.Status != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该提现申请已处理"})
			return
		}

		// 开启事务
		err := db.Transaction(func(tx *gorm.DB) error {
			// 拒绝提现
			withdrawal.Reject(input.Reason, adminID.(uint))
			if err := tx.Save(&withdrawal).Error; err != nil {
				return err
			}

			// 解冻余额
			var agent models.Agent
			if err := tx.First(&agent, withdrawal.AgentID).Error; err != nil {
				return err
			}

			agent.UnfreezeAmount(withdrawal.Amount)
			return tx.Save(&agent).Error
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "已拒绝",
			"data":    withdrawal,
		})
	}
}

// CompleteWithdrawal 标记提现完成（管理员）
// POST /api/admin/withdrawals/:id/complete
func CompleteWithdrawal(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var withdrawal models.Withdrawal
		if err := db.Preload("Agent").First(&withdrawal, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "提现记录不存在"})
			return
		}

		if withdrawal.Status != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "只能完成已审核通过的提现"})
			return
		}

		// 开启事务
		err := db.Transaction(func(tx *gorm.DB) error {
			// 标记完成
			withdrawal.Complete()
			if err := tx.Save(&withdrawal).Error; err != nil {
				return err
			}

			// 扣除冻结余额
			var agent models.Agent
			if err := tx.First(&agent, withdrawal.AgentID).Error; err != nil {
				return err
			}

			agent.FrozenBalance -= withdrawal.Amount
			return tx.Save(&agent).Error
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "提现已完成",
			"data":    withdrawal,
		})
	}
}
