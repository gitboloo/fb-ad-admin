package repositories

import (
	"fmt"
	"strings"
	"time"

	"github.com/ad-platform/backend/internal/database"
	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/types"
	"gorm.io/gorm"
)

// AuthCodeRepository 授权码仓库
type AuthCodeRepository struct {
	db *gorm.DB
}

// NewAuthCodeRepository 创建授权码仓库
func NewAuthCodeRepository() *AuthCodeRepository {
	return &AuthCodeRepository{
		db: database.DB,
	}
}

// List 获取授权码列表
func (acr *AuthCodeRepository) List(req *types.FilterRequest) ([]*models.AuthCode, int64, error) {
	var authCodes []*models.AuthCode
	var total int64

	query := acr.db.Model(&models.AuthCode{})

	// 搜索条件
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("code LIKE ?", searchPattern)
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
		Find(&authCodes).Error; err != nil {
		return nil, 0, err
	}

	return authCodes, total, nil
}

// GetByID 根据ID获取授权码
func (acr *AuthCodeRepository) GetByID(id uint) (*models.AuthCode, error) {
	var authCode models.AuthCode
	if err := acr.db.First(&authCode, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("授权码不存在")
		}
		return nil, err
	}
	return &authCode, nil
}

// GetByCode 根据代码获取授权码
func (acr *AuthCodeRepository) GetByCode(code string) (*models.AuthCode, error) {
	var authCode models.AuthCode
	if err := acr.db.Where("code = ?", code).First(&authCode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("授权码不存在")
		}
		return nil, err
	}
	return &authCode, nil
}

// Create 创建授权码
func (acr *AuthCodeRepository) Create(authCode *models.AuthCode) error {
	return acr.db.Create(authCode).Error
}

// Update 更新授权码
func (acr *AuthCodeRepository) Update(authCode *models.AuthCode) error {
	return acr.db.Save(authCode).Error
}

