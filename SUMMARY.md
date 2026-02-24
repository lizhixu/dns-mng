# DNS Manager - 项目总结

## 项目信息

- **仓库地址**: https://github.com/lizhixu/dns-mng.git
- **技术栈**: Go + React + SQLite
- **部署方式**: Docker Compose

## 主要功能

✅ 多 DNS 提供商支持（当前支持 Dynu）
✅ 用户认证系统（JWT）
✅ 账户管理
✅ 域名管理
✅ DNS 记录管理（A、AAAA、CNAME、MX、TXT、SPF、SRV）
✅ 根域名记录支持
✅ 亮色/暗色主题切换（支持跟随系统）
✅ 中英文双语支持
✅ 响应式设计
✅ Docker 一键部署

## 快速开始

```bash
# 克隆项目
git clone https://github.com/lizhixu/dns-mng.git
cd dns-mng

# 使用 Docker 启动
chmod +x start.sh
./start.sh

# 访问应用
# 前端: http://localhost
# 后端: http://localhost:8080
```

## 项目结构

```
dns-mng/
├── backend/          # Go 后端
│   ├── config/      # 配置
│   ├── database/    # 数据库
│   ├── handler/     # HTTP 处理器
│   ├── middleware/  # 中间件
│   ├── models/      # 数据模型
│   ├── provider/    # DNS 提供商
│   └── service/     # 业务逻辑
├── frontend/        # React 前端
│   └── src/
│       ├── components/  # 组件
│       ├── pages/       # 页面
│       └── locales/     # 国际化
└── docker-compose.yaml  # Docker 配置
```

## 设计特点

- **Vercel 风格 UI**: 简洁、现代、高对比度
- **极简设计**: 去除不必要的装饰，专注功能
- **响应式布局**: 适配各种屏幕尺寸
- **主题系统**: 支持亮色/暗色/跟随系统
- **国际化**: 完整的中英文支持

## 技术亮点

1. **后端**
   - RESTful API 设计
   - JWT 身份验证
   - 提供商抽象接口，易于扩展
   - SQLite 轻量级数据库

2. **前端**
   - React Hooks
   - Context API 状态管理
   - CSS 变量主题系统
   - Vite 构建工具

3. **部署**
   - Docker 多阶段构建
   - Nginx 反向代理
   - 数据持久化
   - 健康检查

## 文档

- [README.md](README.md) - 项目介绍
- [DOCKER_DEPLOYMENT.md](DOCKER_DEPLOYMENT.md) - Docker 部署指南
- [OPTIMIZATION_NOTES.md](OPTIMIZATION_NOTES.md) - 优化记录

## 许可证

MIT License
