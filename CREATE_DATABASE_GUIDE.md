# 创建数据库指南

## 最简单方法：使用phpStudy界面

### 步骤1：打开phpStudy数据库管理
![](1) 打开 **phpStudy Pro** 软件
![](2) 在左侧菜单找到 **数据库**
![](3) 点击 **创建数据库** 按钮

### 步骤2：填写数据库信息
- **数据库名称**: `ad_platform`
- **字符集**: `utf8mb4`
- **排序规则**: `utf8mb4_unicode_ci`

### 步骤3：点击创建
点击 **确定** 或 **创建** 按钮

---

## 方法2：使用phpMyAdmin（网页界面）

### 步骤1：打开phpMyAdmin
在浏览器中访问：http://localhost/phpmyadmin/

### 步骤2：登录
- **用户名**: `root`
- **密码**: `root` 或 `123456` 或 留空

### 步骤3：创建数据库
1. 点击左上角的 **新建** 或 **New**
2. 输入数据库名：`ad_platform`
3. 选择字符集：`utf8mb4_unicode_ci`
4. 点击 **创建** 或 **Create**

---

## 方法3：使用Navicat或其他数据库工具

如果您安装了Navicat、HeidiSQL或DBeaver等工具：

1. 连接到localhost的MySQL
   - 主机: `localhost` 或 `127.0.0.1`
   - 端口: `3306`
   - 用户: `root`
   - 密码: `root` 或 `123456`

2. 右键点击连接，选择"新建数据库"

3. 输入数据库名：`ad_platform`
   字符集：`utf8mb4`

---

## 数据库创建成功后

运行以下命令启动服务器：

```bash
cd D:\phpstudy_pro\WWW\ad\ad-platform\backend
start_server.bat
```

或者直接运行：

```bash
go run cmd\main.go
```

---

## 常见问题

### Q: 不知道MySQL密码？
A: phpStudy默认密码通常是：
- `root` (最常见)
- `123456`
- 空密码（不输入）

### Q: 找不到数据库管理入口？
A: 在phpStudy主界面：
- 查找"数据库"标签
- 或者找"MySQL管理器"
- 或者点击"软件管理"找到phpMyAdmin并启动

### Q: 创建失败？
A: 检查MySQL服务是否启动：
1. 在phpStudy中查看MySQL状态
2. 如果是红色，点击启动
3. 等待变成绿色后再创建数据库