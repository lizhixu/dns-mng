# 故障排查指南

## 前端打包后无法访问接口

### 问题描述
前端打包后（`npm run build`），访问页面时无法连接到后端 API。

### 可能的原因和解决方案

#### 1. 使用 Docker 部署（推荐）

如果使用 Docker Compose 部署，Nginx 会自动代理 `/api` 请求到后端：

```bash
docker-compose up -d
```

访问 `http://localhost`，API 请求会自动转发到后端。

#### 2. 单独部署前端

如果只部署前端静态文件，需要配置 Web 服务器代理：

**Nginx 配置示例**：
```nginx
server {
    listen 80;
    root /path/to/dist;
    
    # API 代理
    location /api {
        proxy_pass http://your-backend-server:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    # SPA 路由
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

**Apache 配置示例**：
```apache
<VirtualHost *:80>
    DocumentRoot /path/to/dist
    
    # API 代理
    ProxyPass /api http://your-backend-server:8080/api
    ProxyPassReverse /api http://your-backend-server:8080/api
    
    # SPA 路由
    <Directory /path/to/dist>
        RewriteEngine On
        RewriteBase /
        RewriteRule ^index\.html$ - [L]
        RewriteCond %{REQUEST_FILENAME} !-f
        RewriteCond %{REQUEST_FILENAME} !-d
        RewriteRule . /index.html [L]
    </Directory>
</VirtualHost>
```

#### 3. 开发环境

开发环境使用 Vite 的代理功能：

```bash
npm run dev
```

Vite 会自动将 `/api` 请求代理到 `http://localhost:8080`。

#### 4. 检查后端是否运行

```bash
# 检查后端是否启动
curl http://localhost:8080/api/providers

# 或使用浏览器访问
# http://localhost:8080/api/providers
```

#### 5. 检查 CORS 配置

后端已配置 CORS 中间件，允许跨域请求。如果仍有问题，检查 `backend/middleware/cors.go`。

#### 6. 检查网络连接

在浏览器控制台查看网络请求：

1. 打开浏览器开发者工具（F12）
2. 切换到 Network 标签
3. 刷新页面
4. 查看 `/api` 请求的状态

常见错误：
- `404 Not Found`: 后端未运行或路由配置错误
- `502 Bad Gateway`: Nginx 无法连接到后端
- `CORS Error`: CORS 配置问题
- `ERR_CONNECTION_REFUSED`: 后端服务未启动

### 推荐部署方式

#### 方式一：Docker Compose（最简单）

```bash
# 1. 克隆项目
git clone https://github.com/lizhixu/dns-mng.git
cd dns-mng

# 2. 启动服务
docker-compose up -d

# 3. 访问
# http://localhost
```

#### 方式二：分离部署

**后端**：
```bash
cd backend
go build
./dns-mng
# 后端运行在 :8080
```

**前端**：
```bash
cd frontend
npm install
npm run build

# 将 dist 目录部署到 Nginx
# 配置 Nginx 代理 /api 到后端
```

#### 方式三：使用 Docker Hub 镜像

```bash
# 拉取镜像
docker pull jacyli/dns-mng:backend
docker pull jacyli/dns-mng:frontend

# 使用 docker-compose
docker-compose pull
docker-compose up -d
```

## 其他常见问题

### 登录后跳转到空白页

检查路由配置，确保默认路由正确：
- 登录后应跳转到 `/domains`
- 根路径 `/` 应重定向到 `/domains`

### 主题切换不生效

清除浏览器缓存：
```bash
# Chrome/Edge
Ctrl + Shift + Delete

# 或使用无痕模式测试
Ctrl + Shift + N
```

### 数据库文件权限问题

Docker 环境中，确保数据卷有正确的权限：
```bash
docker-compose down
docker volume rm dns-mng_dns-data
docker-compose up -d
```

### 端口被占用

修改 `docker-compose.yaml` 中的端口映射：
```yaml
services:
  frontend:
    ports:
      - "3000:80"  # 改为 3000 端口
  backend:
    ports:
      - "8081:8080"  # 改为 8081 端口
```

## 获取帮助

如果问题仍未解决：

1. 查看日志：
```bash
# Docker 日志
docker-compose logs -f

# 后端日志
docker-compose logs backend

# 前端日志
docker-compose logs frontend
```

2. 检查服务状态：
```bash
docker-compose ps
```

3. 提交 Issue：
https://github.com/lizhixu/dns-mng/issues
