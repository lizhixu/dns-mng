# 推送镜像到 Docker Hub

## 方法一：使用脚本（推荐）

### Windows
```cmd
build-and-push.bat
```

或指定版本：
```cmd
build-and-push.bat v1.0.0
```

### Linux/Mac
```bash
chmod +x build-and-push.sh
./build-and-push.sh
```

或指定版本：
```bash
./build-and-push.sh v1.0.0
```

## 方法二：手动构建和推送

### 1. 登录 Docker Hub
```bash
docker login
# 输入用户名: jacyli
# 输入密码: [your-password]
```

### 2. 构建镜像

构建后端镜像：
```bash
docker build -t jacyli/dns-mng:backend ./backend
docker tag jacyli/dns-mng:backend jacyli/dns-mng:backend-v1.0.0
```

构建前端镜像：
```bash
docker build -t jacyli/dns-mng:frontend ./frontend
docker tag jacyli/dns-mng:frontend jacyli/dns-mng:frontend-v1.0.0
```

### 3. 推送镜像

推送后端镜像：
```bash
docker push jacyli/dns-mng:backend
docker push jacyli/dns-mng:backend-v1.0.0
```

推送前端镜像：
```bash
docker push jacyli/dns-mng:frontend
docker push jacyli/dns-mng:frontend-v1.0.0
```

## 使用 Docker Hub 镜像

### 直接使用
```bash
# 拉取镜像
docker pull jacyli/dns-mng:backend
docker pull jacyli/dns-mng:frontend

# 使用 docker-compose 启动
docker-compose pull
docker-compose up -d
```

### 单独运行后端
```bash
docker run -d \
  --name dns-mng-backend \
  -p 8080:8080 \
  -v dns-data:/data \
  -e JWT_SECRET=your-secret-key \
  jacyli/dns-mng:backend
```

### 单独运行前端
```bash
docker run -d \
  --name dns-mng-frontend \
  -p 80:80 \
  --link dns-mng-backend:backend \
  jacyli/dns-mng:frontend
```

## 镜像信息

- **仓库**: jacyli/dns-mng
- **标签**:
  - `backend` - 后端最新版本
  - `backend-v1.0.0` - 后端特定版本
  - `frontend` - 前端最新版本
  - `frontend-v1.0.0` - 前端特定版本

## 镜像大小

- 后端镜像: ~30MB (Alpine Linux + Go binary)
- 前端镜像: ~25MB (Alpine Linux + Nginx + 静态文件)

## 注意事项

1. 确保已登录 Docker Hub
2. 构建过程可能需要几分钟
3. 推送大镜像需要良好的网络连接
4. 建议为每个版本打标签以便回滚

## 故障排查

### 构建失败
```bash
# 清理 Docker 缓存
docker builder prune -a

# 重新构建
docker-compose build --no-cache
```

### 推送失败
```bash
# 检查登录状态
docker info | grep Username

# 重新登录
docker logout
docker login
```

### 镜像过大
```bash
# 查看镜像大小
docker images | grep dns-mng

# 查看镜像层
docker history jacyli/dns-mng:backend
```
