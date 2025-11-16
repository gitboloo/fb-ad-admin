package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/repositories"
)

// SystemService 系统服务
type SystemService struct {
	systemConfigRepo *repositories.SystemConfigRepository
	productRepo      *repositories.ProductRepository
	campaignRepo     *repositories.CampaignRepository
	customerRepo     *repositories.CustomerRepository
	transactionRepo  *repositories.TransactionRepository
	couponRepo       *repositories.CouponRepository
	authCodeRepo     *repositories.AuthCodeRepository
}

// NewSystemService 创建系统服务
func NewSystemService() *SystemService {
	return &SystemService{
		systemConfigRepo: repositories.NewSystemConfigRepository(),
		productRepo:      repositories.NewProductRepository(),
		campaignRepo:     repositories.NewCampaignRepository(),
		customerRepo:     repositories.NewCustomerRepository(),
		transactionRepo:  repositories.NewTransactionRepository(),
		couponRepo:       repositories.NewCouponRepository(),
		authCodeRepo:     repositories.NewAuthCodeRepository(),
	}
}

// GetAllConfigs 获取所有系统配置
func (ss *SystemService) GetAllConfigs() (map[string]*models.SystemConfig, error) {
	return ss.systemConfigRepo.GetAllConfigs()
}

// GetConfigByKey 根据键获取配置
func (ss *SystemService) GetConfigByKey(key string) (*models.SystemConfig, error) {
	return ss.systemConfigRepo.GetByKey(key)
}

// UpdateConfigs 批量更新配置
func (ss *SystemService) UpdateConfigs(configs map[string]string) error {
	return ss.systemConfigRepo.UpdateConfigs(configs)
}

// UpdateConfig 更新单个配置
func (ss *SystemService) UpdateConfig(key, value, description string) error {
	config, err := ss.systemConfigRepo.GetByKey(key)
	if err != nil {
		// 如果配置不存在，创建新的
		config = &models.SystemConfig{
			Key:         key,
			Value:       value,
			Description: description,
		}
		return ss.systemConfigRepo.Create(config)
	}

	config.Value = value
	if description != "" {
		config.Description = description
	}
	return ss.systemConfigRepo.Update(config)
}

// GetSystemStats 获取系统统计
func (ss *SystemService) GetSystemStats() (map[string]interface{}, error) {
	// 获取各模块统计
	productStats, _ := ss.productRepo.GetStatistics()
	campaignStats, _ := ss.campaignRepo.GetStatistics()
	customerStats, _ := ss.customerRepo.GetStatistics()
	transactionStats, _ := ss.transactionRepo.GetStatistics()
	couponStats, _ := ss.couponRepo.GetStatistics()
	authCodeStats, _ := ss.authCodeRepo.GetStatistics()

	// 系统运行信息
	systemInfo := ss.getSystemRuntime()

	return map[string]interface{}{
		"products":     productStats,
		"campaigns":    campaignStats,
		"customers":    customerStats,
		"transactions": transactionStats,
		"coupons":      couponStats,
		"auth_codes":   authCodeStats,
		"system":       systemInfo,
	}, nil
}

// GetDashboardData 获取仪表板数据
func (ss *SystemService) GetDashboardData() (map[string]interface{}, error) {
	// 获取各种计数
	productCount, _ := ss.productRepo.GetStatistics()
	customerCount, _ := ss.customerRepo.GetActiveCustomerCount()
	totalTransactions, _ := ss.transactionRepo.GetTotalAmount()
	totalBalance, _ := ss.customerRepo.GetTotalBalance()

	// 最近活动
	recentTransactions, _ := ss.transactionRepo.GetRecentTransactions(10)
	recentCustomers, _ := ss.customerRepo.GetRecentCustomers(10)

	// 系统健康状态
	health, _ := ss.GetHealthStatus()

	return map[string]interface{}{
		"overview": map[string]interface{}{
			"total_products":     productCount.Total,
			"active_customers":   customerCount,
			"total_transactions": totalTransactions,
			"total_balance":      totalBalance,
		},
		"recent_activity": map[string]interface{}{
			"transactions": recentTransactions,
			"customers":    recentCustomers,
		},
		"health": health,
	}, nil
}

