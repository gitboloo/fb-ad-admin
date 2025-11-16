package services

import (
	"time"

	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/repositories"
	"github.com/ad-platform/backend/internal/types"
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
	// 验证投放规则
	if err := cs.validateDeliveryRules(campaign.DeliveryRules); err != nil {
		return err
	}

	return cs.campaignRepo.Create(campaign)
}

// Update 更新计划
func (cs *CampaignService) Update(campaign *models.Campaign) error {
	// 验证投放规则
	if err := cs.validateDeliveryRules(campaign.DeliveryRules); err != nil {
		return err
	}

	return cs.campaignRepo.Update(campaign)
}

// Delete 删除计划
func (cs *CampaignService) Delete(id uint) error {
	// 检查计划是否存在
	campaign, err := cs.campaignRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 检查计划是否正在运行
	if campaign.IsRunning() {
		return &ServiceError{
			Code:    400,
			Message: "正在运行的计划无法删除，请先暂停或结束计划",
		}
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
		"campaign_id":       campaign.ID,
		"campaign_name":     campaign.Name,
		"status":            campaign.Status,
		"is_running":        campaign.IsRunning(),
		"remaining_budget":  campaign.GetRemainingBudget(),
		"total_budget":      0.0,
		"daily_budget":      0.0,
		"start_date":        nil,
		"end_date":          nil,
		"impressions":       0,
		"clicks":            0,
		"conversions":       0,
		"click_rate":        0.0,
		"conversion_rate":   0.0,
		"cost_per_click":    0.0,
		"cost_per_conversion": 0.0,
		"spend":             0.0,
	}

	if campaign.DeliveryRules != nil {
		stats["total_budget"] = campaign.DeliveryRules.TotalBudget
		stats["daily_budget"] = campaign.DeliveryRules.DailyBudget
		stats["start_date"] = campaign.DeliveryRules.StartDate
		stats["end_date"] = campaign.DeliveryRules.EndDate
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

// validateDeliveryRules 验证投放规则
func (cs *CampaignService) validateDeliveryRules(rules *models.DeliveryRules) error {
	if rules == nil {
		return &ServiceError{
			Code:    400,
			Message: "投放规则不能为空",
		}
	}

	// 验证日期
	if rules.StartDate.After(rules.EndDate) {
		return &ServiceError{
			Code:    400,
			Message: "开始日期不能晚于结束日期",
		}
	}

	// 验证预算
	if rules.TotalBudget <= 0 {
		return &ServiceError{
			Code:    400,
			Message: "总预算必须大于0",
		}
	}

	if rules.DailyBudget <= 0 {
		return &ServiceError{
			Code:    400,
			Message: "日预算必须大于0",
		}
	}

	if rules.DailyBudget > rules.TotalBudget {
		return &ServiceError{
			Code:    400,
			Message: "日预算不能超过总预算",
		}
	}

	// 验证出价
	if rules.BidAmount <= 0 {
		return &ServiceError{
			Code:    400,
			Message: "出价必须大于0",
		}
	}

	return nil
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

	// 检查日期限制
	if newStatus == models.CampaignStatusActive && campaign.DeliveryRules != nil {
		now := time.Now()
		if now.After(campaign.DeliveryRules.EndDate) {
			return &ServiceError{
				Code:    400,
				Message: "计划已过期，无法激活",
			}
		}
	}

	return nil
}