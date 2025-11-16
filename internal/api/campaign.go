package api

import (
	"strconv"

	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/services"
	"github.com/ad-platform/backend/internal/types"
	"github.com/ad-platform/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// CampaignRequest 计划请求结构
type CampaignRequest struct {
	Name            string                    `json:"name" binding:"required,max=255"`
	ProductID       uint                      `json:"product_id" binding:"required,min=1"`
	Description     string                    `json:"description"`
	Status          models.CampaignStatus     `json:"status" binding:"min=0,max=3"`
	DeliveryContent *models.DeliveryContent   `json:"delivery_content" binding:"required"`
	DeliveryRules   *models.DeliveryRules     `json:"delivery_rules" binding:"required"`
	UserTargeting   *models.UserTargeting     `json:"user_targeting" binding:"required"`
}

// CampaignController 计划控制器
type CampaignController struct {
	campaignService *services.CampaignService
	productService  *services.ProductService
}

// NewCampaignController 创建计划控制器
func NewCampaignController() *CampaignController {
	return &CampaignController{
		campaignService: services.NewCampaignService(),
		productService:  services.NewProductService(),
	}
}

// List 获取计划列表
func (cc *CampaignController) List(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 获取产品ID筛选
	productIDStr := c.Query("product_id")
	var productID *uint
	if productIDStr != "" {
		if id, err := strconv.ParseUint(productIDStr, 10, 32); err == nil {
			uid := uint(id)
			productID = &uid
		}
	}

	campaigns, total, err := cc.campaignService.List(&req, productID)
	if err != nil {
		utils.InternalServerError(c, "获取计划列表失败")
		return
	}

	utils.PagedSuccess(c, campaigns, total, req.GetPage(), req.GetSize())
}

// GetByID 获取计划详情
func (cc *CampaignController) GetByID(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	campaign, err := cc.campaignService.GetByID(req.ID)
	if err != nil {
		if err.Error() == "计划不存在" {
			utils.NotFound(c, "计划不存在")
		} else {
			utils.InternalServerError(c, "获取计划详情失败")
		}
		return
	}

	utils.Success(c, campaign)
}

// Create 创建计划
func (cc *CampaignController) Create(c *gin.Context) {
	var req CampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 验证产品是否存在
	product, err := cc.productService.GetByID(req.ProductID)
	if err != nil {
		if err.Error() == "产品不存在" {
			utils.BadRequest(c, "指定的产品不存在")
		} else {
			utils.InternalServerError(c, "验证产品失败")
		}
		return
	}

	// 检查产品是否激活
	if !product.IsActive() {
		utils.BadRequest(c, "只能为激活状态的产品创建计划")
		return
	}

	campaign := &models.Campaign{
		Name:            req.Name,
		ProductID:       req.ProductID,
		Description:     req.Description,
		Status:          req.Status,
		DeliveryContent: req.DeliveryContent,
		DeliveryRules:   req.DeliveryRules,
		UserTargeting:   req.UserTargeting,
	}

	if err := cc.campaignService.Create(campaign); err != nil {
		utils.InternalServerError(c, "创建计划失败")
		return
	}

	utils.Created(c, campaign)
}

// Update 更新计划
func (cc *CampaignController) Update(c *gin.Context) {
	var uriReq types.IDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.ValidateError(c, err)
		return
	}

	var req CampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 检查计划是否存在
	campaign, err := cc.campaignService.GetByID(uriReq.ID)
	if err != nil {
		if err.Error() == "计划不存在" {
			utils.NotFound(c, "计划不存在")
		} else {
			utils.InternalServerError(c, "获取计划信息失败")
		}
		return
	}

	// 验证产品是否存在
	if req.ProductID != campaign.ProductID {
		product, err := cc.productService.GetByID(req.ProductID)
		if err != nil {
			if err.Error() == "产品不存在" {
				utils.BadRequest(c, "指定的产品不存在")
			} else {
				utils.InternalServerError(c, "验证产品失败")
			}
			return
		}

		if !product.IsActive() {
			utils.BadRequest(c, "只能关联激活状态的产品")
			return
		}
	}

	// 更新字段
	campaign.Name = req.Name
	campaign.ProductID = req.ProductID
	campaign.Description = req.Description
	campaign.Status = req.Status
	campaign.DeliveryContent = req.DeliveryContent
	campaign.DeliveryRules = req.DeliveryRules
	campaign.UserTargeting = req.UserTargeting

	if err := cc.campaignService.Update(campaign); err != nil {
		utils.InternalServerError(c, "更新计划失败")
		return
	}

	utils.Updated(c, campaign)
}

