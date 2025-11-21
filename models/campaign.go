package models

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

type CampaignStatus int

const (
	CampaignStatusInactive CampaignStatus = 0
	CampaignStatusActive   CampaignStatus = 1
	CampaignStatusPaused   CampaignStatus = 2
	CampaignStatusEnded    CampaignStatus = 3
)

// CustomField 自定义字段项(标题+内容对)
type CustomField struct {
	Title   string `json:"title"`   // 字段标题/表头
	Content string `json:"content"` // 字段内容
}

// CustomFieldList 自定义字段列表类型
type CustomFieldList []CustomField

// Value 实现 driver.Valuer 接口
func (cfl CustomFieldList) Value() (driver.Value, error) {
	return json.Marshal(cfl)
}

// Scan 实现 sql.Scanner 接口
func (cfl *CustomFieldList) Scan(value any) error {
	if value == nil {
		*cfl = CustomFieldList{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*cfl = CustomFieldList{}
		return nil
	}

	// 如果是空字符串或null，返回空数组
	if len(bytes) == 0 || string(bytes) == "null" {
		*cfl = CustomFieldList{}
		return nil
	}

	// 尝试解析为数组
	var list []CustomField
	if err := json.Unmarshal(bytes, &list); err == nil {
		*cfl = list
		return nil
	}

	// 如果解析数组失败，尝试解析为单个对象
	var single CustomField
	if err := json.Unmarshal(bytes, &single); err == nil {
		*cfl = []CustomField{single}
		return nil
	}

	// 如果是空对象 {}，返回空数组
	trimmed := strings.TrimSpace(string(bytes))
	if trimmed == "{}" || trimmed == "[]" {
		*cfl = CustomFieldList{}
		return nil
	}

	// 都失败了，返回空数组而不是错误
	*cfl = CustomFieldList{}
	return nil
}

type Campaign struct {
	ID              uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	Name            string          `json:"name" gorm:"type:varchar(255);not null;comment:计划名称"`
	CampaignNumber  string          `json:"campaign_number" gorm:"type:varchar(50);uniqueIndex;comment:计划编号"`
	ProductID       uint            `json:"product_id" gorm:"not null;index;comment:关联产品ID"`
	Description     string          `json:"description" gorm:"type:text;comment:计划简介"`
	Status          CampaignStatus  `json:"status" gorm:"type:tinyint;not null;default:1;comment:状态"`
	MainImage       string          `json:"main_image" gorm:"type:varchar(500);comment:主图URL"`
	Video           string          `json:"video" gorm:"type:varchar(500);comment:视频URL"`
	DeliveryContent CustomFieldList `json:"delivery_content" gorm:"type:json;comment:投放内容(自定义字段数组)"`
	DeliveryRules   CustomFieldList `json:"delivery_rules" gorm:"type:json;comment:投放规则(自定义字段数组)"`
	UserTargeting   CustomFieldList `json:"user_targeting" gorm:"type:json;comment:用户定向(自定义字段数组)"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `json:"-" gorm:"index"`

	// 关联
	Product *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

func (Campaign) TableName() string {
	return "campaigns"
}

// IsActive 检查计划是否激活
func (c *Campaign) IsActive() bool {
	return c.Status == CampaignStatusActive
}

// GetFieldValue 从自定义字段列表中获取指定标题的值
func (c *Campaign) GetFieldValue(fields CustomFieldList, title string) string {
	for _, field := range fields {
		if field.Title == title {
			return field.Content
		}
	}
	return ""
}
