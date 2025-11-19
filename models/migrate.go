package models

import (
	"log"

	"backend/database"

	"gorm.io/gorm"
)

// AllModels 包含所有需要迁移的模型
func AllModels() []interface{} {
	return []interface{}{
		&Admin{},
		&Role{},
		&Permission{},
		&AdminRoleAssoc{},
		&RolePermission{},
		&Product{},
		&Campaign{},
		&Coupon{},
		&UserCoupon{},
		&AuthCode{},
		&Transaction{},
		&Customer{},
		&SystemConfig{},

		// 代理商系统模型
		&Agent{},
	}
}

// AutoMigrate 自动迁移所有表结构
func AutoMigrate() {
	db := database.GetDB()

	// 先删除可能存在的问题外键
	log.Println("Dropping problematic foreign keys if they exist...")
	var fkCount int64
	db.Raw(`SELECT count(*) FROM INFORMATION_SCHEMA.table_constraints
		WHERE constraint_schema = DATABASE()
		AND table_name = 'permissions'
		AND constraint_name = 'fk_permissions_children'`).Scan(&fkCount)

	if fkCount > 0 {
		log.Println("Found fk_permissions_children, dropping it...")
		db.Exec("ALTER TABLE permissions DROP FOREIGN KEY fk_permissions_children")
	}

	models := AllModels()
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("Failed to migrate model %T: %v", model, err)
		}
	}

	log.Println("Database migration completed successfully")

	// 执行额外的数据库迁移操作
	CleanupOldAgentTables()
	AddAgentInviteCode()
}

// CleanupOldAgentTables 删除旧的代理商相关表
func CleanupOldAgentTables() {
	db := database.GetDB()

	log.Println("Cleaning up old agent tables...")

	// 删除 agent_customers 表
	if db.Migrator().HasTable("agent_customers") {
		log.Println("Dropping table agent_customers...")
		if err := db.Exec("DROP TABLE IF EXISTS agent_customers").Error; err != nil {
			log.Printf("Warning: Failed to drop agent_customers table: %v", err)
		} else {
			log.Println("✅ agent_customers table dropped successfully")
		}
	}

	// 删除 agent_auth_codes 表
	if db.Migrator().HasTable("agent_auth_codes") {
		log.Println("Dropping table agent_auth_codes...")
		if err := db.Exec("DROP TABLE IF EXISTS agent_auth_codes").Error; err != nil {
			log.Printf("Warning: Failed to drop agent_auth_codes table: %v", err)
		} else {
			log.Println("✅ agent_auth_codes table dropped successfully")
		}
	}
}

// AddAgentInviteCode 为 agents 表添加 invite_code 字段
func AddAgentInviteCode() {
	db := database.GetDB()

	log.Println("Checking if agents table needs invite_code column...")

	if !db.Migrator().HasColumn("agents", "invite_code") {
		log.Println("Adding invite_code column to agents table...")
		if err := db.Migrator().AddColumn(&Agent{}, "invite_code"); err != nil {
			log.Printf("Warning: Failed to add invite_code column: %v", err)
			return
		}
		log.Println("✅ invite_code column added successfully")

		// 为所有没有邀请码的代理商生成邀请码
		log.Println("Generating invite codes for existing agents...")
		var agents []Agent
		if err := db.Where("invite_code = '' OR invite_code IS NULL").Find(&agents).Error; err == nil && len(agents) > 0 {
			for i := range agents {
				agents[i].InviteCode = generateInviteCode()
				if err := db.Model(&agents[i]).Update("invite_code", agents[i].InviteCode).Error; err != nil {
					log.Printf("Warning: Failed to update invite_code for agent %d: %v", agents[i].ID, err)
				}
			}
			log.Printf("✅ Generated invite codes for %d agents", len(agents))
		}

		// 创建UNIQUE索引（只考虑非空值）
		log.Println("Creating UNIQUE constraint on invite_code...")
		if err := db.Exec(`
			ALTER TABLE agents 
			ADD CONSTRAINT uk_agents_invite_code UNIQUE(invite_code)
		`).Error; err != nil {
			log.Printf("Warning: Failed to create UNIQUE constraint: %v", err)
		} else {
			log.Println("✅ UNIQUE constraint created successfully")
		}
	} else {
		log.Println("✓ invite_code column already exists in agents table")

		// 检查是否有空值
		var count int64
		if err := db.Model(&Agent{}).Where("invite_code = '' OR invite_code IS NULL").Count(&count).Error; err == nil && count > 0 {
			log.Printf("Found %d agents without invite codes, generating...", count)

			var agents []Agent
			if err := db.Where("invite_code = '' OR invite_code IS NULL").Find(&agents).Error; err == nil && len(agents) > 0 {
				for i := range agents {
					agents[i].InviteCode = generateInviteCode()
					if err := db.Model(&agents[i]).Update("invite_code", agents[i].InviteCode).Error; err != nil {
						log.Printf("Warning: Failed to update invite_code for agent %d: %v", agents[i].ID, err)
					}
				}
				log.Printf("✅ Generated invite codes for %d agents", len(agents))
			}
		}
	}
} // CreateIndexes 创建额外的索引
func CreateIndexes() {
	db := database.GetDB()

	// 为常用查询添加复合索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_products_type_status ON products(type, status)",
		"CREATE INDEX IF NOT EXISTS idx_campaigns_product_status ON campaigns(product_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_user_coupons_user_status ON user_coupons(user_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_user_type_status ON transactions(user_id, type, status)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_auth_codes_status_expired ON auth_codes(status, expired_at)",

		// 代理商系统索引
		"CREATE INDEX IF NOT EXISTS idx_agents_status_level ON agents(status, agent_level)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %s, error: %v", indexSQL, err)
		}
	}

	log.Println("Database indexes created successfully")
}

