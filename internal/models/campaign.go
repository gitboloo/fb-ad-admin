package models

import (
	"database/sql/driver"
	"encoding/json"
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

type DeliveryContent struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
	Videos      []string `json:"videos"`
	CallToAction string  `json:"call_to_action"`
}

type DeliveryRules struct {
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	DailyBudget     float64   `json:"daily_budget"`
	TotalBudget     float64   `json:"total_budget"`
	BidAmount       float64   `json:"bid_amount"`
	FrequencyCap    int       `json:"frequency_cap"`
	DeliveryPacing  string    `json:"delivery_pacing"` // standard, accelerated
}

type UserTargeting struct {
	AgeRange        []int    `json:"age_range"`        // [min, max]
	Genders         []string `json:"genders"`          // male, female, all
	Countries       []string `json:"countries"`
	Languages       []string `json:"languages"`
	Interests       []string `json:"interests"`
	Behaviors       []string `json:"behaviors"`
	DeviceTypes     []string `json:"device_types"`     // mobile, tablet, desktop
	OperatingSystems []string `json:"operating_systems"` // ios, android, windows
	CustomAudiences []string `json:"custom_audiences"`
}

// Value 实现 driver.Valuer 接口
func (dc DeliveryContent) Value() (driver.Value, error) {
	return json.Marshal(dc)
}

func (dc *DeliveryContent) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}
	
	return json.Unmarshal(bytes, dc)
}

func (dr DeliveryRules) Value() (driver.Value, error) {
	return json.Marshal(dr)
}

func (dr *DeliveryRules) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}
	
	return json.Unmarshal(bytes, dr)
}

func (ut UserTargeting) Value() (driver.Value, error) {
	return json.Marshal(ut)
}

func (ut *UserTargeting) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}
	
	return json.Unmarshal(bytes, ut)
}

type Campaign struct {
	ID              uint             `json:"id" gorm:"primaryKey;autoIncrement"`
	Name            string           `json:"name" gorm:"type:varchar(255);not null"`
	ProductID       uint             `json:"product_id" gorm:"not null;index"`
	Description     string           `json:"description" gorm:"type:text"`
	Status          CampaignStatus   `json:"status" gorm:"type:tinyint;not null;default:1"`
	Logo            string           `json:"logo" gorm:"type:varchar(500)"`
	DeliveryContent *DeliveryContent `json:"delivery_content" gorm:"type:json"`
	DeliveryRules   *DeliveryRules   `json:"delivery_rules" gorm:"type:json"`
	UserTargeting   *UserTargeting   `json:"user_targeting" gorm:"type:json"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `json:"-" gorm:"index"`
	
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

// IsRunning 检查计划是否正在运行
func (c *Campaign) IsRunning() bool {
	if !c.IsActive() {
		return false
	}
	
	if c.DeliveryRules == nil {
		return false
	}
	
	now := time.Now()
	return now.After(c.DeliveryRules.StartDate) && now.Before(c.DeliveryRules.EndDate)
}

// GetRemainingBudget 获取剩余预算
func (c *Campaign) GetRemainingBudget() float64 {
	if c.DeliveryRules == nil {
		return 0
	}
	// 这里应该从实际花费记录中计算，暂时返回总预算
	return c.DeliveryRules.TotalBudget
}