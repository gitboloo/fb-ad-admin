package repositories

import (
	"fmt"
	"strings"

	"github.com/ad-platform/backend/internal/database"
	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/types"
	"gorm.io/gorm"
)

// CouponRepository 优惠券仓库
type CouponRepository struct {
	db *gorm.DB
}

// NewCouponRepository 创建优惠券仓库
func NewCouponRepository() *CouponRepository {
	return &CouponRepository{
		db: database.DB,
	}
}

// List 获取优惠券列表
func (cr *CouponRepository) List(req *types.FilterRequest) ([]*models.Coupon, int64, error) {
	var coupons []*models.Coupon
	var total int64

	query := cr.db.Model(&models.Coupon{})

	// 搜索条件
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 分类筛选（优惠券类型）
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
		Find(&coupons).Error; err != nil {
		return nil, 0, err
	}

	return coupons, total, nil
}

// GetByID 根据ID获取优惠券
func (cr *CouponRepository) GetByID(id uint) (*models.Coupon, error) {
	var coupon models.Coupon
	if err := cr.db.First(&coupon, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("优惠券不存在")
		}
		return nil, err
	}
	return &coupon, nil
}

// Create 创建优惠券
func (cr *CouponRepository) Create(coupon *models.Coupon) error {
	return cr.db.Create(coupon).Error
}

// Update 更新优惠券
func (cr *CouponRepository) Update(coupon *models.Coupon) error {
	return cr.db.Save(coupon).Error
}

// Delete 删除优惠券
func (cr *CouponRepository) Delete(id uint) error {
	result := cr.db.Delete(&models.Coupon{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("优惠券不存在")
	}
	return nil
}

// GetByStatus 根据状态获取优惠券
func (cr *CouponRepository) GetByStatus(status models.CouponStatus) ([]*models.Coupon, error) {
	var coupons []*models.Coupon
	if err := cr.db.Where("status = ?", status).Find(&coupons).Error; err != nil {
		return nil, err
	}
	return coupons, nil
}

// GetByType 根据类型获取优惠券
func (cr *CouponRepository) GetByType(couponType models.CouponType) ([]*models.Coupon, error) {
	var coupons []*models.Coupon
	if err := cr.db.Where("type = ?", couponType).Find(&coupons).Error; err != nil {
		return nil, err
	}
	return coupons, nil
}

// Search 搜索优惠券
func (cr *CouponRepository) Search(keyword string, limit int) ([]*models.Coupon, error) {
	var coupons []*models.Coupon
	searchPattern := "%" + strings.TrimSpace(keyword) + "%"
	
	query := cr.db.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&coupons).Error; err != nil {
		return nil, err
	}
	return coupons, nil
}

// GetAvailable 获取可用优惠券
func (cr *CouponRepository) GetAvailable() ([]*models.Coupon, error) {
	var coupons []*models.Coupon
	if err := cr.db.Where("status = ?", models.CouponStatusActive).
		Where("(total_count = 0 OR used_count < total_count)").
		Find(&coupons).Error; err != nil {
		return nil, err
	}
	return coupons, nil
}

// GetStatistics 获取优惠券统计
func (cr *CouponRepository) GetStatistics() (*types.StatisticsResponse, error) {
	var total int64
	var active int64
	var inactive int64

	// 总数统计
	if err := cr.db.Model(&models.Coupon{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// 活动优惠券数
	if err := cr.db.Model(&models.Coupon{}).
		Where("status = ?", models.CouponStatusActive).Count(&active).Error; err != nil {
		return nil, err
	}

	// 非活动优惠券数
	if err := cr.db.Model(&models.Coupon{}).
		Where("status != ?", models.CouponStatusActive).Count(&inactive).Error; err != nil {
		return nil, err
	}

	// 按类型统计
	var typeStats []struct {
		Type  models.CouponType `json:"type"`
		Count int64             `json:"count"`
	}
	if err := cr.db.Model(&models.Coupon{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Find(&typeStats).Error; err != nil {
		return nil, err
	}

	categories := make(map[string]interface{})
	for _, stat := range typeStats {
		var typeName string
		switch stat.Type {
		case models.CouponTypeValueAdded:
			typeName = "增值券"
		case models.CouponTypeDiscount:
			typeName = "折扣券"
		case models.CouponTypeTeam:
			typeName = "团队券"
		case models.CouponTypeCustom:
			typeName = "自定义券"
		case models.CouponTypeFixed:
			typeName = "固定金额券"
		default:
			typeName = "其他"
		}
		categories[typeName] = stat.Count
	}

	// 趋势数据（最近7天）
	var trendData []types.TrendData
	if err := cr.db.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as value 
		FROM coupons 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) 
		GROUP BY DATE(created_at) 
		ORDER BY date ASC
	`).Scan(&trendData).Error; err != nil {
		return nil, err
	}

	// 计算增长率（与上周同期比较）
	var currentWeek int64
	var lastWeek int64
	
	cr.db.Model(&models.Coupon{}).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").Count(&currentWeek)
	cr.db.Model(&models.Coupon{}).
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

// BatchUpdateStatus 批量更新状态
func (cr *CouponRepository) BatchUpdateStatus(ids []uint, status models.CouponStatus) error {
	return cr.db.Model(&models.Coupon{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// GetExpiredCoupons 获取过期的优惠券
func (cr *CouponRepository) GetExpiredCoupons() ([]*models.Coupon, error) {
	var coupons []*models.Coupon
	
	if err := cr.db.Where("validity_type = ?", models.ValidityTypeRange).
		Where("JSON_EXTRACT(date_range, '$.end_date') < NOW()").
		Where("status = ?", models.CouponStatusActive).
		Find(&coupons).Error; err != nil {
		return nil, err
	}
	
	return coupons, nil
}

// GetUsedUpCoupons 获取已用完的优惠券
func (cr *CouponRepository) GetUsedUpCoupons() ([]*models.Coupon, error) {
	var coupons []*models.Coupon
	
	if err := cr.db.Where("total_count > 0").
		Where("used_count >= total_count").
		Where("status = ?", models.CouponStatusActive).
		Find(&coupons).Error; err != nil {
		return nil, err
	}
	
	return coupons, nil
}