// SeedDefaultData 插入默认数据
func SeedDefaultData() {
	db := database.GetDB()

	// 插入默认系统配置
	seedSystemConfigs(db)

	// 插入默认管理员账户
	seedDefaultAdmin(db)

	// 插入默认权限和角色数据
	if err := SeedPermissions(db); err != nil {
		log.Printf("Failed to seed permissions: %v", err)
	}

	// 插入默认代理商测试数据
	seedDefaultAgents(db)

	log.Println("Default data seeded successfully")
}

// seedSystemConfigs 插入默认系统配置
func seedSystemConfigs(db *gorm.DB) {
	configs := GetDefaultConfigs()

	for _, config := range configs {
		var existing SystemConfig
		if err := db.Where("key = ?", config.Key).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&config).Error; err != nil {
					log.Printf("Failed to create system config %s: %v", config.Key, err)
				}
			}
		}
	}
}

// seedDefaultAdmin 插入默认管理员账户
func seedDefaultAdmin(db *gorm.DB) {
	var admin Admin
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			defaultAdmin := Admin{
				Username: "admin",
				Account:  "admin@example.com",
				Password: "admin123", // 将在BeforeCreate中自动加密
				Role:     AdminRoleSuperAdmin,
				Status:   AdminStatusActive,
			}

			if err := db.Create(&defaultAdmin).Error; err != nil {
				log.Printf("Failed to create default admin: %v", err)
			} else {
				log.Println("Default admin created: username=admin, password=admin123")
			}
		}
	}
}

