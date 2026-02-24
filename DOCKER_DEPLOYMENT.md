# Docker 部署指南

## 快速开始

### 1. 准备环境

确保已安装：
- Docker (20.10+)
- Docker Compose (2.0+)

### 2. 配置环境变量

复制环境变量模板：
```bash
cp .env.example .env
```

编辑 `.env` 文件，修改以下配置：
```bash
# 重要：修改 JWT 密钥（生产环境必须修改）
JWT_SECRET=your-secure-random-secret-key
```

### 3. 启动服务

使用 Docker Compose 启动所有服务：
```bash
docker-compose up -d
```

查看服务状态：
```bash
docker-compose ps
```

查看日志：
```bash
# 查看所有服务日志
docker-compose logs -f

# 查看后端日志
docker-compose logs -f backend

# 查看前端日志
docker-compose logs -f frontend
```

### 4. 访问应用

- 前端界面：http://localhost
- 后端 API：http://localhost:8080

### 5. 停止服务

```bash
# 停止服务
docker-compose stop

# 停止并删除容器
docker-compose down

# 停止并删除容器和数据卷（警告：会删除所有数据）
docker-compose down -v
```

## 服务说明

### 后端服务 (backend)

- **端口**：8080
- **数据持久化**：使用 Docker volume `dns-data` 存储数据库
- **健康检查**：每 30 秒检查一次 `/api/providers` 端点

### 前端服务 (frontend)

- **端口**：80
- **反向代理**：Nginx 自动将 `/api` 请求代理到后端
- **SPA 路由**：支持前端路由，刷新页面不会 404

## 生产环境部署建议

### 1. 使用 HTTPS

建议使用 Nginx 或 Traefik 作为反向代理，配置 SSL 证书：

```yaml
# docker-compose.prod.yaml
version: '3.8'

services:
  nginx-proxy:
    image: nginx:alpine
    ports:
      - "443:443"
      - "80:80"
    volumes:
      - ./nginx-ssl.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - frontend
```

### 2. 修改端口映射

如果 80 端口被占用，可以修改 `docker-compose.yaml`：

```yaml
services:
  frontend:
    ports:
      - "3000:80"  # 改为 3000 端口
```

### 3. 数据备份

定期备份数据库：

```bash
# 备份数据库
docker run --rm \
  -v dns-mng_dns-data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/dns-backup-$(date +%Y%m%d).tar.gz /data

# 恢复数据库
docker run --rm \
  -v dns-mng_dns-data:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/dns-backup-20240224.tar.gz -C /
```

### 4. 资源限制

为容器设置资源限制：

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### 5. 日志管理

配置日志轮转：

```yaml
services:
  backend:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## 更新应用

### 1. 拉取最新代码

```bash
git pull origin main
```

### 2. 重新构建并启动

```bash
docker-compose up -d --build
```

### 3. 清理旧镜像

```bash
docker image prune -f
```

## 故障排查

### 后端无法启动

检查日志：
```bash
docker-compose logs backend
```

常见问题：
- 端口 8080 被占用：修改 `docker-compose.yaml` 中的端口映射
- 数据库权限问题：检查 volume 权限

### 前端无法访问后端

检查 Nginx 配置：
```bash
docker-compose exec frontend cat /etc/nginx/conf.d/default.conf
```

检查网络连接：
```bash
docker-compose exec frontend ping backend
```

### 数据丢失

确保使用了 volume 持久化：
```bash
docker volume ls | grep dns-data
```

## 开发环境

如果需要在开发环境中使用 Docker：

```bash
# 使用开发配置
docker-compose -f docker-compose.yaml -f docker-compose.dev.yaml up
```

创建 `docker-compose.dev.yaml`：
```yaml
version: '3.8'

services:
  backend:
    volumes:
      - ./backend:/app
    command: go run main.go
    
  frontend:
    volumes:
      - ./frontend:/app
    command: npm run dev
```

## 监控和维护

### 查看资源使用

```bash
docker stats
```

### 查看容器健康状态

```bash
docker-compose ps
```

### 进入容器调试

```bash
# 进入后端容器
docker-compose exec backend sh

# 进入前端容器
docker-compose exec frontend sh
```

## 安全建议

1. **修改默认密钥**：生产环境必须修改 `JWT_SECRET`
2. **使用 HTTPS**：配置 SSL 证书
3. **限制访问**：使用防火墙限制端口访问
4. **定期更新**：及时更新 Docker 镜像和依赖
5. **备份数据**：定期备份数据库文件
