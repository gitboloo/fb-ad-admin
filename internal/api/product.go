package api

import (
	"strconv"

	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/services"
	"github.com/ad-platform/backend/internal/types"
	"github.com/ad-platform/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// ProductRequest 产品请求结构
type ProductRequest struct {
	Name           string               `json:"name" binding:"required,max=255"`
	Type           models.ProductType   `json:"type" binding:"required,min=1,max=4"`
	Company        string               `json:"company" binding:"max=255"`
	Description    string               `json:"description"`
	Status         models.ProductStatus `json:"status" binding:"min=0,max=2"`
	GooglePayLink  string               `json:"google_pay_link" binding:"omitempty,url,max=500"`
	AppStoreLink   string               `json:"app_store_link" binding:"omitempty,url,max=500"`
	AppInfo        *models.AppInfo      `json:"app_info"`
}

// ProductController 产品控制器
type ProductController struct {
	productService *services.ProductService
}

// NewProductController 创建产品控制器
func NewProductController() *ProductController {
	return &ProductController{
		productService: services.NewProductService(),
	}
}

// List 获取产品列表
func (pc *ProductController) List(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	products, total, err := pc.productService.List(&req)
	if err != nil {
		utils.InternalServerError(c, "获取产品列表失败")
		return
	}

	utils.PagedSuccess(c, products, total, req.GetPage(), req.GetSize())
}

// GetByID 获取产品详情
func (pc *ProductController) GetByID(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	product, err := pc.productService.GetByID(req.ID)
	if err != nil {
		if err.Error() == "产品不存在" {
			utils.NotFound(c, "产品不存在")
		} else {
			utils.InternalServerError(c, "获取产品详情失败")
		}
		return
	}

	utils.Success(c, product)
}

// Create 创建产品
func (pc *ProductController) Create(c *gin.Context) {
	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	product := &models.Product{
		Name:          req.Name,
		Type:          req.Type,
		Company:       req.Company,
		Description:   req.Description,
		Status:        req.Status,
		GooglePayLink: req.GooglePayLink,
		AppStoreLink:  req.AppStoreLink,
		AppInfo:       req.AppInfo,
	}

	if err := pc.productService.Create(product); err != nil {
		utils.InternalServerError(c, "创建产品失败")
		return
	}

	utils.Created(c, product)
}

// Update 更新产品
func (pc *ProductController) Update(c *gin.Context) {
	var uriReq types.IDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.ValidateError(c, err)
		return
	}

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 检查产品是否存在
	product, err := pc.productService.GetByID(uriReq.ID)
	if err != nil {
		if err.Error() == "产品不存在" {
			utils.NotFound(c, "产品不存在")
		} else {
			utils.InternalServerError(c, "获取产品信息失败")
		}
		return
	}

	// 更新字段
	product.Name = req.Name
	product.Type = req.Type
	product.Company = req.Company
	product.Description = req.Description
	product.Status = req.Status
	product.GooglePayLink = req.GooglePayLink
	product.AppStoreLink = req.AppStoreLink
	product.AppInfo = req.AppInfo

	if err := pc.productService.Update(product); err != nil {
		utils.InternalServerError(c, "更新产品失败")
		return
	}

	utils.Updated(c, product)
}

// Delete 删除产品
func (pc *ProductController) Delete(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := pc.productService.Delete(req.ID); err != nil {
		if err.Error() == "产品不存在" {
			utils.NotFound(c, "产品不存在")
		} else {
			utils.InternalServerError(c, "删除产品失败")
		}
		return
	}

	utils.Deleted(c)
}

// UploadLogo 上传产品Logo
func (pc *ProductController) UploadLogo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的产品ID")
		return
	}

	// 检查产品是否存在
	product, err := pc.productService.GetByID(uint(id))
	if err != nil {
		if err.Error() == "产品不存在" {
			utils.NotFound(c, "产品不存在")
		} else {
			utils.InternalServerError(c, "获取产品信息失败")
		}
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("logo")
	if err != nil {
		utils.BadRequest(c, "请选择要上传的Logo文件")
		return
	}

	// 保存文件
	uploadResp, err := utils.SaveUploadedFile(c, file, "products/logos")
	if err != nil {
		utils.BadRequest(c, "上传失败: "+err.Error())
		return
	}

	// 删除旧Logo文件
	if product.Logo != "" {
		utils.DeleteFile(product.Logo)
	}

	// 更新产品Logo字段
	product.Logo = uploadResp.URL
	if err := pc.productService.Update(product); err != nil {
		utils.InternalServerError(c, "更新产品Logo失败")
		return
	}

	utils.Success(c, uploadResp)
}

// UploadImages 上传产品图片
func (pc *ProductController) UploadImages(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的产品ID")
		return
	}

	// 检查产品是否存在
	product, err := pc.productService.GetByID(uint(id))
	if err != nil {
		if err.Error() == "产品不存在" {
			utils.NotFound(c, "产品不存在")
		} else {
			utils.InternalServerError(c, "获取产品信息失败")
		}
		return
	}

	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		utils.BadRequest(c, "获取上传文件失败")
		return
	}

	files, exists := form.File["images"]
	if !exists || len(files) == 0 {
		utils.BadRequest(c, "请选择要上传的图片文件")
		return
	}

	// 保存多个文件
	uploadResponses, err := utils.SaveMultipleFiles(c, files, "products/images")
	if err != nil {
		utils.BadRequest(c, "上传失败: "+err.Error())
		return
	}

	// 获取现有图片列表
	existingImages := product.GetImages()

	// 添加新图片URL
	for _, resp := range uploadResponses {
		existingImages = append(existingImages, resp.URL)
	}

	// 更新产品图片字段
	if err := product.SetImages(existingImages); err != nil {
		utils.InternalServerError(c, "更新产品图片失败")
		return
	}

	if err := pc.productService.Update(product); err != nil {
		utils.InternalServerError(c, "更新产品图片失败")
		return
	}

	utils.Success(c, uploadResponses)
}

// GetStatistics 获取产品统计
func (pc *ProductController) GetStatistics(c *gin.Context) {
	stats, err := pc.productService.GetStatistics()
	if err != nil {
		utils.InternalServerError(c, "获取产品统计失败")
		return
	}

	utils.Success(c, stats)
}