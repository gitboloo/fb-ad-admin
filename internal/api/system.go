package api

import (
	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/services"
	"github.com/ad-platform/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// SystemConfigRequest 系统配置请求结构
type SystemConfigRequest struct {
	Configs map[string]string `json:"configs" binding:"required"`
}

// BackupRequest 备份请求结构
type BackupRequest struct {
	Description string `json:"description" binding:"max=500"`
	Tables      []string `json:"tables"`
}

// RestoreRequest 恢复请求结构
type RestoreRequest struct {
	BackupFile string `json:"backup_file" binding:"required"`
}

// SystemController 系统控制器
type SystemController struct {
	systemService *services.SystemService
}

// NewSystemController 创建系统控制器
func NewSystemController() *SystemController {
	return &SystemController{
		systemService: services.NewSystemService(),
	}
}

// GetConfigs 获取系统配置
func (sc *SystemController) GetConfigs(c *gin.Context) {
	configs, err := sc.systemService.GetAllConfigs()
	if err != nil {
		utils.InternalServerError(c, "获取系统配置失败")
		return
	}

	utils.Success(c, configs)
}

// UpdateConfigs 更新系统配置
func (sc *SystemController) UpdateConfigs(c *gin.Context) {
	var req SystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := sc.systemService.UpdateConfigs(req.Configs); err != nil {
		utils.InternalServerError(c, "更新系统配置失败")
		return
	}

	utils.SuccessWithMessage(c, "系统配置更新成功", nil)
}

// GetConfig 获取单个配置
func (sc *SystemController) GetConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		utils.BadRequest(c, "配置键不能为空")
		return
	}

	config, err := sc.systemService.GetConfigByKey(key)
	if err != nil {
		if err.Error() == "配置不存在" {
			utils.NotFound(c, "配置不存在")
		} else {
			utils.InternalServerError(c, "获取配置失败")
		}
		return
	}

	utils.Success(c, config)
}

// UpdateConfig 更新单个配置
func (sc *SystemController) UpdateConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		utils.BadRequest(c, "配置键不能为空")
		return
	}

	var req struct {
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := sc.systemService.UpdateConfig(key, req.Value, req.Description); err != nil {
		utils.InternalServerError(c, "更新配置失败")
		return
	}

	utils.SuccessWithMessage(c, "配置更新成功", nil)
}

// GetStats 获取系统统计
func (sc *SystemController) GetStats(c *gin.Context) {
	stats, err := sc.systemService.GetSystemStats()
	if err != nil {
		utils.InternalServerError(c, "获取系统统计失败")
		return
	}

	utils.Success(c, stats)
}

// GetDashboard 获取仪表板数据
func (sc *SystemController) GetDashboard(c *gin.Context) {
	dashboard, err := sc.systemService.GetDashboardData()
	if err != nil {
		utils.InternalServerError(c, "获取仪表板数据失败")
		return
	}

	utils.Success(c, dashboard)
}

