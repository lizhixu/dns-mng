# DNS Manager

一个现代化的 DNS 记录管理系统，支持多个 DNS 服务提供商。

## 功能特性

- 🌐 **多提供商支持**：支持 Dynu 等 DNS 服务提供商
- 🔐 **安全认证**：JWT 身份验证
- 🎨 **现代 UI**：Vercel 风格的简洁界面
- 🌓 **主题切换**：支持亮色/暗色模式，可跟随系统
- 🌍 **多语言**：支持中文和英文
- 📱 **响应式设计**：适配各种屏幕尺寸
- 🐳 **Docker 支持**：一键部署

## 技术栈

### 后端
- Go 1.21+
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
git clone <repository-url>
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

## 使用说明

### 1. 注册账号

首次使用需要注册一个账号。

### 2. 添加 DNS 提供商账户

在"账户管理"页面添加你的 DNS 服务提供商账户：
- 输入账户名称
- 选择提供商（如 Dynu）
- 输入 API 密钥

### 3. 管理域名

- 查看所有域名列表
- 点击"管理记录"进入 DNS 记录管理

### 4. 管理 DNS 记录

- 添加、编辑、删除 DNS 记录
- 支持 A、AAAA、CNAME、MX、TXT、SPF、SRV 等记录类型
- 支持根域名记录（@）

## 支持的 DNS 提供商

- ✅ Dynu.com
- 🚧 更多提供商开发中...

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

## 许可证

MIT License

## 更新日志

查看 [OPTIMIZATION_NOTES.md](OPTIMIZATION_NOTES.md) 了解最新的功能更新和优化。
