# 操作日志规范化方案

## 当前问题
1. 日志记录不统一（action、resource 字段混乱）
2. 中文显示不友好（直接显示英文字段）
3. 缺少统一的翻译映射

## 规范化方案

### 1. Action 标准化
统一使用小写英文：
- `create` - 创建
- `update` - 更新
- `delete` - 删除
- `soft_delete` - 软删除
- `restore` - 恢复
- `login` - 登录
- `logout` - 登出
- `test` - 测试
- `refresh` - 刷新
- `batch_update` - 批量更新
- `batch_delete` - 批量删除

### 2. Resource 标准化
统一使用小写英文：
- `account` - 账户
- `domain` - 域名
- `domain_cache` - 域名缓存
- `record` - DNS记录
- `email_config` - 邮件配置
- `notification` - 通知设置
- `user` - 用户
- `password` - 密码

### 3. 前端翻译映射

在 `frontend/src/locales/zh.js` 添加：

```javascript
logs: {
  actions: {
    create: '创建',
    update: '更新',
    delete: '删除',
    soft_delete: '软删除',
    restore: '恢复',
    login: '登录',
    logout: '登出',
    test: '测试',
    refresh: '刷新',
    batch_update: '批量更新',
    batch_delete: '批量删除'
  },
  resources: {
    account: '账户',
    domain: '域名',
    domain_cache: '域名缓存',
    record: 'DNS记录',
    email_config: '邮件配置',
    notification: '通知设置',
    user: '用户',
    password: '密码'
  }
}
```

### 4. 后端日志记录规范

所有 handler 中的日志记录应遵循：

```go
h.logService.CreateLog(
    userID,
    "action",      // 使用标准化的 action
    "resource",    // 使用标准化的 resource
    resourceID,    // 资源ID（如果有）
    map[string]interface{}{
        // 详细信息，使用英文key
        "key": "value",
    },
    c.ClientIP(),
)
```

### 5. 需要修改的文件

#### 后端：
- `backend/handler/dns_handler.go` - 统一 record 相关日志
- `backend/handler/domain_cache_handler.go` - 统一 domain_cache 日志
- `backend/handler/account_handler.go` - 统一 account 日志
- `backend/handler/auth_handler.go` - 统一 login/password 日志
- `backend/handler/notification_handler.go` - 统一 email_config/notification 日志

#### 前端：
- `frontend/src/locales/zh.js` - 添加翻译映射
- `frontend/src/locales/en.js` - 添加英文映射
- `frontend/src/pages/LogsManagement.jsx` - 使用翻译函数显示

### 6. 实现优先级

1. ✅ 高优先级：添加前端翻译映射
2. ✅ 高优先级：修改日志显示组件使用翻译
3. ✅ 高优先级：规范化 auth_handler.go 日志记录
4. 📝 中优先级：检查其他 handler 的日志记录（已基本规范）
5. 📝 低优先级：添加更多详细信息字段

## 已完成的改进

### 后端日志规范化
- ✅ `auth_handler.go` - 已规范化所有日志记录
  - `register` → `create` + `user`
  - `login_failed` → `login` + `user` (添加 success: false)
  - `login` → `login` + `user` (添加 success: true)
  - `update_password` → `update` + `password`

### 前端翻译映射
- ✅ 添加了完整的 action 翻译（中英文）
- ✅ 添加了完整的 resource 翻译（中英文）
- ✅ 日志显示组件已使用翻译函数

### 其他 Handler 检查结果
- ✅ `domain_cache_handler.go` - 已规范（使用标准 action 和 resource）
- ✅ `notification_handler.go` - 已规范
- ✅ `dns_handler.go` - 已规范
- ✅ `account_handler.go` - 已规范

## 示例

### 创建DNS记录
```go
h.logService.CreateLog(userID, "create", "record", record.ID, map[string]interface{}{
    "domain": domainName,
    "type": record.RecordType,
    "name": record.NodeName,
    "value": record.Content,
}, c.ClientIP())
```

前端显示：**创建** DNS记录 #abc123

### 更新邮件配置
```go
h.logService.CreateLog(userID, "update", "email_config", "", map[string]interface{}{
    "smtp_host": req.SMTPHost,
    "enabled": req.Enabled,
}, c.ClientIP())
```

前端显示：**更新** 邮件配置

### 软删除域名
```go
h.logService.CreateLog(userID, "soft_delete", "domain", "", map[string]interface{}{
    "count": len(req.Items),
    "domains": domainNames,
}, c.ClientIP())
```

前端显示：**软删除** 域名 (3个)
