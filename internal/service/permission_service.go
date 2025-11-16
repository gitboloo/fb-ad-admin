package service

import (
	"errors"
	"strconv"

	"github.com/ad-platform/backend/internal/database"
	"github.com/ad-platform/backend/internal/models"
	"github.com/ad-platform/backend/internal/repository"
)

type PermissionService struct {
	adminRepo      *repository.AdminRepository
	permissionRepo *repository.PermissionRepository
	roleRepo       *repository.RoleRepository
}

func NewPermissionService() *PermissionService {
	return &PermissionService{
		adminRepo:      repository.NewAdminRepository(),
		permissionRepo: repository.NewPermissionRepository(),
		roleRepo:       repository.NewRoleRepository(),
	}
}

// GetUserPermissions 获取用户权限列表
func (s *PermissionService) GetUserPermissions(adminID uint) ([]string, error) {
	db := database.GetDB()
	
	// 获取用户的所有权限
	permissions, err := models.GetPermissionsByAdminID(db, adminID)
	if err != nil {
		return nil, err
	}

	// 转换为权限代码列表
	var permCodes []string
	for _, perm := range permissions {
		if perm.Status == 1 {
			permCodes = append(permCodes, perm.Code)
		}
	}

	return permCodes, nil
}

// GetUserMenuTree 获取用户菜单树
func (s *PermissionService) GetUserMenuTree(adminID uint) ([]models.Permission, error) {
	db := database.GetDB()
	
	// 获取用户的所有权限
	permissions, err := models.GetPermissionsByAdminID(db, adminID)
	if err != nil {
		return nil, err
	}

	// 过滤出菜单类型的权限
	var menus []models.Permission
	for _, perm := range permissions {
		if perm.Type == models.PermissionTypeMenu && perm.Status == 1 {
			menus = append(menus, perm)
		}
	}

	// 构建菜单树
	return s.buildMenuTree(menus), nil
}

// buildMenuTree 构建菜单树
func (s *PermissionService) buildMenuTree(permissions []models.Permission) []models.Permission {
	// 创建映射
	permMap := make(map[uint]*models.Permission)
	for i := range permissions {
		permMap[permissions[i].ID] = &permissions[i]
	}

	// 构建树形结构
	var tree []models.Permission
	for i := range permissions {
		if permissions[i].ParentID == 0 {
			tree = append(tree, permissions[i])
		} else {
			if parent, exists := permMap[permissions[i].ParentID]; exists {
				parent.Children = append(parent.Children, permissions[i])
			}
		}
	}

	return tree
}

// GetAllPermissions 获取所有权限
func (s *PermissionService) GetAllPermissions() ([]models.Permission, error) {
	return s.permissionRepo.GetAll()
}

// CreatePermission 创建权限
func (s *PermissionService) CreatePermission(permission *models.Permission) (*models.Permission, error) {
	// 检查权限代码是否已存在
	exists, err := s.permissionRepo.ExistsByCode(permission.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("权限代码已存在")
	}

	err = s.permissionRepo.Create(permission)
	if err != nil {
		return nil, err
	}

	return permission, nil
}

// UpdatePermission 更新权限
func (s *PermissionService) UpdatePermission(id string, permission *models.Permission) error {
	permID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return errors.New("无效的权限ID")
	}

	permission.ID = uint(permID)
	return s.permissionRepo.Update(permission)
}

// DeletePermission 删除权限
func (s *PermissionService) DeletePermission(id string) error {
	permID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return errors.New("无效的权限ID")
	}

	return s.permissionRepo.Delete(uint(permID))
}

// GetAllRoles 获取所有角色
func (s *PermissionService) GetAllRoles() ([]models.Role, error) {
	return s.roleRepo.GetAll()
}

// CreateRole 创建角色
func (s *PermissionService) CreateRole(role *models.Role) (*models.Role, error) {
	// 检查角色代码是否已存在
	exists, err := s.roleRepo.ExistsByCode(role.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("角色代码已存在")
	}

	err = s.roleRepo.Create(role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// UpdateRole 更新角色
func (s *PermissionService) UpdateRole(id string, role *models.Role) error {
	roleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return errors.New("无效的角色ID")
	}

	role.ID = uint(roleID)
	return s.roleRepo.Update(role)
}

// AssignPermissionsToRole 给角色分配权限
func (s *PermissionService) AssignPermissionsToRole(roleID string, permissionIDs []uint) error {
	id, err := strconv.ParseUint(roleID, 10, 32)
	if err != nil {
		return errors.New("无效的角色ID")
	}

	return s.roleRepo.AssignPermissions(uint(id), permissionIDs)
}

// AssignRolesToUser 给用户分配角色
func (s *PermissionService) AssignRolesToUser(userID string, roleIDs []uint) error {
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	return s.adminRepo.AssignRoles(uint(id), roleIDs)
}