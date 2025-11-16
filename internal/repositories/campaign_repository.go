package repositories

import (
	"fmt"
	"strings"

	"github.com/ad-platform/backend/internal/database"
	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/types"
	"gorm.io/gorm"
)

// CampaignRepository 计划仓库
type CampaignRepository struct {
	db *gorm.DB
}

// NewCampaignRepository 创建计划仓库
func NewCampaignRepository() *CampaignRepository {
	return &CampaignRepository{
		db: database.DB,
	}
}

// List 获取计划列表
func (cr *CampaignRepository) List(req *types.FilterRequest, productID *uint) ([]*models.Campaign, int64, error) {
	var campaigns []*models.Campaign
	var total int64

	query := cr.db.Model(&models.Campaign{}).Preload("Product")

	// 产品筛选
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	}

	// 搜索条件
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 日期范围筛选
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", req.EndDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序和分页
	orderClause := fmt.Sprintf("%s %s", req.GetSort(), req.GetOrder())
	if err := query.Order(orderClause).
		Offset(req.GetOffset()).
		Limit(req.GetSize()).
		Find(&campaigns).Error; err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

// GetByID 根据ID获取计划
func (cr *CampaignRepository) GetByID(id uint) (*models.Campaign, error) {
	var campaign models.Campaign
	if err := cr.db.Preload("Product").First(&campaign, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("计划不存在")
		}
		return nil, err
	}
	return &campaign, nil
}

// Create 创建计划
func (cr *CampaignRepository) Create(campaign *models.Campaign) error {
	return cr.db.Create(campaign).Error
}

// Update 更新计划
func (cr *CampaignRepository) Update(campaign *models.Campaign) error {
	return cr.db.Save(campaign).Error
}

// Delete 删除计划
func (cr *CampaignRepository) Delete(id uint) error {
	result := cr.db.Delete(&models.Campaign{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("计划不存在")
	}
	return nil
}

// GetByStatus 根据状态获取计划
func (cr *CampaignRepository) GetByStatus(status models.CampaignStatus) ([]*models.Campaign, error) {
	var campaigns []*models.Campaign
	if err := cr.db.Preload("Product").Where("status = ?", status).Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

// GetByProductID 根据产品ID获取计划
func (cr *CampaignRepository) GetByProductID(productID uint) ([]*models.Campaign, error) {
	var campaigns []*models.Campaign
	if err := cr.db.Where("product_id = ?", productID).Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

// Search 搜索计划
func (cr *CampaignRepository) Search(keyword string, limit int) ([]*models.Campaign, error) {
	var campaigns []*models.Campaign
	searchPattern := "%" + strings.TrimSpace(keyword) + "%"
	
	query := cr.db.Preload("Product").Where("name LIKE ? OR description LIKE ?", 
		searchPattern, searchPattern)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

// BatchUpdateStatus 批量更新状态
func (cr *CampaignRepository) BatchUpdateStatus(ids []uint, status models.CampaignStatus) error {
	return cr.db.Model(&models.Campaign{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// GetStatistics 获取计划统计
func (cr *CampaignRepository) GetStatistics() (*types.StatisticsResponse, error) {
	var total int64
	var active int64
	var inactive int64

	// 总数统计
	if err := cr.db.Model(&models.Campaign{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// 活动计划数
	if err := cr.db.Model(&models.Campaign{}).
		Where("status = ?", models.CampaignStatusActive).Count(&active).Error; err != nil {
		return nil, err
	}

	// 非活动计划数
	if err := cr.db.Model(&models.Campaign{}).
		Where("status != ?", models.CampaignStatusActive).Count(&inactive).Error; err != nil {
		return nil, err
	}

	// 按状态统计
	var statusStats []struct {
		Status models.CampaignStatus `json:"status"`
		Count  int64                 `json:"count"`
	}
	if err := cr.db.Model(&models.Campaign{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusStats).Error; err != nil {
		return nil, err
	}

	categories := make(map[string]interface{})
	for _, stat := range statusStats {
		var statusName string
		switch stat.Status {
		case models.CampaignStatusActive:
			statusName = "活动中"
		case models.CampaignStatusPaused:
			statusName = "已暂停"
		case models.CampaignStatusEnded:
			statusName = "已结束"
		case models.CampaignStatusInactive:
			statusName = "未激活"
		default:
			statusName = "未知"
		}
		categories[statusName] = stat.Count
	}

	// 趋势数据（最近7天）
	var trendData []types.TrendData
	if err := cr.db.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as value 
		FROM campaigns 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) 
		GROUP BY DATE(created_at) 
		ORDER BY date ASC
	`).Scan(&trendData).Error; err != nil {
		return nil, err
	}

	// 计算增长率（与上周同期比较）
	var currentWeek int64
	var lastWeek int64
	
	cr.db.Model(&models.Campaign{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").Count(&currentWeek)
	cr.db.Model(&models.Campaign{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 14 DAY) AND created_at < DATE_SUB(NOW(), INTERVAL 7 DAY)").
		Count(&lastWeek)

	var growth float64
	if lastWeek > 0 {
		growth = float64(currentWeek-lastWeek) / float64(lastWeek) * 100
	}

	return &types.StatisticsResponse{
		Total:      total,
		Active:     active,
		Inactive:   inactive,
		Growth:     growth,
		TrendData:  trendData,
		Categories: categories,
	}, nil
}

// GetRunningCampaigns 获取正在运行的计划
func (cr *CampaignRepository) GetRunningCampaigns() ([]*models.Campaign, error) {
	var campaigns []*models.Campaign
	
	// 查询活动状态且在有效期内的计划
	if err := cr.db.Preload("Product").
		Where("status = ?", models.CampaignStatusActive).
		Where("JSON_EXTRACT(delivery_rules, '$.start_date') <= NOW()").
		Where("JSON_EXTRACT(delivery_rules, '$.end_date') >= NOW()").
		Find(&campaigns).Error; err != nil {
		return nil, err
	}
	
	return campaigns, nil
}

// GetCampaignsByDateRange 根据日期范围获取计划
func (cr *CampaignRepository) GetCampaignsByDateRange(startDate, endDate string) ([]*models.Campaign, error) {
	var campaigns []*models.Campaign
	
	if err := cr.db.Preload("Product").
		Where("JSON_EXTRACT(delivery_rules, '$.start_date') >= ?", startDate).
		Where("JSON_EXTRACT(delivery_rules, '$.end_date') <= ?", endDate).
		Find(&campaigns).Error; err != nil {
		return nil, err
	}
	
	return campaigns, nil
}

// GetExpiredCampaigns 获取已过期的计划
func (cr *CampaignRepository) GetExpiredCampaigns() ([]*models.Campaign, error) {
	var campaigns []*models.Campaign
	
	if err := cr.db.Preload("Product").
		Where("status IN ?", []models.CampaignStatus{models.CampaignStatusActive, models.CampaignStatusPaused}).
		Where("JSON_EXTRACT(delivery_rules, '$.end_date') < NOW()").
		Find(&campaigns).Error; err != nil {
		return nil, err
	}
	
	return campaigns, nil
}