// Backup 系统备份
func (sc *SystemController) Backup(c *gin.Context) {
	var req BackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	backupInfo, err := sc.systemService.CreateBackup(req.Description, req.Tables)
	if err != nil {
		utils.InternalServerError(c, "系统备份失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "系统备份成功", backupInfo)
}

// Restore 系统恢复
func (sc *SystemController) Restore(c *gin.Context) {
	var req RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if err := sc.systemService.RestoreFromBackup(req.BackupFile); err != nil {
		utils.InternalServerError(c, "系统恢复失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "系统恢复成功", nil)
}

// GetBackups 获取备份列表
func (sc *SystemController) GetBackups(c *gin.Context) {
	backups, err := sc.systemService.GetBackupList()
	if err != nil {
		utils.InternalServerError(c, "获取备份列表失败")
		return
	}

	utils.Success(c, backups)
}

// DeleteBackup 删除备份
func (sc *SystemController) DeleteBackup(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		utils.BadRequest(c, "备份文件名不能为空")
		return
	}

	if err := sc.systemService.DeleteBackup(filename); err != nil {
		utils.InternalServerError(c, "删除备份失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "备份删除成功", nil)
}

// GetSystemInfo 获取系统信息
func (sc *SystemController) GetSystemInfo(c *gin.Context) {
	info, err := sc.systemService.GetSystemInfo()
	if err != nil {
		utils.InternalServerError(c, "获取系统信息失败")
		return
	}

	utils.Success(c, info)
}

// UpdateSystemInfo 更新系统信息
func (sc *SystemController) UpdateSystemInfo(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,max=255"`
		Logo        string `json:"logo" binding:"max=500"`
		Description string `json:"description" binding:"max=1000"`
		ContactEmail string `json:"contact_email" binding:"omitempty,email,max=255"`
		ContactPhone string `json:"contact_phone" binding:"max=20"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	configs := map[string]string{
		models.ConfigKeySystemName:        req.Name,
		models.ConfigKeySystemLogo:        req.Logo,
		models.ConfigKeySystemDescription: req.Description,
		models.ConfigKeyContactEmail:      req.ContactEmail,
		models.ConfigKeyContactPhone:      req.ContactPhone,
	}

	if err := sc.systemService.UpdateConfigs(configs); err != nil {
		utils.InternalServerError(c, "更新系统信息失败")
		return
	}

	utils.SuccessWithMessage(c, "系统信息更新成功", nil)
}

// GetMaintenanceMode 获取维护模式状态
func (sc *SystemController) GetMaintenanceMode(c *gin.Context) {
	isEnabled, err := sc.systemService.IsMaintenanceModeEnabled()
	if err != nil {
		utils.InternalServerError(c, "获取维护模式状态失败")
		return
	}

	utils.Success(c, map[string]interface{}{
		"maintenance_mode": isEnabled,
	})
}

// SetMaintenanceMode 设置维护模式
func (sc *SystemController) SetMaintenanceMode(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	value := "false"
	if req.Enabled {
		value = "true"
	}

	if err := sc.systemService.UpdateConfig(models.ConfigKeyMaintenanceMode, value, "维护模式开关"); err != nil {
		utils.InternalServerError(c, "设置维护模式失败")
		return
	}

	message := "维护模式已关闭"
	if req.Enabled {
		message = "维护模式已开启"
	}

	utils.SuccessWithMessage(c, message, map[string]interface{}{
		"maintenance_mode": req.Enabled,
	})
}

// CleanSystem 系统清理
func (sc *SystemController) CleanSystem(c *gin.Context) {
	var req struct {
		CleanLogs      bool `json:"clean_logs"`
		CleanTempFiles bool `json:"clean_temp_files"`
		CleanExpired   bool `json:"clean_expired"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	result, err := sc.systemService.CleanSystem(req.CleanLogs, req.CleanTempFiles, req.CleanExpired)
	if err != nil {
		utils.InternalServerError(c, "系统清理失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "系统清理完成", result)
}

// GetHealth 获取系统健康状态
func (sc *SystemController) GetHealth(c *gin.Context) {
	health, err := sc.systemService.GetHealthStatus()
	if err != nil {
		utils.InternalServerError(c, "获取健康状态失败")
		return
	}

	utils.Success(c, health)
}

// InitSystem 初始化系统
func (sc *SystemController) InitSystem(c *gin.Context) {
	if err := sc.systemService.InitializeSystem(); err != nil {
		utils.InternalServerError(c, "系统初始化失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "系统初始化成功", nil)
}

// ResetSystem 重置系统
func (sc *SystemController) ResetSystem(c *gin.Context) {
	var req struct {
		ConfirmCode string `json:"confirm_code" binding:"required"`
		ResetData   bool   `json:"reset_data"`
		ResetConfig bool   `json:"reset_config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	// 验证确认码
	if req.ConfirmCode != "RESET_SYSTEM_CONFIRM" {
		utils.BadRequest(c, "确认码错误")
		return
	}

	if err := sc.systemService.ResetSystem(req.ResetData, req.ResetConfig); err != nil {
		utils.InternalServerError(c, "系统重置失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "系统重置成功", nil)
}

// ExportData 导出数据
func (sc *SystemController) ExportData(c *gin.Context) {
	var req struct {
		Tables []string `json:"tables"`
		Format string   `json:"format" binding:"oneof=json csv sql"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	if req.Format == "" {
		req.Format = "json"
	}

	exportInfo, err := sc.systemService.ExportData(req.Tables, req.Format)
	if err != nil {
		utils.InternalServerError(c, "数据导出失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "数据导出成功", exportInfo)
}

// ImportData 导入数据
func (sc *SystemController) ImportData(c *gin.Context) {
	var req struct {
		FilePath string `json:"file_path" binding:"required"`
		Format   string `json:"format" binding:"oneof=json csv sql"`
		Override bool   `json:"override"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	result, err := sc.systemService.ImportData(req.FilePath, req.Format, req.Override)
	if err != nil {
		utils.InternalServerError(c, "数据导入失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "数据导入成功", result)
}