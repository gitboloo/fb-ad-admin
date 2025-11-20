# API 路由规范

## 路由前缀规范

### 1. 管理后台 API: `/api/admin/*`
用于后台管理系统，管理员使用

**已迁移到新路由系统 (router/router.go):**
- ✅ `/api/admin/auth/login` - 管理员登录（公开）
- ✅ `/api/admin/auth/me` - 获取管理员信息
- ✅ `/api/admin/auth/logout` - 管理员退出
- ✅ `/api/admin/auth/permissions` - 获取管理员权限
- ✅ `/api/admin/auth/menus` - 获取管理员菜单
- ✅ `/api/admin/auth/password` - 修改管理员密码
- ✅ `/api/admin/agents/*` - 代理商管理
- ✅ `/api/admin/roles/*` - 角色管理
- ✅ `/api/admin/dashboard/*` - 仪表盘数据

**待迁移 (当前在 api/routes.go):**
- ⏳ `/api/products/*` → `/api/admin/products/*` - 产品管理
- ⏳ `/api/campaigns/*` → `/api/admin/campaigns/*` - 广告计划管理
- ⏳ `/api/customers/*` → `/api/admin/customers/*` - 客户管理
- ⏳ `/api/finance/*` → `/api/admin/finance/*` - 财务管理
- ⏳ `/api/coupons/*` → `/api/admin/coupons/*` - 优惠券管理
- ⏳ `/api/authcodes/*` → `/api/admin/authcodes/*` - 授权码管理
- ⏳ `/api/permissions/*` → `/api/admin/permissions/*` - 权限管理
- ⏳ `/api/statistics/*` → `/api/admin/statistics/*` - 统计分析
- ⏳ `/api/admins/*` → `/api/admin/admins/*` - 管理员账户管理
- ⏳ `/api/system/*` → `/api/admin/system/*` - 系统管理

### 2. 客户端 API: `/api/cli/*`
用于客户端应用，普通用户使用

**待实现:**
- ⏳ `/api/cli/auth/login` - 客户登录
- ⏳ `/api/cli/auth/register` - 客户注册
- ⏳ `/api/cli/auth/me` - 获取客户信息
- ⏳ `/api/cli/auth/logout` - 客户退出
- ⏳ `/api/cli/profile` - 客户个人资料
- ⏳ `/api/cli/finance/balance` - 查询余额
- ⏳ `/api/cli/finance/recharge` - 充值
- ⏳ `/api/cli/finance/withdraw` - 提现
- ⏳ `/api/cli/finance/transactions` - 交易记录
- ⏳ `/api/cli/coupons` - 我的优惠券
- ⏳ `/api/cli/products` - 浏览产品
- ⏳ `/api/cli/campaigns` - 我的广告计划
- ⏳ `/api/cli/authcodes/verify` - 验证授权码

## 迁移计划

### Phase 1: 核心功能 ✅
- [x] 认证系统 (auth)
- [x] 代理商管理 (agents)
- [x] 角色权限管理 (roles)
- [x] 仪表盘 (dashboard)

### Phase 2: 业务管理 (待迁移)
- [ ] 产品管理 (products)
- [ ] 广告计划管理 (campaigns)
- [ ] 客户管理 (customers)
- [ ] 优惠券管理 (coupons)
- [ ] 授权码管理 (authcodes)

### Phase 3: 系统功能 (待迁移)
- [ ] 财务管理 (finance)
- [ ] 权限管理 (permissions)
- [ ] 统计分析 (statistics)
- [ ] 系统管理 (system)

### Phase 4: 客户端 API (待实现)
- [ ] 客户端认证
- [ ] 客户端个人中心
- [ ] 客户端财务操作
- [ ] 客户端业务功能

## 路由文件结构

```
backend/
├── router/
│   ├── router.go          # 主路由配置
│   ├── admin.go           # 管理后台路由 (当前在 router.go 中)
│   └── client.go          # 客户端路由 (待创建)
├── controllers/
│   ├── admin/             # 管理后台控制器
│   │   ├── auth.go        ✅
│   │   ├── agent.go       ✅
│   │   ├── role.go        ✅
│   │   ├── dashboard.go   ✅
│   │   ├── product.go     (待创建)
│   │   ├── campaign.go    (待创建)
│   │   └── ...
│   └── client/            # 客户端控制器 (待创建)
│       ├── auth.go
│       ├── profile.go
│       └── ...
└── api/
    └── routes.go          # 旧路由系统 (待废弃)
```

## 认证中间件

- **管理后台**: `middleware.AdminAuthMiddleware()`
- **客户端**: `middleware.ClientAuthMiddleware()` (待实现)

## 注意事项

1. **向后兼容**: 旧路由暂时保留，逐步迁移
2. **前端同步**: 每次后端路由变更需同步更新前端 API 调用
3. **文档更新**: 迁移完成后更新 API 文档
4. **测试**: 每个迁移的模块需进行完整测试
