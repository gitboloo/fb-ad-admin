package models

import (
	"log"
	"gorm.io/gorm"
)

// SeedPermissions 初始化权限数据
func SeedPermissions(db *gorm.DB) error {
	// 创建默认角色
	roles := []Role{
		{
			Name:        "超级管理员",
			Code:        "super_admin",
			Description: "拥有所有权限",
			Status:      1,
		},
		{
			Name:        "管理员",
			Code:        "admin",
			Description: "拥有大部分管理权限",
			Status:      1,
		},
		{
			Name:        "运营人员",
			Code:        "operator",
			Description: "负责日常运营管理",
			Status:      1,
		},
		{
			Name:        "查看者",
			Code:        "viewer",
			Description: "只能查看，不能操作",
			Status:      1,
		},
	}

	for _, role := range roles {
		var existingRole Role
		if err := db.Where("code = ?", role.Code).First(&existingRole).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&role).Error; err != nil {
				log.Printf("Failed to create role %s: %v", role.Code, err)
				return err
			}
			log.Printf("Created role: %s", role.Name)
		}
	}

	// 创建默认权限（菜单权限）
	permissions := []Permission{
		// 仪表盘
		{
			Name:      "仪表盘",
			Code:      "dashboard",
			Type:      PermissionTypeMenu,
			ParentID:  0,
			Path:      "/dashboard",
			Component: "dashboard/index",
			Icon:      "Dashboard",
			Sort:      1,
			Status:    1,
		},
		
		// 产品管理
		{
			Name:      "产品管理",
			Code:      "products",
			Type:      PermissionTypeMenu,
			ParentID:  0,
			Path:      "/products",
			Component: "Layout",
			Icon:      "Goods",
			Sort:      2,
			Status:    1,
		},
		{
			Name:      "产品列表",
			Code:      "products.list",
			Type:      PermissionTypePage,
			ParentID:  0, // 将在创建后更新
			Path:      "/products/list",
			Component: "products/ProductList",
			Sort:      1,
			Status:    1,
		},
		{
			Name:      "创建产品",
			Code:      "products.create",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		{
			Name:      "编辑产品",
			Code:      "products.edit",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		{
			Name:      "删除产品",
			Code:      "products.delete",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		
		// 计划管理
		{
			Name:      "计划管理",
			Code:      "campaigns",
			Type:      PermissionTypeMenu,
			ParentID:  0,
			Path:      "/campaigns",
			Component: "Layout",
			Icon:      "Promotion",
			Sort:      3,
			Status:    1,
		},
		{
			Name:      "计划列表",
			Code:      "campaigns.list",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/campaigns/list",
			Component: "campaigns/CampaignList",
			Sort:      1,
			Status:    1,
		},
		{
			Name:      "创建计划",
			Code:      "campaigns.create",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		{
			Name:      "编辑计划",
			Code:      "campaigns.edit",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		{
			Name:      "删除计划",
			Code:      "campaigns.delete",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		
		// 优惠券管理
		{
			Name:      "优惠券管理",
			Code:      "coupons",
			Type:      PermissionTypeMenu,
			ParentID:  0,
			Path:      "/coupons",
			Component: "Layout",
			Icon:      "Ticket",
			Sort:      4,
			Status:    1,
		},
		{
			Name:      "优惠券列表",
			Code:      "coupons.list",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/coupons/list",
			Component: "coupons/CouponList",
			Sort:      1,
			Status:    1,
		},
		{
			Name:      "创建优惠券",
			Code:      "coupons.create",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		{
			Name:      "编辑优惠券",
			Code:      "coupons.edit",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		{
			Name:      "删除优惠券",
			Code:      "coupons.delete",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		
		// 财务管理
		{
			Name:      "财务管理",
			Code:      "finance",
			Type:      PermissionTypeMenu,
			ParentID:  0,
			Path:      "/finance",
			Component: "Layout",
			Icon:      "Money",
			Sort:      5,
			Status:    1,
		},
		{
			Name:      "交易记录",
			Code:      "finance.transactions",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/finance/transactions",
			Component: "finance/TransactionList",
			Sort:      1,
			Status:    1,
		},
		{
			Name:      "充值管理",
			Code:      "finance.recharge",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/finance/recharge",
			Component: "finance/RechargeForm",
			Sort:      2,
			Status:    1,
		},
		{
			Name:      "提现管理",
			Code:      "finance.withdraw",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/finance/withdraw",
			Component: "finance/WithdrawForm",
			Sort:      3,
			Status:    1,
		},
		{
			Name:      "审核交易",
			Code:      "finance.audit",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		
		// 客户管理
		{
			Name:      "客户管理",
			Code:      "customers",
			Type:      PermissionTypeMenu,
			ParentID:  0,
			Path:      "/customers",
			Component: "Layout",
			Icon:      "User",
			Sort:      6,
			Status:    1,
		},
		{
			Name:      "客户列表",
			Code:      "customers.list",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/customers/list",
			Component: "customers/CustomerList",
			Sort:      1,
			Status:    1,
		},
		{
			Name:      "编辑客户",
			Code:      "customers.edit",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		{
			Name:      "删除客户",
			Code:      "customers.delete",
			Type:      PermissionTypeButton,
			ParentID:  0,
			Status:    1,
		},
		
		// 系统管理
		{
			Name:      "系统管理",
			Code:      "system",
			Type:      PermissionTypeMenu,
			ParentID:  0,
			Path:      "/system",
			Component: "Layout",
			Icon:      "Setting",
			Sort:      7,
			Status:    1,
		},
		{
			Name:      "管理员管理",
			Code:      "system.admins",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/system/admins",
			Component: "system/AdminList",
			Sort:      1,
			Status:    1,
		},
		{
			Name:      "角色管理",
			Code:      "system.roles",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/system/roles",
			Component: "system/RoleList",
			Sort:      2,
			Status:    1,
		},
		{
			Name:      "权限管理",
			Code:      "system.permissions",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/system/permissions",
			Component: "system/PermissionList",
			Sort:      3,
			Status:    1,
		},
		{
			Name:      "系统配置",
			Code:      "system.config",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/system/config",
			Component: "system/SystemConfig",
			Sort:      4,
			Status:    1,
		},
		
		// 统计分析
		{
			Name:      "统计分析",
			Code:      "statistics",
			Type:      PermissionTypeMenu,
			ParentID:  0,
			Path:      "/statistics",
			Component: "Layout",
			Icon:      "DataAnalysis",
			Sort:      8,
			Status:    1,
		},
		{
			Name:      "总览统计",
			Code:      "statistics.overview",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/statistics/overview",
			Component: "statistics/Overview",
			Sort:      1,
			Status:    1,
		},
		{
			Name:      "产品统计",
			Code:      "statistics.products",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/statistics/products",
			Component: "statistics/ProductStats",
			Sort:      2,
			Status:    1,
		},
		{
			Name:      "收入统计",
			Code:      "statistics.revenue",
			Type:      PermissionTypePage,
			ParentID:  0,
			Path:      "/statistics/revenue",
			Component: "statistics/RevenueStats",
			Sort:      3,
			Status:    1,
		},
	}

	// 创建权限并建立父子关系
	permMap := make(map[string]uint)
	for _, perm := range permissions {
		var existingPerm Permission
		if err := db.Where("code = ?", perm.Code).First(&existingPerm).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&perm).Error; err != nil {
				log.Printf("Failed to create permission %s: %v", perm.Code, err)
				return err
			}
			permMap[perm.Code] = perm.ID
			log.Printf("Created permission: %s", perm.Name)
		} else {
			permMap[perm.Code] = existingPerm.ID
		}
	}

	// 更新父子关系
	parentChildMap := map[string][]string{
		"products": {"products.list", "products.create", "products.edit", "products.delete"},
		"campaigns": {"campaigns.list", "campaigns.create", "campaigns.edit", "campaigns.delete"},
		"coupons": {"coupons.list", "coupons.create", "coupons.edit", "coupons.delete"},
		"finance": {"finance.transactions", "finance.recharge", "finance.withdraw", "finance.audit"},
		"customers": {"customers.list", "customers.edit", "customers.delete"},
		"system": {"system.admins", "system.roles", "system.permissions", "system.config"},
		"statistics": {"statistics.overview", "statistics.products", "statistics.revenue"},
	}

	for parentCode, childCodes := range parentChildMap {
		parentID := permMap[parentCode]
		for _, childCode := range childCodes {
			if childID, exists := permMap[childCode]; exists {
				db.Model(&Permission{}).Where("id = ?", childID).Update("parent_id", parentID)
			}
		}
	}

	// 为角色分配权限
	var superAdminRole Role
	if err := db.Where("code = ?", "super_admin").First(&superAdminRole).Error; err == nil {
		// 超级管理员拥有所有权限
		var allPerms []Permission
		db.Find(&allPerms)
		db.Model(&superAdminRole).Association("Permissions").Replace(&allPerms)
		log.Println("Assigned all permissions to super_admin role")
	}

	var adminRole Role
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err == nil {
		// 管理员拥有除系统管理外的所有权限
		var adminPerms []Permission
		db.Where("code NOT LIKE ?", "system.%").Find(&adminPerms)
		db.Model(&adminRole).Association("Permissions").Replace(&adminPerms)
		log.Println("Assigned permissions to admin role")
	}

	var operatorRole Role
	if err := db.Where("code = ?", "operator").First(&operatorRole).Error; err == nil {
		// 运营人员拥有产品、计划、优惠券、客户的权限
		var operatorPerms []Permission
		db.Where("code LIKE ? OR code LIKE ? OR code LIKE ? OR code LIKE ? OR code = ?",
			"products.%", "campaigns.%", "coupons.%", "customers.%", "dashboard").Find(&operatorPerms)
		db.Model(&operatorRole).Association("Permissions").Replace(&operatorPerms)
		log.Println("Assigned permissions to operator role")
	}

	var viewerRole Role
	if err := db.Where("code = ?", "viewer").First(&viewerRole).Error; err == nil {
		// 查看者只有查看权限
		var viewerPerms []Permission
		db.Where("type IN ? AND code NOT LIKE ? AND code NOT LIKE ? AND code NOT LIKE ?",
			[]string{PermissionTypeMenu, PermissionTypePage}, "%.create", "%.edit", "%.delete").Find(&viewerPerms)
		db.Model(&viewerRole).Association("Permissions").Replace(&viewerPerms)
		log.Println("Assigned permissions to viewer role")
	}

	// 为默认的admin用户分配超级管理员角色
	var defaultAdmin Admin
	if err := db.Where("username = ?", "admin").First(&defaultAdmin).Error; err == nil {
		var superRole Role
		if err := db.Where("code = ?", "super_admin").First(&superRole).Error; err == nil {
			db.Model(&defaultAdmin).Association("Roles").Replace([]Role{superRole})
			log.Println("Assigned super_admin role to default admin user")
		}
	}

	return nil
}