// seedDefaultAgents 插入默认代理商测试数据（每个代理关联一个Admin）
func seedDefaultAgents(db *gorm.DB) {
	// 检查是否已有代理商数据
	var count int64
	db.Model(&Agent{}).Count(&count)
	if count > 0 {
		log.Println("Agents already exist, skipping seed")
		return
	}

	// 获取默认admin用户（一级代理的创建者）
	var defaultAdmin Admin
	if err := db.Where("username = ?", "admin").First(&defaultAdmin).Error; err != nil {
		log.Printf("Warning: Default admin not found, skipping agent seed: %v", err)
		return
	}

	// ===== 创建一级代理 =====
	// 一级代理1 - 为其创建新的Admin账户
	agent1Admin := Admin{
		Username: "agent_level1_1",
		Account:  "agent_level1_1@example.com",
		Password: "agent123",
		Role:     AdminRoleUser,
		Status:   AdminStatusActive,
	}
	if err := db.Create(&agent1Admin).Error; err != nil {
		log.Printf("Failed to create agent1 admin: %v", err)
		return
	}

	agent1 := Agent{
		AdminID:                   agent1Admin.ID,
		Status:                    AgentStatusActive,
		AgentLevel:                AgentLevelFirst,
		EnableGoogleAuth:          false,
		CanDispatchOrders:         true,
		CanModifyCustomerBankCard: false,
		Remark:                    "一级代理商1 - 创建者: admin",
	}
	if err := db.Create(&agent1).Error; err != nil {
		log.Printf("Failed to create agent1: %v", err)
		return
	}

	// 一级代理2 - 为其创建新的Admin账户
	agent2Admin := Admin{
		Username: "agent_level1_2",
		Account:  "agent_level1_2@example.com",
		Password: "agent123",
		Role:     AdminRoleUser,
		Status:   AdminStatusActive,
	}
	if err := db.Create(&agent2Admin).Error; err != nil {
		log.Printf("Failed to create agent2 admin: %v", err)
		return
	}

	agent2 := Agent{
		AdminID:                   agent2Admin.ID,
		Status:                    AgentStatusActive,
		AgentLevel:                AgentLevelFirst,
		EnableGoogleAuth:          false,
		CanDispatchOrders:         true,
		CanModifyCustomerBankCard: false,
		Remark:                    "一级代理商2 - 创建者: admin",
	}
	if err := db.Create(&agent2).Error; err != nil {
		log.Printf("Failed to create agent2: %v", err)
		return
	}

	log.Printf("✅ 一级代理创建成功2个 (InviteCodes: %s, %s)",
		agent1.InviteCode, agent2.InviteCode)

	// ===== 创建二级代理（上级为agent1） =====
	agent3Admin := Admin{
		Username: "agent_level2_1",
		Account:  "agent_level2_1@example.com",
		Password: "agent123",
		Role:     AdminRoleUser,
		Status:   AdminStatusActive,
	}
	if err := db.Create(&agent3Admin).Error; err != nil {
		log.Printf("Failed to create agent3 admin: %v", err)
		return
	}

	agent3 := Agent{
		AdminID:                   agent3Admin.ID,
		Status:                    AgentStatusActive,
		AgentLevel:                AgentLevelSecond,
		ParentAdminID:             &agent1Admin.ID, // 上级是一级代理1
		EnableGoogleAuth:          false,
		CanDispatchOrders:         true,
		CanModifyCustomerBankCard: false,
		Remark:                    "二级代理商 - 上级为一级代理1",
	}
	if err := db.Create(&agent3).Error; err != nil {
		log.Printf("Failed to create agent3: %v", err)
		return
	}

	log.Printf("✅ 二级代理创建成功: %s (上级AdminID: %d)",
		agent3.InviteCode, agent1Admin.ID)

	// ===== 创建三级代理（上级为agent3） =====
	agent4Admin := Admin{
		Username: "agent_level3_1",
		Account:  "agent_level3_1@example.com",
		Password: "agent123",
		Role:     AdminRoleUser,
		Status:   AdminStatusActive,
	}
	if err := db.Create(&agent4Admin).Error; err != nil {
		log.Printf("Failed to create agent4 admin: %v", err)
		return
	}

	agent4 := Agent{
		AdminID:                   agent4Admin.ID,
		Status:                    AgentStatusActive,
		AgentLevel:                AgentLevelThird,
		ParentAdminID:             &agent3Admin.ID, // 上级是二级代理
		EnableGoogleAuth:          false,
		CanDispatchOrders:         false,
		CanModifyCustomerBankCard: false,
		Remark:                    "三级代理商 - 上级为二级代理",
	}
	if err := db.Create(&agent4).Error; err != nil {
		log.Printf("Failed to create agent4: %v", err)
		return
	}

	log.Printf("✅ 三级代理创建成功: %s (上级AdminID: %d)",
		agent4.InviteCode, agent3Admin.ID)

	log.Println("\n✅ 代理商测试数据创建完成！")
	log.Println("【一级代理1】")
	log.Println("  - Admin: agent_level1_1 / agent123")
	log.Println("  - InviteCode: " + agent1.InviteCode)
	log.Println("【一级代理2】")
	log.Println("  - Admin: agent_level1_2 / agent123")
	log.Println("  - InviteCode: " + agent2.InviteCode)
	log.Println("【二级代理】(上级: 一级代理1)")
	log.Println("  - Admin: agent_level2_1 / agent123")
	log.Println("  - InviteCode: " + agent3.InviteCode)
	log.Println("【三级代理】(上级: 二级代理)")
	log.Println("  - Admin: agent_level3_1 / agent123")
	log.Println("  - InviteCode: " + agent4.InviteCode)
}
