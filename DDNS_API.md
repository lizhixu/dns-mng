# DDNS API 文档

## 概述

DDNS (Dynamic DNS) 功能允许用户通过简单的 HTTP 请求自动更新 DNS 记录的 IP 地址。整个系统只需要**一个 Token**，所有域名共享使用。

**API 格式兼容 DuckDNS**，可以使用标准的 DDNS 客户端直接对接。

## 架构设计

### 数据存储
- Token 存储在数据库中（`ddns_tokens` 表）
- **用户级别**: 每个用户只有一个 Token，可更新该用户下所有账户的所有域名
- 支持启用/禁用状态
- 记录最后使用时间和 IP

### 安全性
- Token 存储在数据库中，支持随时撤销
- 支持自定义 token 或自动生成随机 token
- 可以启用/禁用 token 而不删除
- 记录每次使用的 IP 和时间

## API 端点

### 1. 获取 DDNS Token

**端点**: `GET /api/ddns-token`

**认证**: 需要登录

**响应**:
```json
{
  "has_token": true,
  "token": {
    "id": 1,
    "user_id": 1,
    "token": "abc123def456...",
    "enabled": true,
    "last_used_at": "2024-01-01T12:00:00Z",
    "last_ip": "1.2.3.4",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### 2. 创建/更新 DDNS Token

**端点**: `PUT /api/ddns-token`

**认证**: 需要登录

**请求体**:
```json
{
  "token": "my-custom-token",
  "enabled": true
}
```

**说明**:
- `token`: 可选，不传则自动生成随机 token
- `enabled`: 可选，不传则保持当前状态

**响应**: 同获取 token 响应

### 3. 删除 DDNS Token

**端点**: `DELETE /api/ddns-token`

**认证**: 需要登录

**响应**:
```json
{
  "message": "token deleted successfully"
}
```

### 4. DDNS 更新接口（公开，DuckDNS 兼容）

**端点**: `GET /api/ddns/update`

**认证**: 使用 token 参数

**查询参数** (DuckDNS 格式):
- `domains` (必需): 要更新的域名（逗号分隔，例如：`example.com,sub.example.com`）
- `token` (必需): DDNS token（用户级别）
- `ip` (可选): IPv4 地址，不提供则使用客户端 IP
- `ipv6` (可选): IPv6 地址

**示例**:
```bash
# 使用客户端 IP 更新（最常用）
curl "https://your-domain.com/api/ddns/update?domains=example.com&token=your-token"

# 指定 IPv4 地址更新
curl "https://your-domain.com/api/ddns/update?domains=example.com&token=your-token&ip=1.2.3.4"

# 指定 IPv6 地址更新
curl "https://your-domain.com/api/ddns/update?domains=example.com&token=your-token&ipv6=2001:db8::1"

# 同时更新多个域名
curl "https://your-domain.com/api/ddns/update?domains=example.com,sub.example.com&token=your-token&ip=1.2.3.4"
```

**成功响应**:
```
OK
```

**失败响应**:
```
KO
```

## 使用流程

### 1. 获取 Token

登录后调用 `GET /api/ddns-token` 查看是否已有 Token。

如果没有，调用 `PUT /api/ddns-token` 创建：
```json
{
  "token": "my-custom-token"
}
```

或让系统自动生成（不传 token 字段）。

### 2. 路由器/客户端配置

在路由器 DDNS 设置中选择 **DuckDNS** 或自定义 URL：

```
服务提供商: DuckDNS 或 自定义
域名: your-domain-name（在 DNS 管理中添加的域名）
Token: your-ddns-token（从系统获取）
更新 URL: https://your-domain.com/api/ddns/update?domains=%s&token=%s&ip=%s
```

### 3. 定时任务

可以设置 cron 任务定期更新：

```bash
# 每 5 分钟检查一次
*/5 * * * * curl -s "https://your-domain.com/api/ddns/update?domains=example.com&token=your-token" > /dev/null 2>&1
```

### 4. Python 脚本示例

```python
import requests
import time

DOMAINS = "example.com"
TOKEN = "your_ddns_token_here"
URL = f"https://your-domain.com/api/ddns/update?domains={DOMAINS}&token={TOKEN}"

while True:
    try:
        response = requests.get(URL)
        result = response.text
        print(f"Status: {result}")
    except Exception as e:
        print(f"Error: {e}")

    # 每 5 分钟更新一次
    time.sleep(300)
```

## 支持的记录类型

- **A 记录**: IPv4 地址
- **AAAA 记录**: IPv6 地址

系统会自动匹配域名下的 A/AAAA 记录并更新。

## 安全建议

1. **Token 管理**
   - 使用强随机 token（系统自动生成 64 字符）
   - 定期更换 token
   - 不再使用时及时删除

2. **访问控制**
   - 可以通过启用/禁用功能临时停用 token
   - 监控最后使用时间和 IP，发现异常及时处理

3. **HTTPS**
   - 强烈建议使用 HTTPS 传输 token
   - 防止 token 在传输过程中被窃取

## 数据库表结构

```sql
-- DDNS Token 表（用户级别，每个用户一个 Token）
CREATE TABLE IF NOT EXISTS ddns_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    token TEXT NOT NULL UNIQUE,
    enabled INTEGER DEFAULT 1,
    last_used_at DATETIME,
    last_ip TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_ddns_tokens_token ON ddns_tokens(token);
```

## 错误处理

| 响应 | 说明 | 解决方案 |
|------|------|---------|
| `OK` | 更新成功 | - |
| `KO` | 更新失败 | 检查 token 是否正确、域名是否存在 |

## 日志记录

每次 DDNS 更新都会记录操作日志：
- 操作类型: update
- 资源类型: ddns
- 详细信息: 域名、IP、DDNS 标记
- 客户端 IP

可以在操作日志页面查看所有 DDNS 更新历史。

## 功能特点

- ✅ 每个用户只需一个 Token
- ✅ 自动匹配并更新所有账户下的域名记录
- ✅ 支持 IPv4 和 IPv6
- ✅ 自动检测客户端 IP
- ✅ **DuckDNS API 兼容**
- ✅ 启用/禁用功能
- ✅ 记录使用历史（时间和 IP）
- ✅ 完整的操作日志
