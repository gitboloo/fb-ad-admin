-- =====================================
-- 广告推广系统数据库初始化脚本
-- =====================================
-- 角色系统: role=1(超级管理员) role=2(管理员) role=3(普通用户)
-- 默认账户: root(密码:root123) admin(密码:admin123)
-- =====================================

USE ad_platform;

-- =====================================
-- 1. 创建账户
-- =====================================

-- 禁用外键检查（确保ID更新顺利）
SET FOREIGN_KEY_CHECKS = 0;

-- 先清理可能的冲突数据
DELETE FROM admin_roles WHERE admin_id IN (1, 2, 999);

-- 创建root超级管理员（ID=1, role=1, 密码: root123）
-- role: 1=超级管理员, 2=管理员, 3=普通用户
INSERT INTO admins (id, username, account, password, role, status, created_at, updated_at) 
VALUES (1, 'root', 'root@platform.com', '$2a$10$MdOZblPqvtX7gLqHfmadWekxAzHOOaBQ9KO0elrGwN30UttjyWx6q', 1, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE id = 1, role = 1, status = 1, updated_at = NOW();

-- 创建admin管理员账户（ID=2, role=2, 密码: admin123）  
INSERT INTO admins (id, username, account, password, role, status, created_at, updated_at)
VALUES (2, 'admin', 'admin@platform.com', '$2a$10$3JgDDQHH.N8jOCR8mx5lOujOqIIHQY8vQ0lm6hYfGZTaKeFXHT/Iy', 2, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE id = 2, role = 2, status = 1, updated_at = NOW();

-- 重新启用外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- =====================================
-- 2. 产品数据
-- =====================================
-- type: 1=App, 2=Game, 3=Web, 4=Other
-- status: 0=Inactive, 1=Active, 2=Suspended

INSERT INTO products (name, type, company, description, status, logo, created_at, updated_at) VALUES
('Facebook广告SDK', 1, 'Meta', 'Facebook广告推广服务SDK，支持多平台投放', 1, '/uploads/products/facebook.png', NOW(), NOW()),
('Google Ads平台', 3, 'Google', 'Google搜索和展示广告平台，覆盖全球用户', 1, '/uploads/products/google.png', NOW(), NOW()),
('TikTok推广游戏', 2, 'ByteDance', 'TikTok游戏推广服务，精准触达年轻用户', 1, '/uploads/products/tiktok.png', NOW(), NOW()),
('Instagram营销工具', 1, 'Meta', 'Instagram营销推广应用，视觉营销首选', 1, '/uploads/products/instagram.png', NOW(), NOW()),
('YouTube广告系统', 3, 'Google', 'YouTube视频广告投放系统，视频营销平台', 1, '/uploads/products/youtube.png', NOW(), NOW()),
('微信小程序推广', 4, 'Tencent', '微信小程序推广服务，触达10亿用户', 1, '/uploads/products/wechat.png', NOW(), NOW()),
('抖音直播推广', 4, 'ByteDance', '抖音直播间推广工具，提升直播间人气', 1, '/uploads/products/douyin.png', NOW(), NOW()),
('Twitter广告平台', 3, 'Twitter Inc', 'Twitter社交媒体广告，实时营销利器', 1, '/uploads/products/twitter.png', NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- =====================================
-- 3. 客户数据
-- =====================================
-- status: 0=Inactive, 1=Active, 2=Blocked

INSERT INTO customers (name, email, phone, company, status, balance, created_at, updated_at) VALUES
('张三', 'zhangsan@example.com', '13800138001', '北京科技有限公司', 1, 10000.00, NOW(), NOW()),
('李四', 'lisi@example.com', '13800138002', '上海贸易有限公司', 1, 25000.00, NOW(), NOW()),
('王五', 'wangwu@example.com', '13800138003', '深圳电商有限公司', 1, 50000.00, NOW(), NOW()),
('赵六', 'zhaoliu@example.com', '13800138004', '广州文化传媒', 0, 0, NOW(), NOW()),
('钱七', 'qianqi@example.com', '13800138005', '杭州互联网科技', 1, 15000.00, NOW(), NOW()),
('孙八', 'sunba@example.com', '13800138006', '成都游戏开发', 1, 30000.00, NOW(), NOW()),
('周九', 'zhoujiu@example.com', '13800138007', '南京教育科技', 1, 8000.00, NOW(), NOW()),
('吴十', 'wushi@example.com', '13800138008', '武汉医疗健康', 2, 0, NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- =====================================
-- 4. 营销计划数据（使用JSON字段）
-- =====================================
-- status: 0=Inactive, 1=Active, 2=Paused, 3=Ended
-- 注意：没有budget字段，预算信息在delivery_rules的JSON中

INSERT INTO campaigns (name, product_id, description, status, logo, delivery_content, delivery_rules, user_targeting, created_at, updated_at) VALUES
(
    '双11大促销推广',
    1,
    '双11期间的Facebook广告推广计划，提升品牌曝光和销售转化',
    1,
    '/uploads/campaigns/double11.png',
    '{"title":"双11狂欢节","description":"全场5折起","images":["/uploads/ads/banner1.jpg"],"videos":[],"call_to_action":"立即购买"}',
    '{"start_date":"2025-11-01T00:00:00Z","end_date":"2025-11-30T23:59:59Z","daily_budget":3000,"total_budget":100000,"bid_amount":2.5,"frequency_cap":3,"delivery_pacing":"standard"}',
    '{"age_range":[18,45],"genders":["all"],"countries":["CN"],"languages":["zh"],"interests":["shopping","fashion"],"behaviors":["online_shopping"],"device_types":["mobile","desktop"],"operating_systems":["ios","android"],"custom_audiences":[]}',
    NOW(),
    NOW()
),
(
    '新品发布推广',
    2,
    'Google搜索广告推广新产品，提高品牌知名度',
    1,
    '/uploads/campaigns/newproduct.png',
    '{"title":"新品上市","description":"限时优惠","images":["/uploads/ads/new.jpg"],"videos":[],"call_to_action":"了解更多"}',
    '{"start_date":"2025-09-01T00:00:00Z","end_date":"2025-09-30T23:59:59Z","daily_budget":2000,"total_budget":50000,"bid_amount":1.8,"frequency_cap":5,"delivery_pacing":"accelerated"}',
    '{"age_range":[20,40],"genders":["all"],"countries":["CN","US"],"languages":["zh","en"],"interests":["technology"],"behaviors":[],"device_types":["mobile"],"operating_systems":["android"],"custom_audiences":[]}',
    NOW(),
    NOW()
),
(
    '品牌宣传计划',
    3,
    'TikTok品牌曝光活动，提升品牌影响力',
    2,
    '/uploads/campaigns/brand.png',
    '{"title":"品牌故事","description":"与你同行","images":[],"videos":["/uploads/ads/brand_video.mp4"],"call_to_action":"关注我们"}',
    '{"start_date":"2025-08-01T00:00:00Z","end_date":"2025-10-31T23:59:59Z","daily_budget":2500,"total_budget":75000,"bid_amount":3.0,"frequency_cap":10,"delivery_pacing":"standard"}',
    '{"age_range":[16,35],"genders":["all"],"countries":["CN"],"languages":["zh"],"interests":["entertainment","music"],"behaviors":["video_watcher"],"device_types":["mobile"],"operating_systems":["ios","android"],"custom_audiences":[]}',
    NOW(),
    NOW()
),
(
    '春节营销活动',
    6,
    '春节期间微信推广，送祝福赢红包',
    0,
    '/uploads/campaigns/spring.png',
    '{"title":"新春祝福","description":"红包雨来袭","images":["/uploads/ads/spring.jpg"],"videos":[],"call_to_action":"抢红包"}',
    '{"start_date":"2026-01-25T00:00:00Z","end_date":"2026-02-10T23:59:59Z","daily_budget":5000,"total_budget":80000,"bid_amount":1.5,"frequency_cap":20,"delivery_pacing":"standard"}',
    '{"age_range":[18,60],"genders":["all"],"countries":["CN"],"languages":["zh"],"interests":["festival","family"],"behaviors":[],"device_types":["mobile"],"operating_systems":["ios","android"],"custom_audiences":[]}',
    NOW(),
    NOW()
)
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- =====================================
-- 5. 交易记录
-- =====================================
-- type: 1=充值, 2=提现, 3=消费, 4=退款, 5=奖励
-- status: 1=待处理, 2=成功, 3=失败, 4=已取消, 5=处理中

INSERT INTO transactions (user_id, type, amount, status, description, order_no, payment_method, balance_before, balance_after, created_at, updated_at) VALUES
(1, 1, 10000.00, 2, '账户充值', CONCAT('R', DATE_FORMAT(NOW(), '%Y%m%d%H%i%s'), '001'), 'alipay', 0, 10000.00, NOW(), NOW()),
(2, 1, 25000.00, 2, '账户充值', CONCAT('R', DATE_FORMAT(NOW(), '%Y%m%d%H%i%s'), '002'), 'wechat', 0, 25000.00, NOW(), NOW()),
(3, 1, 50000.00, 2, '账户充值', CONCAT('R', DATE_FORMAT(NOW(), '%Y%m%d%H%i%s'), '003'), 'bank_transfer', 0, 50000.00, NOW(), NOW()),
(1, 3, 5000.00, 2, 'Facebook广告消费', CONCAT('C', DATE_FORMAT(NOW(), '%Y%m%d%H%i%s'), '004'), '', 10000.00, 5000.00, NOW(), NOW()),
(2, 3, 8000.00, 2, 'Google广告消费', CONCAT('C', DATE_FORMAT(NOW(), '%Y%m%d%H%i%s'), '005'), '', 25000.00, 17000.00, NOW(), NOW()),
(3, 2, 10000.00, 1, '提现申请', CONCAT('W', DATE_FORMAT(NOW(), '%Y%m%d%H%i%s'), '006'), 'bank_transfer', 50000.00, 50000.00, NOW(), NOW()),
(5, 1, 15000.00, 2, '账户充值', CONCAT('R', DATE_FORMAT(NOW(), '%Y%m%d%H%i%s'), '007'), 'alipay', 0, 15000.00, NOW(), NOW()),
(6, 1, 30000.00, 2, '账户充值', CONCAT('R', DATE_FORMAT(NOW(), '%Y%m%d%H%i%s'), '008'), 'wechat', 0, 30000.00, NOW(), NOW()),
(1, 5, 1000.00, 2, '推荐新用户奖励', CONCAT('B', DATE_FORMAT(NOW(), '%Y%m%d%H%i%s'), '009'), '', 5000.00, 6000.00, NOW(), NOW()),
(2, 4, 2000.00, 2, '广告费退款', CONCAT('F', DATE_FORMAT(NOW(), '%Y%m%d%H%i%s'), '010'), '', 17000.00, 19000.00, NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- =====================================
-- 6. 角色关联
-- =====================================

-- 给root分配超级管理员角色
INSERT INTO admin_roles (admin_id, role_id) 
SELECT a.id, r.id FROM admins a, roles r 
WHERE a.username = 'root' AND r.code = 'super_admin'
ON DUPLICATE KEY UPDATE admin_id = admin_id;

-- 给admin分配管理员角色
DELETE FROM admin_roles WHERE admin_id = (SELECT id FROM admins WHERE username = 'admin');
INSERT INTO admin_roles (admin_id, role_id) 
SELECT a.id, r.id FROM admins a, roles r 
WHERE a.username = 'admin' AND r.code = 'admin';

-- =====================================
-- 7. 显示结果
-- =====================================

SELECT '========== 数据导入完成 ==========' as '信息';

SELECT '账户信息:' as '信息';
SELECT username as '用户名', 
       CASE role 
         WHEN 1 THEN '超级管理员'
         WHEN 2 THEN '管理员'
         WHEN 3 THEN '普通用户'
       END as '角色',
       '密码' as '密码说明'
FROM admins WHERE username IN ('root', 'admin')
UNION ALL
SELECT 'root', '', 'root123'
UNION ALL  
SELECT 'admin', '', 'admin123';

SELECT '' as '';
SELECT '数据统计:' as '信息';
SELECT 
  CONCAT('产品: ', COUNT(*), ' 条') as '统计'
FROM products
UNION ALL
SELECT CONCAT('客户: ', COUNT(*), ' 条') FROM customers
UNION ALL
SELECT CONCAT('计划: ', COUNT(*), ' 条') FROM campaigns
UNION ALL
SELECT CONCAT('交易: ', COUNT(*), ' 条') FROM transactions;