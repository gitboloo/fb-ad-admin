package services

import (
	"backend/models"
	"backend/repositories"
	"backend/types"
	"fmt"
	"time"
)

// CampaignService 计划服务
type CampaignService struct {
	campaignRepo *repositories.CampaignRepository
}

// NewCampaignService 创建计划服务
func NewCampaignService() *CampaignService {
	return &CampaignService{
		campaignRepo: repositories.NewCampaignRepository(),
	}
}

// List 获取计划列表
func (cs *CampaignService) List(req *types.FilterRequest, productID *uint) ([]*models.Campaign, int64, error) {
	return cs.campaignRepo.List(req, productID)
}

// GetByID 根据ID获取计划
func (cs *CampaignService) GetByID(id uint) (*models.Campaign, error) {
	return cs.campaignRepo.GetByID(id)
}

// Create 创建计划
func (cs *CampaignService) Create(campaign *models.Campaign) error {
	// 如果没有计划编号，自动生成
	if campaign.CampaignNumber == "" {
		campaign.CampaignNumber = cs.generateCampaignNumber()
	}
	return cs.campaignRepo.Create(campaign)
}

// generateCampaignNumber 生成计划编号
func (cs *CampaignService) generateCampaignNumber() string {
	// 格式：CP + 年月日 + 时分秒 + 3位随机数
	// 例如：CP20250905164531001
	now := time.Now()
	return fmt.Sprintf("CP%s%03d",
		now.Format("20060102150405"),
		now.Nanosecond()/1000000%1000,
	)
}

// Update 更新计划
func (cs *CampaignService) Update(campaign *models.Campaign) error {
	// 如果计划编号为空，自动生成
	if campaign.CampaignNumber == "" {
		campaign.CampaignNumber = cs.generateCampaignNumber()
	}
	return cs.campaignRepo.Update(campaign)
}

// Delete 删除计划
func (cs *CampaignService) Delete(id uint) error {
	// 检查计划是否存在
	_, err := cs.campaignRepo.GetByID(id)
	if err != nil {
		return err
	}

	return cs.campaignRepo.Delete(id)
}

// UpdateStatus 更新计划状态
func (cs *CampaignService) UpdateStatus(id uint, status models.CampaignStatus) error {
	campaign, err := cs.campaignRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 状态变更验证
	if err := cs.validateStatusChange(campaign, status); err != nil {
		return err
	}

	campaign.Status = status
	return cs.campaignRepo.Update(campaign)
}

// GetCampaignStats 获取计划统计
func (cs *CampaignService) GetCampaignStats(id uint) (map[string]interface{}, error) {
	campaign, err := cs.campaignRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"campaign_id":         campaign.ID,
		"campaign_name":       campaign.Name,
		"status":              campaign.Status,
		"is_active":           campaign.IsActive(),
		"delivery_content":    campaign.DeliveryContent,
		"delivery_rules":      campaign.DeliveryRules,
		"user_targeting":      campaign.UserTargeting,
		"impressions":         0,
		"clicks":              0,
		"conversions":         0,
		"click_rate":          0.0,
		"conversion_rate":     0.0,
		"cost_per_click":      0.0,
		"cost_per_conversion": 0.0,
		"spend":               0.0,
	}

	// TODO: 这里应该从实际的数据分析系统获取展示、点击、转化等数据
	// 暂时返回模拟数据
	stats["impressions"] = 1234
	stats["clicks"] = 56
	stats["conversions"] = 8
	stats["click_rate"] = 4.54
	stats["conversion_rate"] = 14.29
	stats["cost_per_click"] = 1.25
	stats["cost_per_conversion"] = 8.75
	stats["spend"] = 70.0

	return stats, nil
}

// GetActiveCampaigns 获取活动计划列表
func (cs *CampaignService) GetActiveCampaigns() ([]*models.Campaign, error) {
	return cs.campaignRepo.GetByStatus(models.CampaignStatusActive)
}

// GetCampaignsByProduct 获取产品的计划列表
func (cs *CampaignService) GetCampaignsByProduct(productID uint) ([]*models.Campaign, error) {
	return cs.campaignRepo.GetByProductID(productID)
}

// PauseCampaign 暂停计划
func (cs *CampaignService) PauseCampaign(id uint) error {
	return cs.UpdateStatus(id, models.CampaignStatusPaused)
}

// ResumeCampaign 恢复计划
func (cs *CampaignService) ResumeCampaign(id uint) error {
	return cs.UpdateStatus(id, models.CampaignStatusActive)
}

// EndCampaign 结束计划
func (cs *CampaignService) EndCampaign(id uint) error {
	return cs.UpdateStatus(id, models.CampaignStatusEnded)
}

// BatchUpdateStatus 批量更新状态
func (cs *CampaignService) BatchUpdateStatus(ids []uint, status models.CampaignStatus) error {
	return cs.campaignRepo.BatchUpdateStatus(ids, status)
}

// GetStatistics 获取计划总体统计
func (cs *CampaignService) GetStatistics() (*types.StatisticsResponse, error) {
	return cs.campaignRepo.GetStatistics()
}

// validateStatusChange 验证状态变更
func (cs *CampaignService) validateStatusChange(campaign *models.Campaign, newStatus models.CampaignStatus) error {
	currentStatus := campaign.Status

	// 已结束的计划不能更改状态
	if currentStatus == models.CampaignStatusEnded {
		return &ServiceError{
			Code:    400,
			Message: "已结束的计划不能更改状态",
		}
	}

	return nil
}
