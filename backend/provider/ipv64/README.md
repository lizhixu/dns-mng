# IPv64.net DNS Provider

IPv64.net 是一个提供免费 DNS 托管服务的供应商，支持动态 DNS (DDNS) 功能。

## 功能特点

- ✅ 免费 DNS 托管
- ✅ 支持 DDNS
- ✅ 支持多种记录类型（A, AAAA, CNAME, MX, TXT, NS, SRV）
- ✅ 简单的 API
- ✅ 自动创建子域名

## API 信息

- **Base URL**: `https://ipv64.net/api.php`
- **认证方式**: Bearer Token
- **API 限制**: 
  - 账户限制：根据账户等级（默认 64 次/24小时）
  - 调用限制：10秒内最多 5 次请求

## 获取 API Token

1. 访问 [IPv64.net](https://ipv64.net)
2. 注册账户并登录
3. 进入账户设置
4. 找到 API Key
5. 复制 API Key 用于配置

## 配置说明

在系统中添加 IPv64.net 账户时：

- **账户名称**: 自定义名称（如 "My IPv64 Account"）
- **供应商**: 选择 "IPv64.net"
- **API Key**: 粘贴你的 API Token

## API 端点

### 获取所有域名和记录
```
GET /api.php?get_domains
Authorization: Bearer {api_key}
```

### 创建域名
```
POST /api.php
Authorization: Bearer {api_key}
Content-Type: application/x-www-form-urlencoded

add_domain=domainname.ipv64.net
```

### 删除域名
```
DELETE /api.php
Authorization: Bearer {api_key}
Content-Type: application/x-www-form-urlencoded

del_domain=domainname.ipv64.net
```

### 添加 DNS 记录
```
POST /api.php
Authorization: Bearer {api_key}
Content-Type: application/x-www-form-urlencoded

add_record=domainname.ipv64.net
praefix=www
type=A
content=1.2.3.4
```

### 删除 DNS 记录（通过 ID）
```
DELETE /api.php
Authorization: Bearer {api_key}
Content-Type: application/x-www-form-urlencoded

del_record=603103
```

### 删除 DNS 记录（通过详细信息）
```
DELETE /api.php
Authorization: Bearer {api_key}
Content-Type: application/x-www-form-urlencoded

del_record=domainname.ipv64.net
praefix=www
type=A
content=1.2.3.4
```

## 支持的记录类型

- A (IPv4 地址)
- AAAA (IPv6 地址)
- CNAME (别名)
- MX (邮件交换)
- TXT (文本记录)
- NS (名称服务器)
- SRV (服务记录)

## 响应格式

### 成功响应
```json
{
  "info": "success",
  "status": "200 OK"
}
```

### 获取域名响应
```json
{
  "subdomains": {
    "example.ipv64.net": {
      "updates": 0,
      "wildcard": 1,
      "domain_update_hash": "...",
      "ipv6prefix": "",
      "dualstack": "",
      "deactivated": 0,
      "records": [
        {
          "record_id": 603103,
          "content": "1.2.3.4",
          "ttl": 60,
          "type": "A",
          "praefix": "www",
          "last_update": "2026-04-13 04:16:29",
          "record_key": "...",
          "deactivated": 0,
          "failover_policy": "0"
        }
      ]
    }
  },
  "info": "success",
  "status": "200 OK"
}
```

## 注意事项

1. **域名格式**
   - 必须使用 IPv64.net 的子域名格式（如 `example.ipv64.net`）
   - 或使用其他支持的域名后缀（如 `any64.de`）

2. **记录前缀（praefix）**
   - 空字符串表示根记录（@）
   - 其他值表示子域名（如 "www"）

3. **TTL 设置**
   - 默认 TTL 为 60 秒
   - 系统会自动设置，无法通过 API 修改

4. **更新记录**
   - IPv64 API 不支持直接更新记录
   - 系统会先删除旧记录，再创建新记录

5. **API 限制**
   - 注意 API 调用频率限制（10秒内最多 5 次）
   - 注意每日调用次数限制（默认 64 次/24小时）

## DDNS 支持

IPv64.net 原生支持 DDNS 功能，每个域名都有一个 `domain_update_hash`：

```bash
# 使用域名更新哈希进行 DDNS 更新
curl "https://ipv64.net/update.php?domain=example.ipv64.net&hash=YOUR_HASH"
```

也可以使用本系统的 DDNS 功能来管理 IPv64.net 的记录。

## 示例

### 添加 A 记录
```bash
curl -X POST https://ipv64.net/api.php \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d "add_record=example.ipv64.net" \
  -d "praefix=www" \
  -d "type=A" \
  -d "content=1.2.3.4"
```

### 添加根记录
```bash
curl -X POST https://ipv64.net/api.php \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d "add_record=example.ipv64.net" \
  -d "praefix=" \
  -d "type=A" \
  -d "content=1.2.3.4"
```

### 删除记录（通过 ID）
```bash
curl -X DELETE https://ipv64.net/api.php \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d "del_record=603103"
```

## 故障排除

### 认证失败
- 检查 API Key 是否正确
- 确认 API Key 没有过期
- 验证使用了正确的认证方式（Bearer Token）

### 域名不存在
- 确认域名格式正确（必须是 IPv64.net 的子域名）
- 检查域名是否已创建

### API 限制错误（429）
- 等待一段时间后重试
- 减少 API 调用频率
- 考虑升级账户等级

### 记录创建失败
- 检查记录类型是否支持
- 验证记录内容格式是否正确
- 确认没有重复的记录

## 相关链接

- [IPv64.net 官网](https://ipv64.net)
- [IPv64.net API 文档](https://ipv64.net/api)

## 技术支持

如遇到问题，可以：
1. 查看 IPv64.net 官方文档
2. 联系 IPv64.net 技术支持
3. 在本项目提交 Issue
