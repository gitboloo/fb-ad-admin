package types

import "time"

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PagedResponse 分页响应结构
type PagedResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
	Pages   int         `json:"pages"`
}

// PageRequest 分页请求参数
type PageRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1" json:"page"`
	Size     int    `form:"size" binding:"omitempty,min=1,max=100" json:"size"`
	Sort     string `form:"sort" json:"sort"`
	Order    string `form:"order" binding:"omitempty,oneof=asc desc" json:"order"`
	Search   string `form:"search" json:"search"`
	Status   *int   `form:"status" json:"status"`
	Category string `form:"category" json:"category"`
}

// GetPage 获取页码，默认为1
func (p *PageRequest) GetPage() int {
	if p.Page <= 0 {
		return 1
	}
	return p.Page
}

// GetSize 获取每页大小，默认为20
func (p *PageRequest) GetSize() int {
	if p.Size <= 0 {
		return 20
	}
	if p.Size > 100 {
		return 100
	}
	return p.Size
}

// GetOffset 获取偏移量
func (p *PageRequest) GetOffset() int {
	return (p.GetPage() - 1) * p.GetSize()
}

// GetOrder 获取排序方式，默认为desc
func (p *PageRequest) GetOrder() string {
	if p.Order == "" {
		return "desc"
	}
	return p.Order
}

// GetSort 获取排序字段，默认为id
func (p *PageRequest) GetSort() string {
	if p.Sort == "" {
		return "id"
	}
	return p.Sort
}

// DateRange 日期范围
type DateRange struct {
	StartDate *time.Time `form:"start_date" json:"start_date"`
	EndDate   *time.Time `form:"end_date" json:"end_date"`
}

// FilterRequest 筛选请求
type FilterRequest struct {
	PageRequest
	DateRange
}

// IDRequest ID请求参数
type IDRequest struct {
	ID uint `uri:"id" binding:"required,min=1" json:"id"`
}

// IdsRequest 批量ID请求参数
type IdsRequest struct {
	Ids []uint `json:"ids" binding:"required,min=1"`
}

// StatusRequest 状态更新请求
type StatusRequest struct {
	Status int `json:"status" binding:"gte=0,lte=3"`
}

// UploadResponse 上传响应
type UploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// StatisticsResponse 统计响应
type StatisticsResponse struct {
	Total      int64                  `json:"total"`
	Active     int64                  `json:"active"`
	Inactive   int64                  `json:"inactive"`
	Growth     float64                `json:"growth"`
	TrendData  []TrendData            `json:"trend_data"`
	Categories map[string]interface{} `json:"categories"`
}

// TrendData 趋势数据
type TrendData struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Details map[string]interface{} `json:"details,omitempty"`
}
