package admin

import (
	"strconv"

	"backend/models"
	"backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AgentHandler 管理员-代理商管理
type AgentHandler struct {
	db *gorm.DB
}

// NewAgentHandler 创建代理商管理handler
func NewAgentHandler(db *gorm.DB) *AgentHandler {
	return &AgentHandler{db: db}
}

// CreateAgentRequest 创建代理商请求
type CreateAgentRequest struct {
	// Admin 账户信息
	Username string `json:"username" binding:"required,min=3,max=20"` // 昵称/显示名称
	Account  string `json:"account" binding:"required"`               // 登录账号
	Password string `json:"password" binding:"required,min=6"`        // 密码

	// Agent 信息
	AgentLevel int    `json:"agent_level" binding:"required,min=1,max=3"`
	Remark     string `json:"remark"`
}

// List 获取代理商列表
// GET /api/admin/agents
func (h *AgentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	offset := (page - 1) * pageSize
	query := h.db.Model(&models.Agent{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var agents []models.Agent
	var total int64
	query.Count(&total)
	query.Preload("Admin").Preload("ParentAgent").Preload("Children").Offset(offset).Limit(pageSize).Find(&agents)

	utils.Success(c, gin.H{
		"list":  agents,
		"total": total,
	})
}

// Create 创建代理商（同时创建Admin账户和Agent信息）
// POST /api/admin/agents
func (h *AgentHandler) Create(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 获取当前登录管理员ID作为上级代理
	currentAdminID, exists := c.Get("admin_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	parentAdminID := currentAdminID.(uint)

	// 检查用户名是否已存在
	var existingAdmin models.Admin
	if err := h.db.Where("username = ?", req.Username).First(&existingAdmin).Error; err == nil {
		utils.BadRequest(c, "用户名已存在")
		return
	}

	// 开启事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 创建 Admin 账户
	// 注意: Admin模型的BeforeCreate钩子会自动加密密码
	admin := models.Admin{
		Username: req.Account,  // 登录账号
		Account:  req.Username, // 昵称/显示名称
		Password: req.Password, // 原始密码,BeforeCreate会自动加密
		Status:   models.AdminStatusActive,
	}

	if err := tx.Create(&admin).Error; err != nil {
		tx.Rollback()
		utils.ServerError(c, "创建Admin账户失败")
		return
	}

	// 2. 创建 Agent 信息
	agent := models.Agent{
		AdminID:       admin.ID,
		AgentLevel:    models.AgentLevel(req.AgentLevel),
		ParentAdminID: &parentAdminID, // 使用当前登录管理员ID作为上级
		Status:        models.AgentStatusActive,
		Remark:        req.Remark,
	}

	if err := tx.Create(&agent).Error; err != nil {
		tx.Rollback()
		utils.ServerError(c, "创建Agent信息失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		utils.ServerError(c, "提交失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Admin").First(&agent, agent.ID)

	utils.Success(c, gin.H{
		"data":    agent,
		"message": "创建成功",
	})
}

// Detail 获取代理商详情
// GET /api/admin/agents/:id
func (h *AgentHandler) Detail(c *gin.Context) {
	id := c.Param("id")

	var agent models.Agent
	if err := h.db.Preload("Admin").Preload("ParentAgent").Preload("Children").First(&agent, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "代理商不存在")
		} else {
			utils.ServerError(c, "查询失败")
		}
		return
	}

	utils.Success(c, gin.H{"data": agent})
}

// UpdateRequest 更新代理商请求
type UpdateRequest struct {
	AgentLevel                int    `json:"agent_level" binding:"min=1,max=3"`
	Status                    int    `json:"status" binding:"min=0,max=1"`
	EnableGoogleAuth          bool   `json:"enable_google_auth"`
	CanDispatchOrders         bool   `json:"can_dispatch_orders"`
	CanModifyCustomerBankCard bool   `json:"can_modify_customer_bank_card"`
	Remark                    string `json:"remark"`
}

// Update 更新代理商
// PUT /api/admin/agents/:id
func (h *AgentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	var agent models.Agent
	if err := h.db.First(&agent, id).Error; err != nil {
		utils.NotFound(c, "代理商不存在")
		return
	}

	updates := map[string]interface{}{
		"agent_level":                   req.AgentLevel,
		"status":                        req.Status,
		"enable_google_auth":            req.EnableGoogleAuth,
		"can_dispatch_orders":           req.CanDispatchOrders,
		"can_modify_customer_bank_card": req.CanModifyCustomerBankCard,
		"remark":                        req.Remark,
	}

	if err := h.db.Model(&agent).Updates(updates).Error; err != nil {
		utils.ServerError(c, "更新失败")
		return
	}

	utils.Success(c, gin.H{"data": agent})
}

// Delete 删除代理商
// DELETE /api/admin/agents/:id
func (h *AgentHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.db.Delete(&models.Agent{}, id).Error; err != nil {
		utils.ServerError(c, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}