// CreateBackup 创建系统备份
func (ss *SystemService) CreateBackup(description string, tables []string) (map[string]interface{}, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s.sql", timestamp)
	backupPath := filepath.Join("backups", filename)

	// 确保备份目录存在
	if err := os.MkdirAll("backups", 0755); err != nil {
		return nil, err
	}

	// 这里应该实现实际的数据库备份逻辑
	// 暂时创建一个简单的备份信息文件
	backupInfo := map[string]interface{}{
		"filename":    filename,
		"path":        backupPath,
		"description": description,
		"tables":      tables,
		"created_at":  time.Now(),
		"size":        "0 MB", // 实际实现时应该计算真实大小
	}

	// 保存备份信息
	backupInfoJSON, _ := json.Marshal(backupInfo)
	if err := os.WriteFile(backupPath+".info", backupInfoJSON, 0644); err != nil {
		return nil, err
	}

	return backupInfo, nil
}

// RestoreFromBackup 从备份恢复系统
func (ss *SystemService) RestoreFromBackup(backupFile string) error {
	backupPath := filepath.Join("backups", backupFile)
	
	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return &ServiceError{
			Code:    404,
			Message: "备份文件不存在",
		}
	}

	// 这里应该实现实际的数据库恢复逻辑
	// 暂时返回成功
	return nil
}

// GetBackupList 获取备份列表
func (ss *SystemService) GetBackupList() ([]map[string]interface{}, error) {
	backupsDir := "backups"
	var backups []map[string]interface{}

	// 确保备份目录存在
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		return backups, nil
	}

	files, err := os.ReadDir(backupsDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".info" {
			infoPath := filepath.Join(backupsDir, file.Name())
			data, err := os.ReadFile(infoPath)
			if err != nil {
				continue
			}

			var backupInfo map[string]interface{}
			if err := json.Unmarshal(data, &backupInfo); err != nil {
				continue
			}

			backups = append(backups, backupInfo)
		}
	}

	return backups, nil
}

// DeleteBackup 删除备份
func (ss *SystemService) DeleteBackup(filename string) error {
	backupPath := filepath.Join("backups", filename)
	infoPath := backupPath + ".info"

	// 删除备份文件和信息文件
	os.Remove(backupPath)
	os.Remove(infoPath)

	return nil
}

// GetSystemInfo 获取系统信息
func (ss *SystemService) GetSystemInfo() (map[string]interface{}, error) {
	configs, err := ss.systemConfigRepo.GetAllConfigs()
	if err != nil {
		return nil, err
	}

	systemInfo := map[string]interface{}{
		"name":         "",
		"logo":         "",
		"description":  "",
		"contact_email": "",
		"contact_phone": "",
		"version":      "1.0.0",
		"runtime":      ss.getSystemRuntime(),
	}

	// 从配置中获取系统信息
	if config, exists := configs[models.ConfigKeySystemName]; exists {
		systemInfo["name"] = config.Value
	}
	if config, exists := configs[models.ConfigKeySystemLogo]; exists {
		systemInfo["logo"] = config.Value
	}
	if config, exists := configs[models.ConfigKeySystemDescription]; exists {
		systemInfo["description"] = config.Value
	}
	if config, exists := configs[models.ConfigKeyContactEmail]; exists {
		systemInfo["contact_email"] = config.Value
	}
	if config, exists := configs[models.ConfigKeyContactPhone]; exists {
		systemInfo["contact_phone"] = config.Value
	}

	return systemInfo, nil
}

// IsMaintenanceModeEnabled 检查是否启用维护模式
func (ss *SystemService) IsMaintenanceModeEnabled() (bool, error) {
	config, err := ss.systemConfigRepo.GetByKey(models.ConfigKeyMaintenanceMode)
	if err != nil {
		return false, nil // 默认不启用
	}
	return config.IsEnabled(), nil
}

// CleanSystem 系统清理
func (ss *SystemService) CleanSystem(cleanLogs, cleanTempFiles, cleanExpired bool) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"logs_cleaned":       0,
		"temp_files_cleaned": 0,
		"expired_cleaned":    0,
	}

	if cleanLogs {
		// 清理日志文件
		logsCount := ss.cleanLogFiles()
		result["logs_cleaned"] = logsCount
	}

	if cleanTempFiles {
		// 清理临时文件
		tempCount := ss.cleanTempFiles()
		result["temp_files_cleaned"] = tempCount
	}

	if cleanExpired {
		// 清理过期数据
		expiredCount := ss.cleanExpiredData()
		result["expired_cleaned"] = expiredCount
	}

	return result, nil
}

