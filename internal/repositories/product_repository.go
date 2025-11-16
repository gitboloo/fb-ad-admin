package repositories

import (
	"fmt"
	"strings"

	"github.com/ad-platform/backend/internal/database"
	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/types"
	"gorm.io/gorm"
)

// ProductRepository 产品仓库
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建产品仓库
func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		db: database.DB,
	}
}

// List 获取产品列表
func (pr *ProductRepository) List(req *types.FilterRequest) ([]*models.Product, int64, error) {
	var products []*models.Product
	var total int64

	query := pr.db.Model(&models.Product{})

	// 搜索条件
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("name LIKE ? OR company LIKE ? OR description LIKE ?", 
			searchPattern, searchPattern, searchPattern)
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 分类筛选（产品类型）
	if req.Category != "" {
		query = query.Where("type = ?", req.Category)
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
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// GetByID 根据ID获取产品
func (pr *ProductRepository) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	if err := pr.db.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("产品不存在")
		}
		return nil, err
	}
	return &product, nil
}

// Create 创建产品
func (pr *ProductRepository) Create(product *models.Product) error {
	return pr.db.Create(product).Error
}

// Update 更新产品
func (pr *ProductRepository) Update(product *models.Product) error {
	return pr.db.Save(product).Error
}

// Delete 删除产品
func (pr *ProductRepository) Delete(id uint) error {
	result := pr.db.Delete(&models.Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("产品不存在")
	}
	return nil
}

// GetByStatus 根据状态获取产品
func (pr *ProductRepository) GetByStatus(status models.ProductStatus) ([]*models.Product, error) {
	var products []*models.Product
	if err := pr.db.Where("status = ?", status).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// GetByType 根据类型获取产品
func (pr *ProductRepository) GetByType(productType models.ProductType) ([]*models.Product, error) {
	var products []*models.Product
	if err := pr.db.Where("type = ?", productType).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// Search 搜索产品
func (pr *ProductRepository) Search(keyword string, limit int) ([]*models.Product, error) {
	var products []*models.Product
	searchPattern := "%" + strings.TrimSpace(keyword) + "%"
	
	query := pr.db.Where("name LIKE ? OR company LIKE ? OR description LIKE ?", 
		searchPattern, searchPattern, searchPattern)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// GetWithCampaigns 获取产品及其计划
func (pr *ProductRepository) GetWithCampaigns(id uint) (*models.Product, error) {
	var product models.Product
	if err := pr.db.Preload("Campaigns").First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("产品不存在")
		}
		return nil, err
	}
	return &product, nil
}

// GetCampaignsByProductID 获取产品的计划列表
func (pr *ProductRepository) GetCampaignsByProductID(productID uint) ([]*models.Campaign, error) {
	var campaigns []*models.Campaign
	if err := pr.db.Where("product_id = ?", productID).Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

// BatchUpdateStatus 批量更新状态
func (pr *ProductRepository) BatchUpdateStatus(ids []uint, status models.ProductStatus) error {
	return pr.db.Model(&models.Product{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// GetStatistics 获取产品统计
func (pr *ProductRepository) GetStatistics() (*types.StatisticsResponse, error) {
	var total int64
	var active int64
	var inactive int64

	// 总数统计
	if err := pr.db.Model(&models.Product{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// 活动产品数
	if err := pr.db.Model(&models.Product{}).
		Where("status = ?", models.ProductStatusActive).Count(&active).Error; err != nil {
		return nil, err
	}

	// 非活动产品数
	if err := pr.db.Model(&models.Product{}).
		Where("status != ?", models.ProductStatusActive).Count(&inactive).Error; err != nil {
		return nil, err
	}

	// 按类型统计
	var typeStats []struct {
		Type  models.ProductType `json:"type"`
		Count int64              `json:"count"`
	}
	if err := pr.db.Model(&models.Product{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Find(&typeStats).Error; err != nil {
		return nil, err
	}

	categories := make(map[string]interface{})
	for _, stat := range typeStats {
		var typeName string
		switch stat.Type {
		case models.ProductTypeApp:
			typeName = "应用"
		case models.ProductTypeGame:
			typeName = "游戏"
		case models.ProductTypeWeb:
			typeName = "网站"
		case models.ProductTypeOther:
			typeName = "其他"
		default:
			typeName = "未知"
		}
		categories[typeName] = stat.Count
	}

	// 趋势数据（最近7天）
	var trendData []types.TrendData
	if err := pr.db.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as value 
		FROM products 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) 
		GROUP BY DATE(created_at) 
		ORDER BY date ASC
	`).Scan(&trendData).Error; err != nil {
		return nil, err
	}

	// 计算增长率（与上周同期比较）
	var currentWeek int64
	var lastWeek int64
	
	pr.db.Model(&models.Product{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").Count(&currentWeek)
	pr.db.Model(&models.Product{}).
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