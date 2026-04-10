# 域名列表刷新功能优化实现方案

## 实现状态

✅ **已完成**
- 数据库结构更新（添加 deleted_at, last_sync_at, provider_updated_on 字段）
- 显示域名更新时间（updated_on）
- 显示数据缓存时间
- 软删除机制（BatchSoftDeleteDomains API）
- 自动恢复机制（BatchRestoreDomains API）
- 删除确认对话框
- 自动恢复提示
- 定时任务排除已删除域名（notification_service已实现）
- 前端API集成
- 后端完整实现

## 功能说明

1. **域名更新时间显示**
   - 在域名列表中显示域名的最后更新时间（updated_on字段）
   - 如果没有更新时间，显示为空

2. **数据缓存时间**
   - 在页面顶部显示数据的缓存时间
   - 用户可以知道当前数据的新鲜度

3. **软删除机制**
   - 刷新时发现域名不存在于服务商，提示用户是否删除
   - 实现软删除（设置deleted_at字段）
   - 下次刷新时如果域名重新出现，自动激活（清除deleted_at）

4. **定时任务逻辑更新**
   - 定时任务获取域名时，排除已软删除的域名

## 数据库结构

domain_cache表已有字段：
- id
- user_id
- account_id
- domain_id
- domain_name
- renewal_date
- renewal_url
- created_at
- updated_at
- deleted_at (软删除标记)

需要添加的字段：
- last_sync_at (最后同步时间)
- provider_updated_on (服务商返回的更新时间)

## 实现步骤

### 1. 数据库迁移

```sql
ALTER TABLE domain_cache ADD COLUMN last_sync_at DATETIME;
ALTER TABLE domain_cache ADD COLUMN provider_updated_on DATETIME;
```

### 2. 后端API修改

#### 2.1 更新Domain模型
添加字段：
- UpdatedOn (服务商返回的更新时间)
- LastSyncAt (最后同步时间)
- IsDeleted (是否已删除)

#### 2.2 修改RefreshAllDomains接口
返回数据结构：
```json
{
  "domains": [...],
  "domains_to_delete": [...],
  "cache_timestamp": "2024-01-01T12:00:00Z",
  "has_changes": true
}
```

#### 2.3 添加软删除/恢复接口
- POST /api/domains/batch-soft-delete
- POST /api/domains/batch-restore

#### 2.4 修改domain_cache_service
- 软删除：设置deleted_at字段
- 恢复：清除deleted_at字段，更新updated_at
- 查询时默认排除deleted_at不为空的记录

### 3. 前端UI修改

#### 3.1 显示缓存时间
在页面顶部添加：
```jsx
<div style={{ 
  padding: '0.5rem 1rem', 
  backgroundColor: 'var(--bg-secondary)',
  borderRadius: 'var(--radius-sm)',
  fontSize: '0.875rem',
  color: 'var(--text-secondary)'
}}>
  数据缓存时间: {cacheTimestamp}
</div>
```

#### 3.2 显示域名更新时间
在域名卡片中显示updated_on字段

#### 3.3 删除确认对话框
当刷新发现有域名需要删除时，显示确认对话框：
```
发现以下域名在服务商中已不存在：
- example.com
- test.com

是否将这些域名标记为已删除？
[取消] [确认删除]
```

#### 3.4 自动恢复提示
当刷新发现已删除的域名重新出现时，自动恢复并提示：
```
以下域名已重新激活：
- example.com
```

### 4. 定时任务修改

修改notification_service.go中的GetExpiringDomains方法：
- 查询时添加条件：deleted_at IS NULL
- 只获取未删除的域名

## 实现优先级

1. ✅ 高优先级
   - 显示域名更新时间
   - 显示缓存时间
   - 软删除机制

2. 🔄 中优先级
   - 删除确认对话框
   - 自动恢复提示

3. 📝 低优先级
   - 批量操作优化
   - 性能优化

## 测试场景

1. 正常刷新：所有域名都存在
2. 域名被删除：服务商中删除域名后刷新
3. 域名恢复：重新添加域名后刷新
4. 定时任务：确保不会通知已删除的域名
5. 并发刷新：多个账户同时刷新

## 注意事项

1. 软删除的域名不应该出现在域名列表中
2. 软删除的域名不应该触发到期通知
3. 恢复域名时需要更新所有相关字段
4. 缓存时间应该精确到秒
5. 需要处理时区问题
