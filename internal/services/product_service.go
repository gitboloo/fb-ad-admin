package services

import (
	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/repositories"
	"github.com/ad-platform/backend/internal/types"
)

// ProductService 产品服务
type ProductService struct {
	productRepo *repositories.ProductRepository
}

// NewProductService 创建产品服务
func NewProductService() *ProductService {
	return &ProductService{
		productRepo: repositories.NewProductRepository(),
	}
}

// List 获取产品列表
func (ps *ProductService) List(req *types.FilterRequest) ([]*models.Product, int64, error) {
	return ps.productRepo.List(req)
}

// GetByID 根据ID获取产品
func (ps *ProductService) GetByID(id uint) (*models.Product, error) {
	return ps.productRepo.GetByID(id)
}

// Create 创建产品
func (ps *ProductService) Create(product *models.Product) error {
	return ps.productRepo.Create(product)
}

// Update 更新产品
func (ps *ProductService) Update(product *models.Product) error {
	return ps.productRepo.Update(product)
}

// Delete 删除产品
func (ps *ProductService) Delete(id uint) error {
	// 检查是否有关联的活动计划
	campaigns, err := ps.productRepo.GetCampaignsByProductID(id)
	if err != nil {
		return err
	}

	if len(campaigns) > 0 {
		// 检查是否有活动的计划
		for _, campaign := range campaigns {
			if campaign.IsActive() {
				return &ServiceError{
					Code:    400,
					Message: "该产品下存在活动的计划，无法删除",
				}
			}
		}
	}

	return ps.productRepo.Delete(id)
}

// UpdateStatus 更新产品状态
func (ps *ProductService) UpdateStatus(id uint, status models.ProductStatus) error {
	product, err := ps.productRepo.GetByID(id)
	if err != nil {
		return err
	}

	product.Status = status
	return ps.productRepo.Update(product)
}

// GetStatistics 获取产品统计
func (ps *ProductService) GetStatistics() (*types.StatisticsResponse, error) {
	stats, err := ps.productRepo.GetStatistics()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetProductsByStatus 根据状态获取产品列表
func (ps *ProductService) GetProductsByStatus(status models.ProductStatus) ([]*models.Product, error) {
	return ps.productRepo.GetByStatus(status)
}

// GetProductsByType 根据类型获取产品列表
func (ps *ProductService) GetProductsByType(productType models.ProductType) ([]*models.Product, error) {
	return ps.productRepo.GetByType(productType)
}

// SearchProducts 搜索产品
func (ps *ProductService) SearchProducts(keyword string, limit int) ([]*models.Product, error) {
	return ps.productRepo.Search(keyword, limit)
}

// GetActiveProducts 获取活动产品列表
func (ps *ProductService) GetActiveProducts() ([]*models.Product, error) {
	return ps.GetProductsByStatus(models.ProductStatusActive)
}

// BatchUpdateStatus 批量更新状态
func (ps *ProductService) BatchUpdateStatus(ids []uint, status models.ProductStatus) error {
	return ps.productRepo.BatchUpdateStatus(ids, status)
}

// GetProductWithCampaigns 获取产品及其计划列表
func (ps *ProductService) GetProductWithCampaigns(id uint) (*models.Product, error) {
	return ps.productRepo.GetWithCampaigns(id)
}