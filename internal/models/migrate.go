package models

import (
	"log"

	"github.com/ad-platform/backend/internal/database"
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
		&AgentCustomer{},
		&Commission{},
		&Withdrawal{},
		&AgentAuthCode{},
	}
}

// AutoMigrate 自动迁移所有表结构
func AutoMigrate() {
	db := database.GetDB()
	
	models := AllModels()
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("Failed to migrate model %T: %v", model, err)
		}
	}
	
	log.Println("Database migration completed successfully")
}

// CreateIndexes 创建额外的索引
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
		"CREATE INDEX IF NOT EXISTS idx_commissions_agent_status ON commissions(agent_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_withdrawals_agent_status ON withdrawals(agent_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_agent_customers_agent_id ON agent_customers(agent_id)",
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

// seedDefaultAgents 插入默认代理商测试数据
func seedDefaultAgents(db *gorm.DB) {
	// 检查是否已有代理商数据
	var count int64
	db.Model(&Agent{}).Count(&count)
	if count > 0 {
		log.Println("Agents already exist, skipping seed")
		return
	}

	// 创建一级代理商
	agent1 := Agent{
		Username:           "agent_zhang",
		Account:            "zhang@example.com",
		Password:           "agent123", // 将在BeforeCreate中自动加密
		RealName:           "张三",
		Phone:              "13800138001",
		Email:              "zhangsan@example.com",
		AgentLevel:         AgentLevelFirst,
		CommissionRate:     35.00,
		SelfCommissionRate: 30.00,
		Status:             AgentStatusActive,
		IsVerified:         true,
	}

	agent2 := Agent{
		Username:           "agent_li",
		Account:            "li@example.com",
		Password:           "agent123",
		RealName:           "李四",
		Phone:              "13800138002",
		Email:              "lisi@example.com",
		AgentLevel:         AgentLevelFirst,
		CommissionRate:     35.00,
		SelfCommissionRate: 30.00,
		Status:             AgentStatusActive,
		IsVerified:         true,
	}

	// 创建一级代理
	if err := db.Create(&agent1).Error; err != nil {
		log.Printf("Failed to create agent1: %v", err)
		return
	}
	if err := db.Create(&agent2).Error; err != nil {
		log.Printf("Failed to create agent2: %v", err)
		return
	}

	log.Printf("一级代理创建成功: %s (编码: %s), %s (编码: %s)",
		agent1.Username, agent1.AgentCode,
		agent2.Username, agent2.AgentCode)

	// 创建二级代理商（属于张三）
	agent1ID := agent1.ID
	agent3 := Agent{
		Username:           "agent_wang",
		Account:            "wang@example.com",
		Password:           "agent123",
		RealName:           "王五",
		Phone:              "13800138003",
		Email:              "wangwu@example.com",
		AgentLevel:         AgentLevelSecond,
		ParentID:           &agent1ID,
		CommissionRate:     25.00,
		SelfCommissionRate: 20.00,
		Status:             AgentStatusActive,
		IsVerified:         true,
	}

	if err := db.Create(&agent3).Error; err != nil {
		log.Printf("Failed to create agent3: %v", err)
		return
	}

	log.Printf("二级代理创建成功: %s (编码: %s, 上级: %s)",
		agent3.Username, agent3.AgentCode, agent1.Username)

	// 创建三级代理商（属于王五）
	agent3ID := agent3.ID
	agent4 := Agent{
		Username:           "agent_zhao",
		Account:            "zhao@example.com",
		Password:           "agent123",
		RealName:           "赵六",
		Phone:              "13800138004",
		Email:              "zhaoliu@example.com",
		AgentLevel:         AgentLevelThird,
		ParentID:           &agent3ID,
		CommissionRate:     20.00,
		SelfCommissionRate: 15.00,
		Status:             AgentStatusPending, // 待审核状态
		IsVerified:         false,
	}

	if err := db.Create(&agent4).Error; err != nil {
		log.Printf("Failed to create agent4: %v", err)
		return
	}

	log.Printf("三级代理创建成功: %s (编码: %s, 上级: %s, 状态: 待审核)",
		agent4.Username, agent4.AgentCode, agent3.Username)

	log.Println("✅ 代理商测试数据创建完成！")
	log.Println("测试账户:")
	log.Println("  - agent_zhang / agent123 (一级代理)")
	log.Println("  - agent_li / agent123 (一级代理)")
	log.Println("  - agent_wang / agent123 (二级代理)")
	log.Println("  - agent_zhao / agent123 (三级代理-待审核)")
}