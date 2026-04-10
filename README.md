# DNS Manager

一个现代化的 DNS 记录管理系统，支持多个 DNS 服务提供商。

## 功能特性

- 🌐 **多提供商支持**：支持 Cloudflare、腾讯云 DNSPod、Dynu、NDJP NET、deSEC 等 DNS 服务提供商
- 🔐 **安全认证**：JWT 身份验证
- 🎨 **现代 UI**：Vercel 风格的简洁界面
- 🌓 **主题切换**：支持亮色/暗色/跟随系统三种模式
- 🌍 **多语言**：支持中文和英文
- 📱 **响应式设计**：适配各种屏幕尺寸
- 🐳 **Docker 支持**：一键部署
- 📊 **统计功能**：域名数量统计
- 🔍 **搜索过滤**：快速查找域名和记录
- 📝 **操作日志**：记录所有操作历史

## 技术栈

### 后端
- Go 1.24+
- Gin Web Framework
- SQLite 数据库
- JWT 认证

### 前端
- React 18
- React Router
- Vite
- Lucide Icons

## 快速开始

### 使用 Docker（推荐）

1. 克隆仓库：
```bash
git clone https://github.com/lizhixu/dns-mng.git
cd dns-mng
```

2. 启动服务：
```bash
chmod +x start.sh
./start.sh
```

3. 访问应用：
- 前端：http://localhost
- 后端 API：http://localhost:8080

4. 首次登录：
- 直接输入任意用户名和密码登录
- 系统会自动创建该账户

详细的 Docker 部署说明请查看 [DOCKER_DEPLOYMENT.md](DOCKER_DEPLOYMENT.md)

### 手动部署

#### 后端

1. 安装依赖：
```bash
cd backend
go mod download
```

2. 运行：
```bash
go run main.go
```

#### 前端

1. 安装依赖：
```bash
cd frontend
npm install
```

2. 开发模式：
```bash
npm run dev
```

3. 生产构建：
```bash
npm run build
```

## 环境变量

创建 `.env` 文件：

```bash
# JWT 密钥（生产环境必须修改）
JWT_SECRET=your-secure-random-secret-key

# 数据库路径
DB_PATH=dns-mng.db

# 服务器端口
SERVER_PORT=8080
```

## 定时任务

系统包含自动化的域名到期通知定时任务：

### 执行时间
- **首次运行**：后端服务启动后立即执行一次
- **定时运行**：每天早上 9:00 自动执行
- **执行间隔**：每 24 小时一次

### 查看日志

启动后端服务后，控制台会显示定时任务日志：

```bash
Starting domain expiry notification scheduler...
Next notification check scheduled at: 2026-04-10 09:00:00 (in 15h23m45s)
Checking for expiring domains...
Found 3 domain(s) that need notification
Sent notification for domain: example.com (expires in 15 days)
User 1: Notified about 2 domain(s): [example.com, test.com]
```

### 手动测试

重启后端服务即可立即触发一次检查：

```bash
cd backend
go run main.go
```

或使用测试脚本：

```bash
# Windows
cd backend
test_notification.bat

# Linux/Mac
cd backend
chmod +x test_notification.sh
./test_notification.sh
```

### 修改执行时间

编辑 `backend/service/scheduler_service.go`，修改小时数（0-23）：

```go
nextRun := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
// 将 9 改为你想要的小时数，例如 14 表示下午 2:00
```

详细的通知功能使用指南请查看 [backend/NOTIFICATION_GUIDE.md](backend/NOTIFICATION_GUIDE.md)

## 使用说明

### 1. 登录/注册

首次使用时，直接在登录页面输入用户名和密码：
- 如果用户不存在，系统会自动创建该账户
- 如果用户已存在，则验证密码登录

### 2. 添加 DNS 提供商账户

在"账户管理"页面添加你的 DNS 服务提供商账户：
- 输入账户名称
- 选择提供商
- 输入 API 密钥或 Token

### 3. 管理域名

- 在"所有域名"页面查看所有账户下的域名
- 支持搜索和过滤
- 显示域名数量统计
- 点击"管理记录"进入 DNS 记录管理

### 4. 管理 DNS 记录

- 添加、编辑、删除 DNS 记录
- 支持 A、AAAA、CNAME、MX、TXT、SPF、SRV 等记录类型
- 支持根域名记录（@）
- 支持记录启用/禁用（部分提供商）
- 显示记录状态和更新时间

## 支持的 DNS 提供商

| 提供商 | 状态 | 特性 |
|--------|------|------|
| Cloudflare | ✅ | 全功能支持 |
| 腾讯云 DNSPod | ✅ | 支持记录启用/禁用 |
| Dynu | ✅ | 免费动态 DNS |
| NDJP NET | ✅ | 日本 DNS 服务 |
| deSEC | ✅ | 免费开源，支持 DNSSEC |

### API 认证说明

- **Cloudflare**: API Token（推荐）或 Global API Key
- **腾讯云 DNSPod**: SecretId,SecretKey（逗号分隔）
- **Dynu**: API Key
- **NDJP NET**: Bearer Token
- **deSEC**: Token

## 添加新的 DNS 提供商

查看 [PROVIDER_GUIDE.md](PROVIDER_GUIDE.md) 了解如何添加新的 DNS 提供商支持。

## 开发

### 项目结构

```
.
├── backend/              # Go 后端
│   ├── config/          # 配置
│   ├── database/        # 数据库
│   ├── handler/         # HTTP 处理器
│   ├── middleware/      # 中间件
│   ├── models/          # 数据模型
│   ├── provider/        # DNS 提供商实现
│   │   ├── cloudflare/
│   │   ├── tencentcloud/
│   │   ├── dynu/
│   │   ├── ndjp/
│   │   └── desec/
│   └── service/         # 业务逻辑
├── frontend/            # React 前端
│   ├── src/
│   │   ├── components/  # 组件
│   │   ├── pages/       # 页面
│   │   ├── locales/     # 国际化
│   │   └── ...
│   └── ...
└── docker-compose.yaml  # Docker 配置
```

### 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

MIT License

## 更新日志

### v1.0.0 (2026-03-27)

- ✨ 支持 5 个 DNS 服务提供商
- 🎨 现代化 UI 设计
- 🌓 主题切换功能
- 🌍 中英文双语支持
- 📊 域名数量统计
- 📝 操作日志记录
- 🔍 搜索和过滤功能
- 🐳 Docker 一键部署

查看 [OPTIMIZATION_NOTES.md](OPTIMIZATION_NOTES.md) 了解详细的功能更新和优化。

## 链接

- GitHub: https://github.com/lizhixu/dns-mng
- Docker Hub: https://hub.docker.com/r/jacyli/dns-mng

## 致谢

感谢所有贡献者和使用者的支持！
