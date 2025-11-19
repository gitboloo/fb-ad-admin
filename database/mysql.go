package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"backend/configs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitMySQL() {
	cfg := configs.AppConfig.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	var err error
	logLevel := logger.Info
	if configs.AppConfig.Server.Env == "production" {
		logLevel = logger.Error
	}

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		// 禁用默认事务（提高性能）
		SkipDefaultTransaction: true,
		// 预编译语句缓存
		PrepareStmt: true,
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 获取底层的 *sql.DB 以配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying SQL database:", err)
	}

	// 配置连接池参数
	configureConnectionPool(sqlDB)

	log.Printf("Database connected successfully: %s@%s:%d/%s", 
		cfg.User, cfg.Host, cfg.Port, cfg.DBName)
	
	// 打印连接池状态
	stats := sqlDB.Stats()
	log.Printf("Connection pool - MaxOpenConns: %d, OpenConns: %d, InUse: %d, Idle: %d",
		stats.MaxOpenConnections, stats.OpenConnections, stats.InUse, stats.Idle)
}

// 配置数据库连接池参数
func configureConnectionPool(sqlDB *sql.DB) {
	// 设置最大空闲连接数
	// 空闲连接数应该根据平均负载设置，太少会导致频繁创建连接，太多会浪费资源
	sqlDB.SetMaxIdleConns(10)
	
	// 设置最大打开连接数
	// 根据数据库服务器的承载能力和应用需求设置
	// MySQL默认最大连接数通常是151，需要给其他服务预留
	sqlDB.SetMaxOpenConns(100)
	
	// 设置连接的最大生命周期
	// 避免使用太旧的连接，MySQL默认wait_timeout是8小时
	// 设置略小于wait_timeout的值，避免使用已经被MySQL关闭的连接
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	// 设置连接的最大空闲时间
	// 超过这个时间的空闲连接将被关闭
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)
}

// 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// 检查数据库连接健康状态
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	
	// Ping数据库检查连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	return nil
}

// 获取连接池统计信息
func GetPoolStats() (stats sql.DBStats, err error) {
	if DB == nil {
		return stats, fmt.Errorf("database not initialized")
	}
	
	sqlDB, err := DB.DB()
	if err != nil {
		return stats, err
	}
	
	return sqlDB.Stats(), nil
}

// 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		
		// 打印最终的连接池状态
		stats := sqlDB.Stats()
		log.Printf("Closing database - Final stats: OpenConns: %d, InUse: %d, Idle: %d",
			stats.OpenConnections, stats.InUse, stats.Idle)
		
		return sqlDB.Close()
	}
	return nil
}