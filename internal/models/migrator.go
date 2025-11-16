package models

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Migrator 数据库迁移器（处理GORM不支持的操作）
type Migrator struct {
	db *gorm.DB
}

// NewMigrator 创建迁移器
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// Migration 迁移定义
type Migration struct {
	Version     string
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
}

// RunMigrations 执行所有迁移
func (m *Migrator) RunMigrations(migrations []Migration) error {
	// 创建迁移记录表
	if err := m.ensureMigrationsTable(); err != nil {
		return err
	}

	for _, migration := range migrations {
		// 检查是否已执行
		if m.isMigrationExecuted(migration.Version) {
			log.Printf("Migration %s already executed, skipping", migration.Version)
			continue
		}

		log.Printf("Running migration: %s - %s", migration.Version, migration.Description)

		// 执行迁移
		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Version, err)
		}

		// 记录迁移
		if err := m.recordMigration(migration.Version, migration.Description); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		log.Printf("Migration %s completed successfully", migration.Version)
	}

	return nil
}

// ensureMigrationsTable 确保迁移记录表存在
func (m *Migrator) ensureMigrationsTable() error {
	return m.db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INT AUTO_INCREMENT PRIMARY KEY,
			version VARCHAR(50) NOT NULL UNIQUE,
			description VARCHAR(255),
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

// isMigrationExecuted 检查迁移是否已执行
func (m *Migrator) isMigrationExecuted(version string) bool {
	var count int64
	m.db.Raw("SELECT COUNT(*) FROM migrations WHERE version = ?", version).Scan(&count)
	return count > 0
}

// recordMigration 记录迁移
func (m *Migrator) recordMigration(version, description string) error {
	return m.db.Exec(
		"INSERT INTO migrations (version, description) VALUES (?, ?)",
		version, description,
	).Error
}

// GetAllMigrations 获取所有迁移定义
func GetAllMigrations() []Migration {
	return []Migration{
		{
			Version:     "2024_01_01_remove_agent_old_field",
			Description: "删除agents表的旧字段示例",
			Up: func(db *gorm.DB) error {
				// 删除字段
				return db.Exec("ALTER TABLE agents DROP COLUMN IF EXISTS old_field").Error
			},
			Down: func(db *gorm.DB) error {
				// 回滚：重新添加字段
				return db.Exec("ALTER TABLE agents ADD COLUMN old_field VARCHAR(100)").Error
			},
		},
		{
			Version:     "2024_01_02_modify_username_length",
			Description: "修改username字段长度从50到100",
			Up: func(db *gorm.DB) error {
				return db.Exec("ALTER TABLE agents MODIFY COLUMN username VARCHAR(100)").Error
			},
			Down: func(db *gorm.DB) error {
				return db.Exec("ALTER TABLE agents MODIFY COLUMN username VARCHAR(50)").Error
			},
		},
		{
			Version:     "2024_01_03_add_agent_indexes",
			Description: "添加agents表性能索引",
			Up: func(db *gorm.DB) error {
				return db.Exec("CREATE INDEX IF NOT EXISTS idx_agents_phone ON agents(phone)").Error
			},
			Down: func(db *gorm.DB) error {
				return db.Exec("DROP INDEX IF EXISTS idx_agents_phone ON agents").Error
			},
		},
		// 添加更多迁移...
	}
}

// RunAllMigrations 执行所有迁移（供main.go调用）
func RunAllMigrations(db *gorm.DB) error {
	migrator := NewMigrator(db)
	migrations := GetAllMigrations()
	return migrator.RunMigrations(migrations)
}
