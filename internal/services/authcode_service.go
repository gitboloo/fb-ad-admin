package services

import (
	"encoding/hex"
	"regexp"
	"time"

	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/repositories"
	"github.com/ad-platform/backend/internal/types"
)

// AuthCodeService 授权码服务
type AuthCodeService struct {
	authCodeRepo *repositories.AuthCodeRepository
}

// NewAuthCodeService 创建授权码服务
func NewAuthCodeService() *AuthCodeService {
	return &AuthCodeService{
		authCodeRepo: repositories.NewAuthCodeRepository(),
	}
}

// List 获取授权码列表
func (acs *AuthCodeService) List(req *types.FilterRequest) ([]*models.AuthCode, int64, error) {
	return acs.authCodeRepo.List(req)
}

// GetByID 根据ID获取授权码
func (acs *AuthCodeService) GetByID(id uint) (*models.AuthCode, error) {
	return acs.authCodeRepo.GetByID(id)
}

// GetByCode 根据代码获取授权码
func (acs *AuthCodeService) GetByCode(code string) (*models.AuthCode, error) {
	return acs.authCodeRepo.GetByCode(code)
}

// GenerateBatch 批量生成授权码
func (acs *AuthCodeService) GenerateBatch(count int, validDays int) ([]*models.AuthCode, error) {
	if count <= 0 || count > 1000 {
		return nil, &ServiceError{
			Code:    400,
			Message: "生成数量必须在1-1000之间",
		}
	}

	if validDays <= 0 || validDays > 365 {
		return nil, &ServiceError{
			Code:    400,
			Message: "有效天数必须在1-365之间",
		}
	}

	var authCodes []*models.AuthCode
	expireTime := time.Now().AddDate(0, 0, validDays)

	for i := 0; i < count; i++ {
		authCode := &models.AuthCode{
			Status:    models.AuthCodeStatusUnused,
			ExpiredAt: expireTime,
		}

		// 确保授权码唯一
		for {
			authCode.Code = models.GenerateAuthCode()
			existing, _ := acs.authCodeRepo.GetByCode(authCode.Code)
			if existing == nil {
				break
			}
		}

		if err := acs.authCodeRepo.Create(authCode); err != nil {
			return nil, err
		}

		authCodes = append(authCodes, authCode)
	}

	return authCodes, nil
}

// VerifyCode 验证授权码
func (acs *AuthCodeService) VerifyCode(code string, userID uint) (map[string]interface{}, error) {
	if !acs.ValidateCodeFormat(code) {
		return nil, &ServiceError{
			Code:    400,
			Message: "授权码格式无效",
		}
	}

	authCode, err := acs.authCodeRepo.GetByCode(code)
	if err != nil {
		return nil, &ServiceError{
			Code:    404,
			Message: "授权码不存在",
		}
	}

	if !authCode.IsUsable() {
		if authCode.Status == models.AuthCodeStatusUsed {
			return nil, &ServiceError{
				Code:    400,
				Message: "授权码已被使用",
			}
		} else if authCode.Status == models.AuthCodeStatusExpired || time.Now().After(authCode.ExpiredAt) {
			return nil, &ServiceError{
				Code:    400,
				Message: "授权码已过期",
			}
		} else {
			return nil, &ServiceError{
				Code:    400,
				Message: "授权码不可用",
			}
		}
	}

	// 如果提供了用户ID，则标记为已使用
	if userID > 0 {
		authCode.Use(userID)
		if err := acs.authCodeRepo.Update(authCode); err != nil {
			return nil, err
		}
	}

	result := map[string]interface{}{
		"code":          authCode.Code,
		"status":        authCode.Status,
		"expired_at":    authCode.ExpiredAt,
		"remaining_time": authCode.GetRemainingTime().String(),
		"is_used":       authCode.Status == models.AuthCodeStatusUsed,
	}

	if userID > 0 {
		result["used_by"] = userID
		result["used_at"] = authCode.UsedAt
	}

	return result, nil
}

// RevokeCode 撤销授权码
func (acs *AuthCodeService) RevokeCode(id uint) error {
	authCode, err := acs.authCodeRepo.GetByID(id)
	if err != nil {
		return err
	}

	if authCode.Status == models.AuthCodeStatusUsed {
		return &ServiceError{
			Code:    400,
			Message: "已使用的授权码无法撤销",
		}
	}

	if authCode.Status == models.AuthCodeStatusExpired {
		return &ServiceError{
			Code:    400,
			Message: "已过期的授权码无法撤销",
		}
	}

	authCode.Expire()
	return acs.authCodeRepo.Update(authCode)
}

// BatchRevoke 批量撤销授权码
func (acs *AuthCodeService) BatchRevoke(ids []uint) (map[string]interface{}, error) {
	var successCount int
	var failedCount int
	var failedIds []uint

	for _, id := range ids {
		if err := acs.RevokeCode(id); err != nil {
			failedCount++
			failedIds = append(failedIds, id)
		} else {
			successCount++
		}
	}

	return map[string]interface{}{
		"success_count": successCount,
		"failed_count":  failedCount,
		"failed_ids":    failedIds,
		"total_count":   len(ids),
	}, nil
}