// GetHealthStatus 获取系统健康状态
func (ss *SystemService) GetHealthStatus() (map[string]interface{}, error) {
	health := map[string]interface{}{
		"status": "healthy",
		"checks": map[string]interface{}{},
	}

	checks := health["checks"].(map[string]interface{})

	// 数据库连接检查
	if err := ss.systemConfigRepo.HealthCheck(); err != nil {
		checks["database"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
		health["status"] = "unhealthy"
	} else {
		checks["database"] = map[string]interface{}{
			"status": "ok",
		}
	}

	// 磁盘空间检查
	diskUsage := ss.getDiskUsage()
	if diskUsage > 90 {
		checks["disk"] = map[string]interface{}{
			"status": "warning",
			"usage":  diskUsage,
		}
		if health["status"] == "healthy" {
			health["status"] = "warning"
		}
	} else {
		checks["disk"] = map[string]interface{}{
			"status": "ok",
			"usage":  diskUsage,
		}
	}

	// 内存使用检查
	memUsage := ss.getMemoryUsage()
	checks["memory"] = map[string]interface{}{
		"status": "ok",
		"usage":  memUsage,
	}

	return health, nil
}

// InitializeSystem 初始化系统
func (ss *SystemService) InitializeSystem() error {
	// 初始化默认配置
	defaultConfigs := models.GetDefaultConfigs()
	for _, config := range defaultConfigs {
		if err := ss.systemConfigRepo.CreateOrUpdate(&config); err != nil {
			return err
		}
	}

	return nil
}

// ResetSystem 重置系统
func (ss *SystemService) ResetSystem(resetData, resetConfig bool) error {
	if resetData {
		// 清理所有业务数据
		// 这里应该实现数据清理逻辑
	}

	if resetConfig {
		// 重置配置到默认值
		return ss.InitializeSystem()
	}

	return nil
}

// ExportData 导出数据
func (ss *SystemService) ExportData(tables []string, format string) (map[string]interface{}, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("export_%s.%s", timestamp, format)
	exportPath := filepath.Join("exports", filename)

	// 确保导出目录存在
	if err := os.MkdirAll("exports", 0755); err != nil {
		return nil, err
	}

	// 这里应该实现实际的数据导出逻辑
	exportInfo := map[string]interface{}{
		"filename":   filename,
		"path":       exportPath,
		"format":     format,
		"tables":     tables,
		"created_at": time.Now(),
		"size":       "0 MB",
	}

	return exportInfo, nil
}

// ImportData 导入数据
func (ss *SystemService) ImportData(filePath, format string, override bool) (map[string]interface{}, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, &ServiceError{
			Code:    404,
			Message: "导入文件不存在",
		}
	}

	// 这里应该实现实际的数据导入逻辑
	result := map[string]interface{}{
		"imported_records": 0,
		"skipped_records":  0,
		"error_records":    0,
		"success":          true,
	}

	return result, nil
}

// 辅助方法

func (ss *SystemService) getSystemRuntime() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"go_version":    runtime.Version(),
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH,
		"cpu_count":     runtime.NumCPU(),
		"goroutines":    runtime.NumGoroutine(),
		"memory_alloc":  m.Alloc,
		"memory_total":  m.TotalAlloc,
		"memory_sys":    m.Sys,
		"gc_count":      m.NumGC,
	}
}

func (ss *SystemService) cleanLogFiles() int {
	// 实现日志清理逻辑
	return 0
}

func (ss *SystemService) cleanTempFiles() int {
	// 实现临时文件清理逻辑
	return 0
}

func (ss *SystemService) cleanExpiredData() int {
	// 实现过期数据清理逻辑
	count := 0

	// 清理过期的授权码
	if expiredCodes, err := ss.authCodeRepo.GetExpiredCodesForCleanup(); err == nil {
		for _, code := range expiredCodes {
			if code.Status == models.AuthCodeStatusUnused {
				code.Expire()
				ss.authCodeRepo.Update(code)
				count++
			}
		}
	}

	return count
}

func (ss *SystemService) getDiskUsage() float64 {
	// 实现磁盘使用率检查
	return 50.0 // 示例值
}

func (ss *SystemService) getMemoryUsage() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"alloc":       m.Alloc / 1024 / 1024,       // MB
		"total_alloc": m.TotalAlloc / 1024 / 1024, // MB
		"sys":         m.Sys / 1024 / 1024,        // MB
		"num_gc":      m.NumGC,
	}
}