// Delete 删除授权码
func (acr *AuthCodeRepository) Delete(id uint) error {
	result := acr.db.Delete(&models.AuthCode{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("授权码不存在")
	}
	return nil
}

// GetByStatus 根据状态获取授权码
func (acr *AuthCodeRepository) GetByStatus(status models.AuthCodeStatus) ([]*models.AuthCode, error) {
	var authCodes []*models.AuthCode
	if err := acr.db.Where("status = ?", status).Find(&authCodes).Error; err != nil {
		return nil, err
	}
	return authCodes, nil
}

// GetExpiredCodes 获取已过期的授权码
func (acr *AuthCodeRepository) GetExpiredCodes(req *types.FilterRequest) ([]*models.AuthCode, int64, error) {
	var authCodes []*models.AuthCode
	var total int64

	query := acr.db.Model(&models.AuthCode{}).Where("expired_at <= NOW()")

	// 搜索条件
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("code LIKE ?", searchPattern)
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
		Find(&authCodes).Error; err != nil {
		return nil, 0, err
	}

	return authCodes, total, nil
}

// GetExpiredCodesForCleanup 获取需要清理的过期授权码
func (acr *AuthCodeRepository) GetExpiredCodesForCleanup() ([]*models.AuthCode, error) {
	var authCodes []*models.AuthCode
	if err := acr.db.Where("expired_at <= NOW()").
		Where("status = ?", models.AuthCodeStatusUnused).
		Find(&authCodes).Error; err != nil {
		return nil, err
	}
	return authCodes, nil
}

// GetUsedByUser 获取用户使用的授权码
func (acr *AuthCodeRepository) GetUsedByUser(userID uint, req *types.FilterRequest) ([]*models.AuthCode, int64, error) {
	var authCodes []*models.AuthCode
	var total int64

	query := acr.db.Model(&models.AuthCode{}).Where("used_by = ?", userID)

	// 搜索条件
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("code LIKE ?", searchPattern)
	}

	// 日期范围筛选
	if req.StartDate != nil {
		query = query.Where("used_at >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("used_at <= ?", req.EndDate)
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
		Find(&authCodes).Error; err != nil {
		return nil, 0, err
	}

	return authCodes, total, nil
}

// Search 搜索授权码
func (acr *AuthCodeRepository) Search(keyword string, limit int) ([]*models.AuthCode, error) {
	var authCodes []*models.AuthCode
	searchPattern := "%" + strings.TrimSpace(keyword) + "%"
	
	query := acr.db.Where("code LIKE ?", searchPattern)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&authCodes).Error; err != nil {
		return nil, err
	}
	return authCodes, nil
}

// GetStatistics 获取授权码统计
func (acr *AuthCodeRepository) GetStatistics() (*types.StatisticsResponse, error) {
	var total int64
	var used int64
	var expired int64

	// 总数统计
	if err := acr.db.Model(&models.AuthCode{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// 已使用数量
	if err := acr.db.Model(&models.AuthCode{}).
		Where("status = ?", models.AuthCodeStatusUsed).Count(&used).Error; err != nil {
		return nil, err
	}

	// 已过期数量
	if err := acr.db.Model(&models.AuthCode{}).
		Where("status = ? OR expired_at <= NOW()", models.AuthCodeStatusExpired).Count(&expired).Error; err != nil {
		return nil, err
	}

	available := total - used - expired

	// 按状态统计
	var statusStats []struct {
		Status models.AuthCodeStatus `json:"status"`
		Count  int64                 `json:"count"`
	}
	if err := acr.db.Model(&models.AuthCode{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusStats).Error; err != nil {
		return nil, err
	}

	categories := make(map[string]interface{})
	for _, stat := range statusStats {
		var statusName string
		switch stat.Status {
		case models.AuthCodeStatusUnused:
			statusName = "未使用"
		case models.AuthCodeStatusUsed:
			statusName = "已使用"
		case models.AuthCodeStatusExpired:
			statusName = "已过期"
		default:
			statusName = "其他"
		}
		categories[statusName] = stat.Count
	}

	// 趋势数据（最近7天使用量）
	var trendData []types.TrendData
	if err := acr.db.Raw(`
		SELECT DATE(used_at) as date, COUNT(*) as value 
		FROM auth_codes 
		WHERE used_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) 
		AND status = ?
		GROUP BY DATE(used_at) 
		ORDER BY date ASC
	`, models.AuthCodeStatusUsed).Scan(&trendData).Error; err != nil {
		return nil, err
	}

	// 计算使用率增长（与上周同期比较）
	var currentWeekUsed int64
	var lastWeekUsed int64
	
	acr.db.Model(&models.AuthCode{}).
		Where("used_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").
		Where("status = ?", models.AuthCodeStatusUsed).Count(&currentWeekUsed)
	acr.db.Model(&models.AuthCode{}).
		Where("used_at >= DATE_SUB(NOW(), INTERVAL 14 DAY) AND used_at < DATE_SUB(NOW(), INTERVAL 7 DAY)").
		Where("status = ?", models.AuthCodeStatusUsed).Count(&lastWeekUsed)

	var growth float64
	if lastWeekUsed > 0 {
		growth = float64(currentWeekUsed-lastWeekUsed) / float64(lastWeekUsed) * 100
	}

	return &types.StatisticsResponse{
		Total:      total,
		Active:     available,
		Inactive:   used + expired,
		Growth:     growth,
		TrendData:  trendData,
		Categories: categories,
	}, nil
}

// GetTotalCount 获取总数量
func (acr *AuthCodeRepository) GetTotalCount() (int64, error) {
	var count int64
	if err := acr.db.Model(&models.AuthCode{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetUsedCount 获取已使用数量
func (acr *AuthCodeRepository) GetUsedCount() (int64, error) {
	var count int64
	if err := acr.db.Model(&models.AuthCode{}).
		Where("status = ?", models.AuthCodeStatusUsed).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetExpiredCount 获取已过期数量
func (acr *AuthCodeRepository) GetExpiredCount() (int64, error) {
	var count int64
	if err := acr.db.Model(&models.AuthCode{}).
		Where("status = ? OR expired_at <= NOW()", models.AuthCodeStatusExpired).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetRecentUsed 获取最近使用的授权码
func (acr *AuthCodeRepository) GetRecentUsed(limit int) ([]*models.AuthCode, error) {
	var authCodes []*models.AuthCode
	
	query := acr.db.Where("status = ?", models.AuthCodeStatusUsed).
		Order("used_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&authCodes).Error; err != nil {
		return nil, err
	}
	return authCodes, nil
}

// GetExpiringCodes 获取即将过期的授权码
func (acr *AuthCodeRepository) GetExpiringCodes(days int) ([]*models.AuthCode, error) {
	var authCodes []*models.AuthCode
	
	future := time.Now().AddDate(0, 0, days)
	if err := acr.db.Where("status = ?", models.AuthCodeStatusUnused).
		Where("expired_at <= ? AND expired_at > NOW()", future).
		Order("expired_at ASC").
		Find(&authCodes).Error; err != nil {
		return nil, err
	}
	
	return authCodes, nil
}

// BatchUpdateStatus 批量更新状态
func (acr *AuthCodeRepository) BatchUpdateStatus(ids []uint, status models.AuthCodeStatus) error {
	return acr.db.Model(&models.AuthCode{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// GetUsageRateByDate 获取按日期的使用率统计
func (acr *AuthCodeRepository) GetUsageRateByDate(days int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	if err := acr.db.Raw(`
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as total,
			COUNT(CASE WHEN status = ? THEN 1 END) as used,
			(COUNT(CASE WHEN status = ? THEN 1 END) / COUNT(*)) * 100 as usage_rate
		FROM auth_codes 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, models.AuthCodeStatusUsed, models.AuthCodeStatusUsed, days).Scan(&results).Error; err != nil {
		return nil, err
	}
	
	return results, nil
}