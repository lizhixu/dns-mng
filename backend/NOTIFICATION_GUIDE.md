# 域名到期通知功能使用指南

## 功能概述

DNS Manager 提供了自动化的域名到期提醒功能，可以在域名到期前自动发送邮件通知，帮助您及时续费，避免域名过期。

## 功能特性

- ✅ 每个域名可单独设置提前通知天数（1-365天）
- ✅ 支持启用/禁用单个域名的通知
- ✅ 支持 SMTP 邮件配置
- ✅ 邮件配置测试功能
- ✅ 定时任务每天自动运行（早上9:00）
- ✅ 避免重复通知（每天最多一次）
- ✅ 美观的 HTML 邮件模板
- ✅ 包含续费链接（如果已配置）

## 使用步骤

### 1. 配置邮件服务

访问"邮件通知"页面，配置 SMTP 邮件服务：

1. 启用邮件通知
2. 填写 SMTP 服务器信息：
   - SMTP 服务器地址
   - 端口号（通常是 587 或 465）
   - SMTP 用户名
   - SMTP 密码
   - 发件人邮箱
   - 发件人名称（可选）
   - **收件人邮箱**（接收通知的邮箱）

3. 保存配置
4. 发送测试邮件验证配置

### 2. 设置域名通知

在"所有域名"页面，为每个域名设置通知：

1. 点击域名旁边的"铃铛"图标
2. 启用到期提醒
3. 设置提前通知天数（例如：30天）
4. 保存设置

### 3. 配置续费信息（可选）

为了让通知邮件包含续费链接：

1. 点击域名旁边的"编辑"按钮
2. 填写续费日期
3. 填写续费链接（可选）
4. 保存

## 常用邮箱 SMTP 配置

### Gmail

- SMTP 服务器：`smtp.gmail.com`
- 端口：`587`
- 用户名：您的 Gmail 地址
- 密码：应用专用密码（需要在 Google 账户设置中生成）

**注意**：Gmail 需要启用"两步验证"并生成"应用专用密码"。

### QQ 邮箱

- SMTP 服务器：`smtp.qq.com`
- 端口：`587` 或 `465`
- 用户名：您的 QQ 邮箱地址
- 密码：授权码（不是 QQ 密码，需要在邮箱设置中生成）

### 163 邮箱

- SMTP 服务器：`smtp.163.com`
- 端口：`465`
- 用户名：您的 163 邮箱地址
- 密码：授权码（需要在邮箱设置中生成）

### Outlook

- SMTP 服务器：`smtp-mail.outlook.com`
- 端口：`587`
- 用户名：您的 Outlook 邮箱地址
- 密码：您的 Outlook 密码

## 定时任务说明

### 运行时间

- 定时任务每天早上 9:00 自动运行
- 首次启动后端服务时会立即运行一次

### 通知规则

1. 只有启用了通知的域名才会被检查
2. 只有设置了续费日期的域名才会被检查
3. 当剩余天数 ≤ 设置的提前天数时，发送通知
4. 每个域名每天最多通知一次
5. 已过期的域名不会发送通知

### 日志查看

定时任务的运行日志会输出到后端控制台：

```
Starting domain expiry notification scheduler...
Next notification check scheduled at: 2026-04-10 09:00:00
Checking for expiring domains...
Found 3 domain(s) that need notification
Sent notification for domain: example.com (expires in 15 days)
User 1: Notified about 3 domain(s): [example.com, test.com, demo.com]
```

## 邮件模板

通知邮件包含以下信息：

- 域名名称
- 到期日期
- 剩余天数（醒目显示）
- 续费链接按钮（如果已配置）

邮件采用 HTML 格式，美观易读。

## 故障排查

### 邮件发送失败

1. **检查 SMTP 配置**
   - 确认服务器地址和端口正确
   - 确认用户名和密码正确
   - 使用测试功能验证配置

2. **检查邮箱设置**
   - Gmail：确认已启用两步验证并生成应用专用密码
   - QQ/163：确认已生成授权码
   - 检查邮箱是否开启了 SMTP 服务

3. **检查网络连接**
   - 确认服务器可以访问 SMTP 服务器
   - 检查防火墙设置

### 未收到通知

1. **检查通知设置**
   - 确认域名已启用通知
   - 确认设置的提前天数合理
   - 确认域名已设置续费日期

2. **检查邮件配置**
   - 确认邮件通知已启用
   - 确认 SMTP 配置正确

3. **检查垃圾邮件**
   - 通知邮件可能被归类为垃圾邮件
   - 将发件人添加到白名单

4. **查看后端日志**
   - 检查定时任务是否正常运行
   - 查看是否有错误信息

## API 接口

### 获取通知设置

```http
GET /api/accounts/:accountId/domains/:domainId/notification
```

### 更新通知设置

```http
PUT /api/accounts/:accountId/domains/:domainId/notification
Content-Type: application/json

{
  "days_before": 30,
  "enabled": true
}
```

### 获取邮件配置

```http
GET /api/email/config
```

### 更新邮件配置

```http
PUT /api/email/config
Content-Type: application/json

{
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "smtp_username": "your-email@gmail.com",
  "smtp_password": "your-app-password",
  "from_email": "noreply@example.com",
  "from_name": "DNS Manager",
  "to_email": "recipient@example.com",
  "enabled": true
}
```

### 测试邮件配置

```http
POST /api/email/test
```

测试邮件将发送到配置的收件人邮箱（`to_email`）。

## 安全建议

1. **保护 SMTP 密码**
   - 使用应用专用密码而不是主密码
   - 定期更换密码
   - 不要在代码中硬编码密码

2. **使用 TLS 加密**
   - 优先使用支持 TLS 的端口（587）
   - 确保 SMTP 连接加密

3. **限制发件频率**
   - 系统已内置每天最多一次的限制
   - 避免被邮件服务商标记为垃圾邮件

## 常见问题

**Q: 可以修改定时任务的运行时间吗？**

A: 可以。修改 `backend/service/scheduler_service.go` 中的时间设置：

```go
nextRun := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
```

将 `9` 改为您想要的小时数（0-23）。

**Q: 可以设置多个收件人吗？**

A: 当前版本每个用户只能设置一个邮箱（用户名）。如需多个收件人，可以：
1. 使用邮箱的转发功能
2. 使用邮件组/分发列表

**Q: 邮件会被标记为垃圾邮件吗？**

A: 可能性较小，但建议：
1. 使用可信的 SMTP 服务
2. 配置正确的发件人信息
3. 将发件人添加到白名单

**Q: 可以自定义邮件模板吗？**

A: 可以。修改 `backend/service/email_service.go` 中的 `SendExpiryNotification` 方法。

## 更新日志

### v1.0.0
- ✨ 初始版本
- ✅ 支持 SMTP 邮件通知
- ✅ 支持单域名通知设置
- ✅ 每日定时任务
- ✅ HTML 邮件模板
