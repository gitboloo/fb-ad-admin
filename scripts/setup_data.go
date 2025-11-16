package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 连接数据库
	dsn := "root:@tcp(localhost:3306)/ad_platform?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 生成root密码hash
	rootPassword, _ := bcrypt.GenerateFromPassword([]byte("root123"), bcrypt.DefaultCost)
	
	fmt.Println("=== 设置账户权限 ===")
	
	// 1. 创建root超级管理员账户
	_, err = db.Exec(`
		INSERT INTO admins (username, account, password, role, status, created_at, updated_at) 
		VALUES (?, ?, ?, 3, 1, NOW(), NOW())
		ON DUPLICATE KEY UPDATE role = 3, status = 1`,
		"root", "root@platform.com", string(rootPassword))
	if err != nil {
		fmt.Printf("创建root账户失败: %v\n", err)
	} else {
		fmt.Println("✅ root超级管理员账户创建成功 (密码: root123)")
	}
	
	// 2. 将admin改为普通管理员
	_, err = db.Exec("UPDATE admins SET role = 2 WHERE username = 'admin'")
	if err != nil {
		fmt.Printf("更新admin角色失败: %v\n", err)
	} else {
		fmt.Println("✅ admin账户已调整为普通管理员")
	}
	
	fmt.Println("\n=== 插入测试数据 ===")
	
	// 3. 插入产品数据
	products := []struct {
		name, productType, company, desc, status string
	}{
		{"Facebook Ads", "Social Media", "Meta", "Facebook广告推广服务", "active"},
		{"Google Ads", "Search Engine", "Google", "Google搜索和展示广告", "active"},
		{"TikTok Ads", "Social Media", "ByteDance", "TikTok短视频广告投放", "active"},
		{"Instagram Ads", "Social Media", "Meta", "Instagram图片视频广告", "active"},
		{"YouTube Ads", "Video", "Google", "YouTube视频广告投放", "inactive"},
	}
	
	for _, p := range products {
		_, err = db.Exec(`
			INSERT INTO products (name, type, company, description, status, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE updated_at = NOW()`,
			p.name, p.productType, p.company, p.desc, p.status)
		if err != nil {
			fmt.Printf("插入产品 %s 失败: %v\n", p.name, err)
		}
	}
	fmt.Println("✅ 产品数据插入完成")
	
	// 4. 插入客户数据
	customers := []struct {
		name, email, phone, company, status string
		balance float64
	}{
		{"张三", "zhangsan@example.com", "13800138001", "科技有限公司", "active", 10000.00},
		{"李四", "lisi@example.com", "13800138002", "贸易有限公司", "active", 25000.00},
		{"王五", "wangwu@example.com", "13800138003", "电商有限公司", "active", 50000.00},
		{"赵六", "zhaoliu@example.com", "13800138004", "文化传媒公司", "inactive", 0},
		{"钱七", "qianqi@example.com", "13800138005", "互联网科技", "active", 15000.00},
	}
	
	for _, c := range customers {
		_, err = db.Exec(`
			INSERT INTO customers (name, email, phone, company, status, balance, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE updated_at = NOW()`,
			c.name, c.email, c.phone, c.company, c.status, c.balance)
		if err != nil {
			fmt.Printf("插入客户 %s 失败: %v\n", c.name, err)
		}
	}
	fmt.Println("✅ 客户数据插入完成")
	
	// 5. 插入营销计划
	campaigns := []struct {
		name string
		productID int
		desc, status string
		budget float64
	}{
		{"双11大促销推广", 1, "双11期间的Facebook广告推广计划", "running", 100000.00},
		{"新品发布推广", 2, "Google搜索广告推广新产品", "running", 50000.00},
		{"品牌宣传计划", 3, "TikTok品牌曝光活动", "paused", 75000.00},
		{"节日营销活动", 1, "圣诞节Facebook推广", "draft", 30000.00},
		{"Q4季度推广", 2, "第四季度Google广告投放", "completed", 80000.00},
	}
	
	for _, c := range campaigns {
		_, err = db.Exec(`
			INSERT INTO campaigns (name, product_id, description, status, budget, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE updated_at = NOW()`,
			c.name, c.productID, c.desc, c.status, c.budget)
		if err != nil {
			fmt.Printf("插入计划 %s 失败: %v\n", c.name, err)
		}
	}
	fmt.Println("✅ 营销计划数据插入完成")
	
	// 6. 插入交易记录
	transactions := []struct {
		userID int
		transType int  // 1=充值, 2=提现, 3=消费
		amount float64
		status int  // 1=待处理, 2=成功
		desc string
	}{
		{1, 1, 10000.00, 2, "账户充值"},
		{2, 1, 25000.00, 2, "账户充值"},
		{3, 1, 50000.00, 2, "账户充值"},
		{1, 3, 5000.00, 2, "广告消费"},
		{2, 3, 8000.00, 2, "广告消费"},
		{3, 2, 10000.00, 1, "提现申请"},
	}
	
	for i, t := range transactions {
		orderPrefix := ""
		switch t.transType {
		case 1:
			orderPrefix = "R"
		case 2:
			orderPrefix = "W"
		case 3:
			orderPrefix = "C"
		}
		orderNo := fmt.Sprintf("%s%s%03d", orderPrefix, time.Now().Format("20060102150405"), i+1)
		
		_, err = db.Exec(`
			INSERT INTO transactions (user_id, type, amount, status, description, order_no, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			t.userID, t.transType, t.amount, t.status, t.desc, orderNo)
		if err != nil {
			fmt.Printf("插入交易记录失败: %v\n", err)
		}
	}
	fmt.Println("✅ 交易记录插入完成")
	
	// 7. 插入优惠券数据
	coupons := []struct {
		code, name, couponType string
		value, minAmount, maxDiscount float64
		validDays, totalCount, usedCount int
		status string
	}{
		{"WELCOME2025", "新用户优惠券", "discount", 10.00, 100.00, 50.00, 30, 1000, 50, "active"},
		{"DOUBLE11", "双11特惠券", "discount", 20.00, 500.00, 200.00, 7, 500, 120, "active"},
		{"VIP100", "VIP专享券", "amount", 100.00, 1000.00, 100.00, 90, 100, 10, "active"},
		{"NEWYEAR50", "新年优惠券", "amount", 50.00, 200.00, 50.00, 15, 2000, 0, "inactive"},
	}
	
	for _, c := range coupons {
		_, err = db.Exec(`
			INSERT INTO coupons (code, name, type, value, min_amount, max_discount, valid_days, total_count, used_count, status, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE updated_at = NOW()`,
			c.code, c.name, c.couponType, c.value, c.minAmount, c.maxDiscount, c.validDays, c.totalCount, c.usedCount, c.status)
		if err != nil {
			fmt.Printf("插入优惠券 %s 失败: %v\n", c.code, err)
		}
	}
	fmt.Println("✅ 优惠券数据插入完成")
	
	// 8. 分配角色关系
	// 获取root的ID和超级管理员角色ID
	var rootID, adminID int
	var superAdminRoleID, adminRoleID int
	
	db.QueryRow("SELECT id FROM admins WHERE username = 'root'").Scan(&rootID)
	db.QueryRow("SELECT id FROM admins WHERE username = 'admin'").Scan(&adminID)
	db.QueryRow("SELECT id FROM roles WHERE code = 'super_admin'").Scan(&superAdminRoleID)
	db.QueryRow("SELECT id FROM roles WHERE code = 'admin'").Scan(&adminRoleID)
	
	if rootID > 0 && superAdminRoleID > 0 {
		db.Exec("DELETE FROM admin_roles WHERE admin_id = ?", rootID)
		db.Exec("INSERT INTO admin_roles (admin_id, role_id) VALUES (?, ?)", rootID, superAdminRoleID)
		fmt.Println("✅ root已分配超级管理员角色")
	}
	
	if adminID > 0 && adminRoleID > 0 {
		db.Exec("DELETE FROM admin_roles WHERE admin_id = ?", adminID)
		db.Exec("INSERT INTO admin_roles (admin_id, role_id) VALUES (?, ?)", adminID, adminRoleID)
		fmt.Println("✅ admin已分配管理员角色")
	}
	
	// 显示统计信息
	fmt.Println("\n=== 数据统计 ===")
	var count int
	
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	fmt.Printf("产品总数: %d\n", count)
	
	db.QueryRow("SELECT COUNT(*) FROM customers").Scan(&count)
	fmt.Printf("客户总数: %d\n", count)
	
	db.QueryRow("SELECT COUNT(*) FROM campaigns").Scan(&count)
	fmt.Printf("营销计划总数: %d\n", count)
	
	db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&count)
	fmt.Printf("交易记录总数: %d\n", count)
	
	db.QueryRow("SELECT COUNT(*) FROM coupons").Scan(&count)
	fmt.Printf("优惠券总数: %d\n", count)
	
	fmt.Println("\n=== 账户信息 ===")
	fmt.Println("root: 超级管理员 (密码: root123)")
	fmt.Println("admin: 管理员 (密码: admin123)")
	
	fmt.Println("\n✅ 所有设置完成！")
}