package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ProductStatus int

const (
	ProductStatusInactive  ProductStatus = 0
	ProductStatusActive    ProductStatus = 1
	ProductStatusSuspended ProductStatus = 2
)

// AppInfoItem 应用信息项（key-value对）
type AppInfoItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AppInfoList 应用信息列表
type AppInfoList []AppInfoItem

// Value 实现 driver.Valuer 接口，用于数据库写入
func (a AppInfoList) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (a *AppInfoList) Scan(value interface{}) error {
	if value == nil {
		*a = AppInfoList{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*a = AppInfoList{}
		return nil
	}

	return json.Unmarshal(bytes, a)
}

type Product struct {
	ID            uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name          string         `json:"name" gorm:"type:varchar(255);not null"`
	Type          string         `json:"type" gorm:"type:varchar(100)"`
	Company       string         `json:"company" gorm:"type:varchar(255)"`
	Description   string         `json:"description" gorm:"type:text"`
	Status        ProductStatus  `json:"status" gorm:"type:tinyint;not null;default:1"`
	Logo          string         `json:"logo" gorm:"type:varchar(500)"`
	Images        string         `json:"images" gorm:"type:text"` // JSON数组存储多张图片
	GooglePayLink string         `json:"google_pay_link" gorm:"type:varchar(500)"`
	AppStoreLink  string         `json:"app_store_link" gorm:"type:varchar(500)"`
	AppInfo       AppInfoList    `json:"app_info" gorm:"type:json"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Campaigns []Campaign `json:"campaigns,omitempty" gorm:"foreignKey:ProductID"`
}

func (Product) TableName() string {
	return "products"
}

// GetImages 获取图片列表
func (p *Product) GetImages() []string {
	if p.Images == "" {
		return []string{}
	}

	var images []string
	if err := json.Unmarshal([]byte(p.Images), &images); err != nil {
		return []string{}
	}
	return images
}

// SetImages 设置图片列表
func (p *Product) SetImages(images []string) error {
	data, err := json.Marshal(images)
	if err != nil {
		return err
	}
	p.Images = string(data)
	return nil
}

// IsActive 检查产品是否激活
func (p *Product) IsActive() bool {
	return p.Status == ProductStatusActive
}
