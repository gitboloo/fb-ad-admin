package api

import (
	"backend/models"
	"backend/services"
	"backend/types"
	"backend/utils"
	"github.com/gin-gonic/gin"
)

// GenerateAuthCodeRequest 生成授权码请求结构
type GenerateAuthCodeRequest struct {
	Count     int `json:"count" binding:"required,min=1,max=1000"`
	ValidDays int `json:"valid_days" binding:"min=1,max=365"`
}

// VerifyAuthCodeRequest 验证授权码请求结构
type VerifyAuthCodeRequest struct {
	Code string `json:"code" binding:"required,min=1"`
}

// AuthCodeController 授权码控制器
type AuthCodeController struct {
	authCodeService *services.AuthCodeService
}

// NewAuthCodeController 创建授权码控制器
func NewAuthCodeController() *AuthCodeController {
	return &AuthCodeController{
		authCodeService: services.NewAuthCodeService(),
	}
}

// List 获取授权码列表
func (acc *AuthCodeController) List(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	authCodes, total, err := acc.authCodeService.List(&req)
	if err != nil {
		utils.InternalServerError(c, "获取授权码列表失败")
		return
	}

	utils.PagedSuccess(c, authCodes, total, req.GetPage(), req.GetSize())
}

// GetByID 获取授权码详情
func (acc *AuthCodeController) GetByID(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	authCode, err := acc.authCodeService.GetByID(req.ID)
	if err != nil {
		if err.Error() == "授权码不存在" {
			utils.NotFound(c, "授权码不存在")
		} else {
			utils.InternalServerError(c, "获取授权码详情失败")
		}
		return
	}

	utils.Success(c, authCode)
}

// Generate 批量生成授权码
func (acc *AuthCodeController) Generate(c *gin.Context) {
	var req GenerateAuthCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 设置默认有效期
	if req.ValidDays <= 0 {
		req.ValidDays = 7 // 默认7天
	}

	authCodes, err := acc.authCodeService.GenerateBatch(req.Count, req.ValidDays)
	if err != nil {
		utils.InternalServerError(c, "生成授权码失败")
		return
	}

	utils.SuccessWithMessage(c, "授权码生成成功", map[string]interface{}{
		"count":      len(authCodes),
		"valid_days": req.ValidDays,
		"auth_codes": authCodes,
	})
}

// Verify 验证授权码
func (acc *AuthCodeController) Verify(c *gin.Context) {
	var req VerifyAuthCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 获取当前用户ID（如果已登录）
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
	}

	result, err := acc.authCodeService.VerifyCode(req.Code, userID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "授权码验证成功", result)
}

// Revoke 撤销授权码
func (acc *AuthCodeController) Revoke(c *gin.Context) {
	var req types.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := acc.authCodeService.RevokeCode(req.ID); err != nil {
		if err.Error() == "授权码不存在" {
			utils.NotFound(c, "授权码不存在")
		} else {
			utils.BadRequest(c, err.Error())
		}
		return
	}

	utils.SuccessWithMessage(c, "授权码已撤销", nil)
}

// BatchRevoke 批量撤销授权码
func (acc *AuthCodeController) BatchRevoke(c *gin.Context) {
	var req types.IdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	result, err := acc.authCodeService.BatchRevoke(req.Ids)
	if err != nil {
		utils.InternalServerError(c, "批量撤销失败")
		return
	}

	utils.SuccessWithMessage(c, "批量撤销完成", result)
}

// Export 导出授权码
func (acc *AuthCodeController) Export(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 只导出未使用的授权码
	status := int(models.AuthCodeStatusUnused)
	req.Status = &status

	// 设置导出数量限制
	req.Size = 10000

	authCodes, _, err := acc.authCodeService.List(&req)
	if err != nil {
		utils.InternalServerError(c, "导出授权码失败")
		return
	}

	// 简化导出数据
	exportData := make([]map[string]interface{}, 0)
	for _, code := range authCodes {
		exportData = append(exportData, map[string]interface{}{
			"code":       code.Code,
			"status":     code.Status,
			"expired_at": code.ExpiredAt,
			"created_at": code.CreatedAt,
		})
	}

	utils.Success(c, map[string]interface{}{
		"auth_codes":  exportData,
		"count":       len(exportData),
		"exported_at": "now",
	})
}

// GetStatistics 获取授权码统计
func (acc *AuthCodeController) GetStatistics(c *gin.Context) {
	stats, err := acc.authCodeService.GetStatistics()
	if err != nil {
		utils.InternalServerError(c, "获取授权码统计失败")
		return
	}

	utils.Success(c, stats)
}

// GetUsageHistory 获取授权码使用历史
func (acc *AuthCodeController) GetUsageHistory(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 只获取已使用的授权码
	status := int(models.AuthCodeStatusUsed)
	req.Status = &status

	usedCodes, total, err := acc.authCodeService.List(&req)
	if err != nil {
		utils.InternalServerError(c, "获取使用历史失败")
		return
	}

	utils.PagedSuccess(c, usedCodes, total, req.GetPage(), req.GetSize())
}

// GetExpiredCodes 获取已过期的授权码
func (acc *AuthCodeController) GetExpiredCodes(c *gin.Context) {
	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	expiredCodes, total, err := acc.authCodeService.GetExpiredCodes(&req)
	if err != nil {
		utils.InternalServerError(c, "获取过期授权码失败")
		return
	}

	utils.PagedSuccess(c, expiredCodes, total, req.GetPage(), req.GetSize())
}

// CleanExpired 清理过期的授权码
func (acc *AuthCodeController) CleanExpired(c *gin.Context) {
	count, err := acc.authCodeService.CleanExpiredCodes()
	if err != nil {
		utils.InternalServerError(c, "清理过期授权码失败")
		return
	}

	utils.SuccessWithMessage(c, "清理完成", map[string]interface{}{
		"cleaned_count": count,
	})
}

// GetCodeByCode 根据代码查找授权码信息
func (acc *AuthCodeController) GetCodeByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		utils.BadRequest(c, "授权码不能为空")
		return
	}

	authCode, err := acc.authCodeService.GetByCode(code)
	if err != nil {
		if err.Error() == "授权码不存在" {
			utils.NotFound(c, "授权码不存在")
		} else {
			utils.InternalServerError(c, "查询授权码失败")
		}
		return
	}

	utils.Success(c, authCode)
}

// ValidateCodeFormat 验证授权码格式
func (acc *AuthCodeController) ValidateCodeFormat(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	isValid := acc.authCodeService.ValidateCodeFormat(req.Code)
	
	utils.Success(c, map[string]interface{}{
		"code":     req.Code,
		"is_valid": isValid,
	})
}

// GetMyUsedCodes 获取当前用户使用过的授权码
func (acc *AuthCodeController) GetMyUsedCodes(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	var req types.FilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	usedCodes, total, err := acc.authCodeService.GetUsedByUser(userID.(uint), &req)
	if err != nil {
		utils.InternalServerError(c, "获取使用历史失败")
		return
	}

	utils.PagedSuccess(c, usedCodes, total, req.GetPage(), req.GetSize())
}

// CheckCodeAvailability 检查授权码是否可用
func (acc *AuthCodeController) CheckCodeAvailability(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		utils.BadRequest(c, "授权码不能为空")
		return
	}

	isAvailable, info, err := acc.authCodeService.CheckAvailability(code)
	if err != nil {
		utils.InternalServerError(c, "检查授权码失败")
		return
	}

	utils.Success(c, map[string]interface{}{
		"code":         code,
		"is_available": isAvailable,
		"info":         info,
	})
}