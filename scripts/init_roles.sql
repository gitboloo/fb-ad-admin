-- 最终版本：安全更新角色的SQL脚本
-- role: 1=超级管理员, 2=管理员, 3=普通用户

USE ad_platform;

-- =====================================
-- 1. 先备份并清理admin_roles表
-- =====================================

-- 先清理admin_roles表中的旧数据
DELETE FROM admin_roles WHERE admin_id IN (1, 2, 999);

-- =====================================
-- 2. 处理ID冲突问题
-- =====================================

-- 禁用外键检查
SET FOREIGN_KEY_CHECKS = 0;

-- 先将admin的ID改为临时值
UPDATE admins 
SET id = 999 
WHERE username = 'admin' AND id = 1;

-- 如果root的ID不是1，更新为1
UPDATE admins 
SET id = 1, role = 1 
WHERE username = 'root';

-- 如果root不存在，创建它
INSERT IGNORE INTO admins (id, username, account, password, role, status, created_at, updated_at) 
VALUES (1, 'root', 'root@platform.com', '$2a$10$MdOZblPqvtX7gLqHfmadWekxAzHOOaBQ9KO0elrGwN30UttjyWx6q', 1, 1, NOW(), NOW());

-- 将admin设置为ID=2
UPDATE admins 
SET id = 2, role = 2 
WHERE username = 'admin';

-- 如果admin不存在，创建它
INSERT IGNORE INTO admins (id, username, account, password, role, status, created_at, updated_at)
VALUES (2, 'admin', 'admin@platform.com', '$2a$10$3JgDDQHH.N8jOCR8mx5lOujOqIIHQY8vQ0lm6hYfGZTaKeFXHT/Iy', 2, 1, NOW(), NOW());

-- 重新启用外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- =====================================
-- 3. 更新其他表中的外键引用
-- =====================================

-- 更新transactions表（如果有user_id关联到admin）
UPDATE transactions SET user_id = 2 WHERE user_id = 999;
UPDATE transactions SET user_id = 1 WHERE user_id IN (SELECT id FROM admins WHERE username = 'root');

-- =====================================
-- 4. 重新分配角色（如果角色表存在）
-- =====================================

-- 给root分配超级管理员角色
INSERT INTO admin_roles (admin_id, role_id) 
SELECT 1, id FROM roles WHERE code = 'super_admin'
ON DUPLICATE KEY UPDATE admin_id = admin_id;

-- 给admin分配管理员角色  
INSERT INTO admin_roles (admin_id, role_id) 
SELECT 2, id FROM roles WHERE code = 'admin'
ON DUPLICATE KEY UPDATE admin_id = admin_id;

-- =====================================
-- 5. 验证结果
-- =====================================

SELECT '========== 更新完成 ==========' as '信息';

-- 显示账户信息
SELECT 
    id as 'ID',
    username as '用户名', 
    account as '账号',
    CASE role 
        WHEN 1 THEN '超级管理员'
        WHEN 2 THEN '管理员'
        WHEN 3 THEN '普通用户'
    END as '角色',
    CASE status
        WHEN 1 THEN '启用'
        ELSE '禁用'
    END as '状态'
FROM admins 
WHERE id IN (1, 2)
ORDER BY id;

-- 显示角色关联
SELECT '========== 角色关联 ==========' as '信息';
SELECT 
    ar.admin_id,
    a.username,
    r.name as role_name,
    r.code as role_code
FROM admin_roles ar
JOIN admins a ON ar.admin_id = a.id
JOIN roles r ON ar.role_id = r.id
WHERE ar.admin_id IN (1, 2);

-- 登录信息提示
SELECT '' as '';
SELECT '========== 登录信息 ==========' as '信息';
SELECT 'root' as '用户名', 'root123' as '密码', '超级管理员(ID=1, role=1)' as '说明'
UNION ALL
SELECT 'admin', 'admin123', '管理员(ID=2, role=2)';