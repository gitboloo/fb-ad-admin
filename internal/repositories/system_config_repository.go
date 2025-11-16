package repositories

import (
	"fmt"

	"github.com/ad-platform/backend/internal/database"
	"github.com/ad-platform/backend/internal/models"
	"gorm.io/gorm"
)

// SystemConfigRepository 系统配置仓库
type SystemConfigRepository struct {
	db *gorm.DB
}

// NewSystemConfigRepository 创建系统配置仓库
func NewSystemConfigRepository() *SystemConfigRepository {
	return &SystemConfigRepository{
		db: database.DB,
	}
}

// GetAllConfigs 获取所有系统配置
func (scr *SystemConfigRepository) GetAllConfigs() (map[string]*models.SystemConfig, error) {
	var configs []*models.SystemConfig
	if err := scr.db.Find(&configs).Error; err != nil {
		return nil, err
	}

	configMap := make(map[string]*models.SystemConfig)
	for _, config := range configs {
		configMap[config.Key] = config
	}

	return configMap, nil
}

// GetByKey 根据键获取配置
func (scr *SystemConfigRepository) GetByKey(key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	if err := scr.db.Where("key = ?", key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("配置不存在")
		}
		return nil, err
	}
	return &config, nil
}

// Create 创建配置
func (scr *SystemConfigRepository) Create(config *models.SystemConfig) error {
	return scr.db.Create(config).Error
}

// Update 更新配置
func (scr *SystemConfigRepository) Update(config *models.SystemConfig) error {
	return scr.db.Save(config).Error
}

// CreateOrUpdate 创建或更新配置
func (scr *SystemConfigRepository) CreateOrUpdate(config *models.SystemConfig) error {
	var existing models.SystemConfig
	if err := scr.db.Where("key = ?", config.Key).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 不存在则创建
			return scr.db.Create(config).Error
		}
		return err
	}

	// 存在则更新
	existing.Value = config.Value
	existing.Description = config.Description
	return scr.db.Save(&existing).Error
}

// Delete 删除配置
func (scr *SystemConfigRepository) Delete(key string) error {
	result := scr.db.Where("key = ?", key).Delete(&models.SystemConfig{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("配置不存在")
	}
	return nil
}

// UpdateConfigs 批量更新配置
func (scr *SystemConfigRepository) UpdateConfigs(configs map[string]string) error {
	tx := scr.db.Begin()

	for key, value := range configs {
		var config models.SystemConfig
		if err := tx.Where("key = ?", key).First(&config).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 不存在则创建
				config = models.SystemConfig{
					Key:   key,
					Value: value,
				}
				if err := tx.Create(&config).Error; err != nil {
					tx.Rollback()
					return err
				}
			} else {
				tx.Rollback()
				return err
			}
		} else {
			// 存在则更新
			config.Value = value
			if err := tx.Save(&config).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// GetByKeys 根据键列表获取配置
func (scr *SystemConfigRepository) GetByKeys(keys []string) (map[string]*models.SystemConfig, error) {
	var configs []*models.SystemConfig
	if err := scr.db.Where("key IN ?", keys).Find(&configs).Error; err != nil {
		return nil, err
	}

	configMap := make(map[string]*models.SystemConfig)
	for _, config := range configs {
		configMap[config.Key] = config
	}

	return configMap, nil
}

// GetConfigsByPrefix 根据前缀获取配置
func (scr *SystemConfigRepository) GetConfigsByPrefix(prefix string) ([]*models.SystemConfig, error) {
	var configs []*models.SystemConfig
	if err := scr.db.Where("key LIKE ?", prefix+"%").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// GetCount 获取配置总数
func (scr *SystemConfigRepository) GetCount() (int64, error) {
	var count int64
	if err := scr.db.Model(&models.SystemConfig{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// HealthCheck 健康检查
func (scr *SystemConfigRepository) HealthCheck() error {
	return scr.db.Exec("SELECT 1").Error
}

// GetSystemStats 获取系统配置统计
func (scr *SystemConfigRepository) GetSystemStats() (map[string]interface{}, error) {
	var count int64
	if err := scr.db.Model(&models.SystemConfig{}).Count(&count).Error; err != nil {
		return nil, err
	}

	// 按配置类型分组统计
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	// 简单的类型分类逻辑
	configTypes := map[string]string{
		"system_":   "系统配置",
		"email_":    "邮件配置",
		"payment_":  "支付配置",
		"sms_":      "短信配置",
		"storage_":  "存储配置",
		"security_": "安全配置",
	}

	for prefix, typeName := range configTypes {
		var typeCount int64
		scr.db.Model(&models.SystemConfig{}).
			Where("key LIKE ?", prefix+"%").Count(&typeCount)
		typeStats = append(typeStats, struct {
			Type  string `json:"type"`
			Count int64  `json:"count"`
		}{
			Type:  typeName,
			Count: typeCount,
		})
	}

	return map[string]interface{}{
		"total_configs": count,
		"type_stats":    typeStats,
	}, nil
}

// BatchDelete 批量删除配置
func (scr *SystemConfigRepository) BatchDelete(keys []string) error {
	return scr.db.Where("key IN ?", keys).Delete(&models.SystemConfig{}).Error
}

// ResetToDefaults 重置为默认配置
func (scr *SystemConfigRepository) ResetToDefaults() error {
	// 清空现有配置
	if err := scr.db.Exec("TRUNCATE TABLE system_configs").Error; err != nil {
		return err
	}

	// 插入默认配置
	defaultConfigs := models.GetDefaultConfigs()
	for _, config := range defaultConfigs {
		if err := scr.Create(&config); err != nil {
			return err
		}
	}

	return nil
}

// ExportConfigs 导出配置
func (scr *SystemConfigRepository) ExportConfigs() (map[string]string, error) {
	configs, err := scr.GetAllConfigs()
	if err != nil {
		return nil, err
	}

	exportMap := make(map[string]string)
	for key, config := range configs {
		exportMap[key] = config.Value
	}

	return exportMap, nil
}

// ImportConfigs 导入配置
func (scr *SystemConfigRepository) ImportConfigs(configs map[string]string, override bool) error {
	tx := scr.db.Begin()

	for key, value := range configs {
		var existing models.SystemConfig
		err := tx.Where("key = ?", key).First(&existing).Error
		
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 不存在则创建
				config := models.SystemConfig{
					Key:   key,
					Value: value,
				}
				if err := tx.Create(&config).Error; err != nil {
					tx.Rollback()
					return err
				}
			} else {
				tx.Rollback()
				return err
			}
		} else {
			// 存在时根据override决定是否覆盖
			if override {
				existing.Value = value
				if err := tx.Save(&existing).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	return tx.Commit().Error
}

// ValidateConfig 验证配置值
func (scr *SystemConfigRepository) ValidateConfig(key, value string) error {
	// 根据配置键验证值的有效性
	switch key {
	case models.ConfigKeyMaintenanceMode, models.ConfigKeyRegistrationEnabled:
		if value != "true" && value != "false" {
			return fmt.Errorf("布尔值配置只能是 true 或 false")
		}
	case models.ConfigKeyMaxUploadSize, models.ConfigKeyPasswordMinLength, models.ConfigKeySessionTimeout:
		if value == "" {
			return fmt.Errorf("数值配置不能为空")
		}
	case models.ConfigKeyContactEmail:
		// 简单的邮箱格式验证
		if value != "" && len(value) > 0 {
			// 这里可以添加更严格的邮箱验证
		}
	}
	return nil
}