// GetExpiredCodes 获取已过期的授权码
func (acs *AuthCodeService) GetExpiredCodes(req *types.FilterRequest) ([]*models.AuthCode, int64, error) {
	return acs.authCodeRepo.GetExpiredCodes(req)
}

// CleanExpiredCodes 清理过期的授权码
func (acs *AuthCodeService) CleanExpiredCodes() (int64, error) {
	expiredCodes, err := acs.authCodeRepo.GetExpiredCodesForCleanup()
	if err != nil {
		return 0, err
	}

	var cleanedCount int64
	for _, code := range expiredCodes {
		if code.Status == models.AuthCodeStatusUnused {
			code.Expire()
			if err := acs.authCodeRepo.Update(code); err == nil {
				cleanedCount++
			}
		}
	}

	return cleanedCount, nil
}

// GetStatistics 获取授权码统计
func (acs *AuthCodeService) GetStatistics() (*types.StatisticsResponse, error) {
	return acs.authCodeRepo.GetStatistics()
}

// GetUsedByUser 获取用户使用的授权码
func (acs *AuthCodeService) GetUsedByUser(userID uint, req *types.FilterRequest) ([]*models.AuthCode, int64, error) {
	return acs.authCodeRepo.GetUsedByUser(userID, req)
}

// ValidateCodeFormat 验证授权码格式
func (acs *AuthCodeService) ValidateCodeFormat(code string) bool {
	// 授权码应该是32位十六进制字符串
	if len(code) != 32 {
		return false
	}

	// 检查是否为有效的十六进制字符串
	matched, _ := regexp.MatchString("^[0-9a-fA-F]{32}$", code)
	if !matched {
		return false
	}

	// 尝试解码以验证有效性
	_, err := hex.DecodeString(code)
	return err == nil
}

// CheckAvailability 检查授权码可用性
func (acs *AuthCodeService) CheckAvailability(code string) (bool, map[string]interface{}, error) {
	if !acs.ValidateCodeFormat(code) {
		return false, map[string]interface{}{
			"reason": "授权码格式无效",
		}, nil
	}

	authCode, err := acs.authCodeRepo.GetByCode(code)
	if err != nil {
		return false, map[string]interface{}{
			"reason": "授权码不存在",
		}, nil
	}

	info := map[string]interface{}{
		"status":         authCode.Status,
		"expired_at":     authCode.ExpiredAt,
		"remaining_time": authCode.GetRemainingTime().String(),
		"created_at":     authCode.CreatedAt,
	}

	if authCode.IsUsable() {
		return true, info, nil
	}

	// 设置不可用的原因
	if authCode.Status == models.AuthCodeStatusUsed {
		info["reason"] = "授权码已被使用"
		info["used_by"] = authCode.UsedBy
		info["used_at"] = authCode.UsedAt
	} else if authCode.Status == models.AuthCodeStatusExpired || time.Now().After(authCode.ExpiredAt) {
		info["reason"] = "授权码已过期"
	} else {
		info["reason"] = "授权码不可用"
	}

	return false, info, nil
}

// GetUsageStatistics 获取使用统计
func (acs *AuthCodeService) GetUsageStatistics() (map[string]interface{}, error) {
	totalCount, _ := acs.authCodeRepo.GetTotalCount()
	usedCount, _ := acs.authCodeRepo.GetUsedCount()
	expiredCount, _ := acs.authCodeRepo.GetExpiredCount()
	availableCount := totalCount - usedCount - expiredCount

	usageRate := float64(0)
	if totalCount > 0 {
		usageRate = float64(usedCount) / float64(totalCount) * 100
	}

	return map[string]interface{}{
		"total_count":     totalCount,
		"used_count":      usedCount,
		"expired_count":   expiredCount,
		"available_count": availableCount,
		"usage_rate":      usageRate,
	}, nil
}

// GetRecentUsage 获取最近使用记录
func (acs *AuthCodeService) GetRecentUsage(limit int) ([]*models.AuthCode, error) {
	return acs.authCodeRepo.GetRecentUsed(limit)
}

// GetExpiringCodes 获取即将过期的授权码
func (acs *AuthCodeService) GetExpiringCodes(days int) ([]*models.AuthCode, error) {
	return acs.authCodeRepo.GetExpiringCodes(days)
}

// BatchGenerate 高级批量生成（支持自定义配置）
func (acs *AuthCodeService) BatchGenerate(config map[string]interface{}) ([]*models.AuthCode, error) {
	count, ok := config["count"].(int)
	if !ok || count <= 0 || count > 1000 {
		return nil, &ServiceError{
			Code:    400,
			Message: "生成数量必须在1-1000之间",
		}
	}

	validDays, ok := config["valid_days"].(int)
	if !ok || validDays <= 0 {
		validDays = 7 // 默认7天
	}

	prefix, _ := config["prefix"].(string)
	if len(prefix) > 8 {
		return nil, &ServiceError{
			Code:    400,
			Message: "前缀长度不能超过8位",
		}
	}

	return acs.GenerateBatch(count, validDays)
}