// Delete 删除计划
func (cc *CampaignController) Delete(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := cc.campaignService.Delete(req.ID); err != nil {
		if err.Error() == "计划不存在" {
			utils.NotFound(c, "计划不存在")
		} else {
			utils.InternalServerError(c, "删除计划失败")
		}
		return
	}

	utils.Deleted(c)
}

// UploadLogo 上传计划Logo
func (cc *CampaignController) UploadLogo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的计划ID")
		return
	}

	// 检查计划是否存在
	campaign, err := cc.campaignService.GetByID(uint(id))
	if err != nil {
		if err.Error() == "计划不存在" {
			utils.NotFound(c, "计划不存在")
		} else {
			utils.InternalServerError(c, "获取计划信息失败")
		}
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("logo")
	if err != nil {
		utils.BadRequest(c, "请选择要上传的Logo文件")
		return
	}

	// 保存文件
	uploadResp, err := utils.SaveUploadedFile(c, file, "campaigns/logos")
	if err != nil {
		utils.BadRequest(c, "上传失败: "+err.Error())
		return
	}

	// 删除旧Logo文件
	if campaign.Logo != "" {
		utils.DeleteFile(campaign.Logo)
	}

	// 更新计划Logo字段
	campaign.Logo = uploadResp.URL
	if err := cc.campaignService.Update(campaign); err != nil {
		utils.InternalServerError(c, "更新计划Logo失败")
		return
	}

	utils.Success(c, uploadResp)
}

// GetStatistics 获取计划统计
func (cc *CampaignController) GetStatistics(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	stats, err := cc.campaignService.GetCampaignStats(req.ID)
	if err != nil {
		if err.Error() == "计划不存在" {
			utils.NotFound(c, "计划不存在")
		} else {
			utils.InternalServerError(c, "获取计划统计失败")
		}
		return
	}

	utils.Success(c, stats)
}

// UpdateStatus 更新计划状态
func (cc *CampaignController) UpdateStatus(c *gin.Context) {
	var uriReq types.IDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.ValidateError(c, err)
		return
	}

	var req types.StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := cc.campaignService.UpdateStatus(uriReq.ID, models.CampaignStatus(req.Status)); err != nil {
		if err.Error() == "计划不存在" {
			utils.NotFound(c, "计划不存在")
		} else {
			utils.InternalServerError(c, "更新计划状态失败")
		}
		return
	}

	utils.SuccessWithMessage(c, "状态更新成功", nil)
}

// Pause 暂停计划
func (cc *CampaignController) Pause(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := cc.campaignService.UpdateStatus(req.ID, models.CampaignStatusPaused); err != nil {
		if err.Error() == "计划不存在" {
			utils.NotFound(c, "计划不存在")
		} else {
			utils.InternalServerError(c, "暂停计划失败")
		}
		return
	}

	utils.SuccessWithMessage(c, "计划已暂停", nil)
}

// Resume 恢复计划
func (cc *CampaignController) Resume(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := cc.campaignService.UpdateStatus(req.ID, models.CampaignStatusActive); err != nil {
		if err.Error() == "计划不存在" {
			utils.NotFound(c, "计划不存在")
		} else {
			utils.InternalServerError(c, "恢复计划失败")
		}
		return
	}

	utils.SuccessWithMessage(c, "计划已恢复", nil)
}