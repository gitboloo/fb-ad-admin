# 广告平台后端 (Ad Platform Backend)

基于Gin框架的Go后端项目，提供广告平台的完整后端服务。

## 项目特性

- **现代架构**: 采用分层架构设计（Controller-Service-Repository）
- **完整功能**: 包含管理员、产品、广告计划、优惠券、客户等完整业务模块
- **安全认证**: JWT认证 + 角色权限控制
- **数据库**: MySQL + Redis缓存
- **中间件**: CORS、限流、日志、错误恢复等完整中间件
- **配置管理**: 支持环境变量和配置文件
- **生产就绪**: 包含日志记录、错误处理、优雅关闭等生产级特性

## 项目结构

```
backend/
├── api/                   # API路由定义和控制器
├── bin/                   # 编译输出目录
│   └── server.exe
├── cmd/                   # 应用入口
│   └── main.go
├── configs/               # 配置管理
│   └── config.go
├── controllers/           # 控制器层(MVC架构)
│   ├── admin/            # 管理员控制器
│   └── agent/            # 代理商控制器
├── database/              # 数据库连接
│   ├── mysql.go
│   └── redis.go
├── middleware/            # 中间件
│   ├── auth.go
│   ├── cors.go
│   ├── logger.go
│   └── rate_limit.go
├── models/                # 数据模型(ORM)
│   ├── admin.go
│   ├── agent.go
│   ├── product.go
│   └── ...
├── repositories/          # 数据访问层(Repository)
│   ├── admin_repository.go
│   ├── agent_repository.go
│   └── ...
├── router/                # 路由注册
│   └── router.go
├── services/              # 业务逻辑层(Service)
│   ├── admin_service.go
│   ├── agent_service.go
│   └── ...
├── types/                 # 类型定义
├── utils/                 # 工具函数
├── .env                   # 环境变量配置
├── .env.example           # 环境变量示例
├── go.mod                 # Go模块文件
└── README.md              # 项目说明
```

## 数据模型

### 核心模型
- **Admin** (管理员): 系统管理员账户管理
- **Product** (产品): 广告产品信息管理
- **Campaign** (广告计划): 广告投放计划管理
- **Coupon** (优惠券): 优惠券系统
- **UserCoupon** (用户优惠券): 用户优惠券关联
- **AuthCode** (授权码): 授权码管理
- **Transaction** (交易): 交易记录管理
- **Customer** (客户): 客户信息管理
- **SystemConfig** (系统配置): 系统配置管理

## 快速开始

### 环境要求
- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 安装步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd ad-platform/backend
```

2. **安装依赖**
```bash
make deps
```

3. **配置环境**
```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库和Redis连接信息
```

4. **创建数据库**
```sql
CREATE DATABASE ad_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

5. **运行项目**
```bash
# 开发模式
make dev

# 或者构建后运行
make build
make run
```

### 使用Makefile

```bash
# 查看所有可用命令
make help

# 安装依赖
make deps

# 构建项目
make build

# 运行测试
make test

# 代码格式化
make fmt

# 运行开发服务器
make dev
```

## API文档

### 认证接口

#### 登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "account": "admin@example.com",
  "password": "admin123"
}
```

### 管理员接口

#### 获取个人信息
```http
GET /api/v1/profile
Authorization: Bearer <token>
```

#### 更新密码
```http
PUT /api/v1/profile/password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "old_password",
  "new_password": "new_password"
}
```

#### 管理员列表（需要管理员权限）
```http
GET /api/v1/admin/admins?page=1&page_size=20&status=1&role=2
Authorization: Bearer <token>
```

## 配置说明

### 环境变量
详细的配置选项请参考 `.env.example` 文件：

- **数据库配置**: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
- **Redis配置**: REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_DB
- **JWT配置**: JWT_SECRET, JWT_EXPIRE_HOURS
- **服务器配置**: SERVER_PORT, SERVER_HOST, ENV

## 部署

### Docker部署
```bash
# 构建Docker镜像
make docker-build

# 运行容器
docker run -d --name ad-platform-backend \
  -p 8080:8080 \
  -e DB_HOST=your_db_host \
  -e DB_PASSWORD=your_db_password \
  ad-platform-backend
```

### Linux部署
```bash
# 构建Linux二进制文件
make build-linux

# 上传到服务器并运行
./bin/ad-platform-backend_unix
```

## 开发指南

### 添加新模型
1. 在 `models/` 中创建模型文件
2. 在 `models/migrate.go` 中添加到 `AllModels()` 函数
3. 在 `repositories/` 中创建对应的 Repository
4. 在 `services/` 中创建对应的 Service
5. 在 `controllers/` 中添加 Controller
6. 在 `router/` 或 `api/` 中添加路由

### 中间件
项目包含以下中间件：
- **认证中间件**: JWT token验证
- **权限中间件**: 角色权限控制
- **CORS中间件**: 跨域请求支持
- **限流中间件**: API请求限制
- **日志中间件**: 请求日志记录
- **错误恢复中间件**: Panic恢复

### 数据库迁移
项目启动时会自动执行数据库迁移和默认数据插入。

默认管理员账户：
- 用户名: `admin`
- 密码: `admin123`
- 邮箱: `admin@example.com`

## 贡献指南

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证。

## 支持

如有问题或建议，请创建 Issue 或联系开发团队。