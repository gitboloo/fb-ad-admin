package models

import (
	"fmt"
	"strings"
	"time"
)

type SystemConfig struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Key         string    `json:"key" gorm:"type:varchar(100);uniqueIndex;not null"`
	Value       string    `json:"value" gorm:"type:text"`
	Description string    `json:"description" gorm:"type:varchar(500)"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// 预定义的系统配置键
const (
	ConfigKeySystemName        = "system_name"
	ConfigKeySystemLogo        = "system_logo"
	ConfigKeySystemDescription = "system_description"
	ConfigKeyContactEmail      = "contact_email"
	ConfigKeyContactPhone      = "contact_phone"
	ConfigKeyMaintenanceMode   = "maintenance_mode"
	ConfigKeyRegistrationEnabled = "registration_enabled"
	ConfigKeyMaxUploadSize     = "max_upload_size"
	ConfigKeyAllowedFileTypes  = "allowed_file_types"
	ConfigKeyDefaultUserRole   = "default_user_role"
	ConfigKeyPasswordMinLength = "password_min_length"
	ConfigKeySessionTimeout    = "session_timeout"
	ConfigKeyEmailSMTPHost     = "email_smtp_host"
	ConfigKeyEmailSMTPPort     = "email_smtp_port"
	ConfigKeyEmailSMTPUser     = "email_smtp_user"
	ConfigKeyEmailSMTPPassword = "email_smtp_password"
	ConfigKeyPaymentGateway    = "payment_gateway"
	ConfigKeyPaymentPublicKey  = "payment_public_key"
	ConfigKeyPaymentPrivateKey = "payment_private_key"
	ConfigKeySMSProvider       = "sms_provider"
	ConfigKeySMSAPIKey         = "sms_api_key"
	ConfigKeyStorageProvider   = "storage_provider"
	ConfigKeyStorageEndpoint   = "storage_endpoint"
	ConfigKeyStorageAccessKey  = "storage_access_key"
	ConfigKeyStorageSecretKey  = "storage_secret_key"
)

// GetDefaultConfigs 获取默认配置
func GetDefaultConfigs() map[string]SystemConfig {
	return map[string]SystemConfig{
		ConfigKeySystemName: {
			Key:         ConfigKeySystemName,
			Value:       "广告平台",
			Description: "系统名称",
		},
		ConfigKeySystemLogo: {
			Key:         ConfigKeySystemLogo,
			Value:       "/static/logo.png",
			Description: "系统Logo",
		},
		ConfigKeySystemDescription: {
			Key:         ConfigKeySystemDescription,
			Value:       "专业的广告投放管理平台",
			Description: "系统描述",
		},
		ConfigKeyContactEmail: {
			Key:         ConfigKeyContactEmail,
			Value:       "admin@example.com",
			Description: "联系邮箱",
		},
		ConfigKeyContactPhone: {
			Key:         ConfigKeyContactPhone,
			Value:       "400-000-0000",
			Description: "联系电话",
		},
		ConfigKeyMaintenanceMode: {
			Key:         ConfigKeyMaintenanceMode,
			Value:       "false",
			Description: "维护模式开关",
		},
		ConfigKeyRegistrationEnabled: {
			Key:         ConfigKeyRegistrationEnabled,
			Value:       "true",
			Description: "允许用户注册",
		},
		ConfigKeyMaxUploadSize: {
			Key:         ConfigKeyMaxUploadSize,
			Value:       "10485760", // 10MB
			Description: "最大上传文件大小(字节)",
		},
		ConfigKeyAllowedFileTypes: {
			Key:         ConfigKeyAllowedFileTypes,
			Value:       "jpg,jpeg,png,gif,pdf,doc,docx,xls,xlsx,ppt,pptx",
			Description: "允许上传的文件类型",
		},
		ConfigKeyDefaultUserRole: {
			Key:         ConfigKeyDefaultUserRole,
			Value:       "1",
			Description: "默认用户角色",
		},
		ConfigKeyPasswordMinLength: {
			Key:         ConfigKeyPasswordMinLength,
			Value:       "6",
			Description: "密码最小长度",
		},
		ConfigKeySessionTimeout: {
			Key:         ConfigKeySessionTimeout,
			Value:       "3600", // 1小时
			Description: "会话超时时间(秒)",
		},
	}
}

// IsEnabled 检查布尔值配置是否启用
func (sc *SystemConfig) IsEnabled() bool {
	return sc.Value == "true" || sc.Value == "1"
}

// GetIntValue 获取整数值
func (sc *SystemConfig) GetIntValue() int {
	var value int
	if _, err := fmt.Sscanf(sc.Value, "%d", &value); err != nil {
		return 0
	}
	return value
}

// GetFloatValue 获取浮点数值
func (sc *SystemConfig) GetFloatValue() float64 {
	var value float64
	if _, err := fmt.Sscanf(sc.Value, "%f", &value); err != nil {
		return 0
	}
	return value
}

// GetSliceValue 获取切片值（逗号分隔）
func (sc *SystemConfig) GetSliceValue() []string {
	if sc.Value == "" {
		return []string{}
	}
	return strings.Split(sc.Value, ",")
}