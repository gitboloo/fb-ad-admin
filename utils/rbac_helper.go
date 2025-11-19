package utils

import (
	"backend/database"
)

// GetAdminMenus 获取管理员的菜单列表
func GetAdminMenus(adminID uint) ([]map[string]interface{}, error) {
	db := database.GetDB()

	// 查询管理员的角色ID
	var roleIDs []uint
	err := db.Table("admin_roles").
		Where("admin_id = ?", adminID).
		Pluck("role_id", &roleIDs).Error
	if err != nil {
		return nil, err
	}

	if len(roleIDs) == 0 {
		return []map[string]interface{}{}, nil
	}

	// 查询菜单
	var menus []map[string]interface{}
	err = db.Raw(`
		SELECT DISTINCT m.*
		FROM menus m
		INNER JOIN role_menus rm ON rm.menu_id = m.id
		WHERE rm.role_id IN ? AND m.status = 1
		ORDER BY m.sort ASC, m.id ASC
	`, roleIDs).Scan(&menus).Error

	if err != nil {
		return nil, err
	}

	return menus, nil
}

// BuildMenuTree 构建菜单树形结构
func BuildMenuTree(menus []map[string]interface{}) []map[string]interface{} {
	// 创建映射
	menuMap := make(map[uint]map[string]interface{})
	var roots []map[string]interface{}

	// 初始化children
	for i := range menus {
		var id uint
		switch v := menus[i]["id"].(type) {
		case int64:
			id = uint(v)
		case uint64:
			id = uint(v)
		case int:
			id = uint(v)
		case uint:
			id = v
		case float64:
			id = uint(v)
		}

		menuMap[id] = menus[i]
		menus[i]["children"] = []map[string]interface{}{}
	}

	// 构建树形结构
	for _, menu := range menus {
		var parentID uint
		switch v := menu["parent_id"].(type) {
		case int64:
			parentID = uint(v)
		case uint64:
			parentID = uint(v)
		case int:
			parentID = uint(v)
		case uint:
			parentID = v
		case float64:
			parentID = uint(v)
		case nil:
			parentID = 0
		}

		if parentID == 0 {
			roots = append(roots, menu)
		} else {
			if parent, ok := menuMap[parentID]; ok {
				children := parent["children"].([]map[string]interface{})
				parent["children"] = append(children, menu)
			}
		}
	}

	return roots
}

// GetAdminRoles 获取管理员的角色列表
func GetAdminRoles(adminID uint) ([]map[string]interface{}, error) {
	db := database.GetDB()

	var roles []map[string]interface{}
	err := db.Raw(`
		SELECT r.*
		FROM roles r
		INNER JOIN admin_roles ar ON ar.role_id = r.id
		WHERE ar.admin_id = ? AND r.status = 1
	`, adminID).Scan(&roles).Error

	if err != nil {
		return nil, err
	}

	return roles, nil
}

// GetAdminPermissions 获取管理员的权限列表
func GetAdminPermissions(adminID uint) ([]map[string]interface{}, error) {
	db := database.GetDB()

	// 查询管理员的角色ID
	var roleIDs []uint
	err := db.Table("admin_roles").
		Where("admin_id = ?", adminID).
		Pluck("role_id", &roleIDs).Error
	if err != nil {
		return nil, err
	}

	if len(roleIDs) == 0 {
		return []map[string]interface{}{}, nil
	}

	// 查询权限
	var permissions []map[string]interface{}
	err = db.Raw(`
		SELECT DISTINCT p.*
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id IN ? AND p.status = 1
	`, roleIDs).Scan(&permissions).Error

	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// HasPermission 检查管理员是否有指定权限
func HasPermission(adminID uint, permissionName string) bool {
	db := database.GetDB()

	var count int64
	db.Raw(`
		SELECT COUNT(DISTINCT p.id)
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		INNER JOIN admin_roles ar ON ar.role_id = rp.role_id
		WHERE ar.admin_id = ? AND p.name = ? AND p.status = 1
	`, adminID, permissionName).Scan(&count)

	return count > 0
}

// GetAdminDataScope 获取管理员的数据权限范围
func GetAdminDataScope(adminID uint) int8 {
	db := database.GetDB()

	var minDataScope int8 = 4 // 默认仅本人

	// 获取管理员所有角色中的最小data_scope(数值越小权限越大)
	db.Raw(`
		SELECT MIN(r.data_scope)
		FROM roles r
		INNER JOIN admin_roles ar ON ar.role_id = r.id
		WHERE ar.admin_id = ? AND r.status = 1
	`, adminID).Scan(&minDataScope)

	return minDataScope
}
