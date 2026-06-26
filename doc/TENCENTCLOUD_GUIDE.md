# 腾讯云 DNSPod 使用指南

## 重要说明

本功能使用腾讯云 DNSPod API 进行公网 DNS 解析管理。

**注意：** 腾讯云 DNSPod SDK 可能在未来版本中废弃，建议关注腾讯云官方文档的更新。

## 获取 API 密钥

腾讯云 DNSPod 支持两种认证方式，本系统使用**腾讯云 API 密钥**方式（推荐）。

### 认证方式说明

#### 方式一：腾讯云 API 密钥（推荐）✅
- 使用腾讯云统一的 SecretId 和 SecretKey
- 支持子账号和权限管理
- 本系统采用此方式

#### 方式二：DNSPod Token（不支持）
- DNSPod 专用的 API Token
- 本系统暂不支持此方式

### 获取腾讯云 API 密钥

### 1. 登录腾讯云控制台

访问 [腾讯云控制台](https://console.cloud.tencent.com/)

### 2. 进入访问管理

点击右上角账号 → 访问管理 → [API密钥管理](https://console.cloud.tencent.com/cam/capi)

### 3. 创建密钥

1. 点击"新建密钥"
2. 完成身份验证
3. 获取 SecretId 和 SecretKey

**重要提示：**
- SecretKey 只在创建时显示一次，请妥善保存
- 建议使用子账号密钥，并授予最小权限
- 定期轮换密钥以提高安全性

### 4. 授权 DNSPod 权限

如果使用子账号，需要授予以下权限：
- QcloudDNSPodFullAccess（DNSPod 全读写访问权限）

或者自定义策略，包含以下操作：
```json
{
  "version": "2.0",
  "statement": [
    {
      "effect": "allow",
      "action": [
        "dnspod:DescribeDomainList",
        "dnspod:DescribeDomain",
        "dnspod:DescribeRecordList",
        "dnspod:CreateRecord",
        "dnspod:ModifyRecord",
        "dnspod:DeleteRecord"
      ],
      "resource": "*"
    }
  ]
}
```

## 在 DNS Manager 中配置

### 1. 添加账户

在 DNS Manager 的"账户管理"页面：
1. 点击"添加账户"
2. 输入账户名称（例如：我的腾讯云）
3. 选择提供商：腾讯云 DNSPod
4. 输入 API 密钥

### 2. API 密钥格式

API 密钥格式为：`SecretId,SecretKey`

例如：
```
AKIDxxxxxxxxxxxxxxxxxxxxx,xxxxxxxxxxxxxxxxxxxxxxxx
```

**注意：**
- SecretId 和 SecretKey 之间用英文逗号分隔
- 不要有多余的空格

### 3. 验证配置

保存后，系统会自动获取您的域名列表。如果配置正确，您将看到所有在腾讯云 DNSPod 中的域名。

## 功能说明

### 支持的记录类型

- A 记录（IPv4 地址）
- AAAA 记录（IPv6 地址）
- CNAME 记录（别名）
- MX 记录（邮件服务器）
- TXT 记录（文本记录）
- SRV 记录（服务记录）

### 默认值

- TTL：600 秒（10 分钟）
- 解析线路：默认

### 限制说明

- NS 记录不可编辑（由系统管理）
- 记录修改后可能需要几分钟生效
- 免费版 DNSPod 有记录数量限制

## 常见问题

### Q: 提示"invalid API key format"

A: 请检查 API 密钥格式是否正确：
- 格式：`SecretId,SecretKey`
- 确保中间是英文逗号
- 确保没有多余的空格或换行

### Q: 提示"AuthFailure"

A: 可能的原因：
- SecretId 或 SecretKey 错误
- 密钥已被删除或禁用
- 子账号没有 DNSPod 权限

### Q: 看不到域名列表

A: 请确认：
- 域名已添加到腾讯云 DNSPod
- API 密钥有正确的权限
- 域名状态正常（未暂停）

### Q: 记录修改后不生效

A: DNS 记录修改需要时间生效：
- 通常 10 分钟内生效
- 可能受本地 DNS 缓存影响
- 可以使用 `nslookup` 或 `dig` 命令验证

## 安全建议

1. **使用子账号**
   - 不要使用主账号密钥
   - 为 DNS Manager 创建专用子账号
   - 只授予必要的 DNSPod 权限

2. **定期轮换密钥**
   - 建议每 90 天更换一次密钥
   - 删除不再使用的密钥

3. **保护密钥安全**
   - 不要将密钥提交到代码仓库
   - 不要在公开场合分享密钥
   - 如果密钥泄露，立即删除并重新创建

4. **监控 API 调用**
   - 在腾讯云控制台查看 API 调用日志
   - 发现异常调用及时处理

## 相关链接

- [腾讯云 DNSPod 控制台](https://console.dnspod.cn/)
- [腾讯云 API 密钥管理](https://console.cloud.tencent.com/cam/capi)
- [DNSPod API 文档](https://cloud.tencent.com/document/product/1427)
- [访问管理文档](https://cloud.tencent.com/document/product/598)
