# DNS 根域名解析优化说明

## 优化内容

### 1. 根域名记录显示优化

#### 后端优化

修改了 `backend/provider/dynu/provider.go` 中的 `ListRecords` 方法：

1. 在获取 DNS 记录列表之前，先获取域名信息
2. 检查域名的 `IPv4Address` 和 `IPv6Address` 字段
3. 如果存在 IPv4 地址，创建一个虚拟的根 A 记录（ID: `root-a`）
4. 如果存在 IPv6 地址，创建一个虚拟的根 AAAA 记录（ID: `root-aaaa`）
5. 将这些根记录与其他子域名记录一起返回

#### 前端优化

1. 在 `frontend/src/locales/` 中添加了提示文本 `rootRecordHint`
2. 在 `frontend/src/pages/Records.jsx` 的主机名输入框下方添加了提示信息，说明留空或输入 `@` 表示根域名记录

### 2. 隐藏 NS 记录

由于 NS（名称服务器）记录通常由 DNS 提供商管理，不支持用户修改，因此进行了以下优化：

#### 后端优化

在 `ListRecords` 方法中添加过滤逻辑，跳过 NS 类型的记录：

```go
// Skip NS records as they are not editable
if r.RecordType == "NS" {
    continue
}
```

#### 前端优化

1. 从 `RECORD_TYPES` 数组中移除 `'NS'`
2. 从翻译文件中移除 NS 记录的说明文本

### 3. 左侧菜单驻留

为左侧导航菜单添加了路由高亮功能：

- 使用 `useLocation` 检测当前路由
- 自动高亮当前激活的菜单项
- 支持子路径匹配
- 添加悬停效果和平滑过渡动画

### 4. DNS 记录页面域名信息展示

在 DNS 记录列表上方添加了域名信息卡片：

#### 后端

- 添加了 `GetDomain` handler 和路由
- 支持获取单个域名的详细信息

#### 前端

- 显示域名名称、IPv4/IPv6 地址、TTL、状态
- 使用响应式网格布局
- 采用玻璃态设计风格

### 5. 亮色/暗色主题切换

添加了完整的主题系统，支持亮色模式、暗色模式和跟随系统：

#### 功能特性

1. **三种主题模式**：
   - 亮色模式（Light）
   - 暗色模式（Dark）
   - 跟随系统（System）

2. **智能切换**：
   - 自动检测系统主题偏好
   - 实时响应系统主题变化
   - 主题设置持久化到 localStorage

3. **UI 组件**：
   - 在顶部导航栏添加主题切换器
   - 三个按钮分别对应三种模式
   - 当前激活的模式高亮显示

#### 技术实现

1. **CSS 变量**：
   - 为亮色和暗色模式定义了完整的颜色变量
   - 使用 `[data-theme="light"]` 选择器切换主题

2. **ThemeContext**：
   - 创建了 `ThemeContext.jsx` 管理主题状态
   - 监听系统主题变化
   - 提供 `useTheme` hook

3. **自动适配**：
   - 所有页面自动适配主题
   - 无需修改现有组件代码

## 技术细节

### 根记录的特殊处理

- **创建/更新**：当 `NodeName` 为空且记录类型为 A 或 AAAA 时，通过更新域名对象（`/dns/{id}` 端点）来设置根记录
- **删除**：当删除 ID 为 `root-a` 或 `root-aaaa` 的记录时，将域名对象的对应 IP 地址字段清空
- **列表显示**：将域名对象的 IP 地址作为虚拟记录显示在列表中

### NS 记录过滤

- 在后端 `ListRecords` 方法中直接过滤掉 NS 记录
- 前端不再显示 NS 记录类型选项
- 用户无法创建、编辑或删除 NS 记录

### 主题系统

- 使用 CSS 自定义属性实现主题切换
- 通过 `data-theme` 属性控制主题
- 支持 `prefers-color-scheme` 媒体查询
- 主题偏好保存在 localStorage

## 测试建议

1. 列出域名的 DNS 记录，验证根 A/AAAA 记录是否正确显示
2. 验证 NS 记录不再显示在列表中
3. 创建新的根域名记录（主机名留空或输入 `@`）
4. 编辑现有的根域名记录
5. 删除根域名记录，验证是否正确清空
6. 确认记录类型选择器中不再有 NS 选项
7. 测试左侧菜单的路由高亮功能
8. 查看 DNS 记录页面的域名信息卡片
9. 切换亮色/暗色主题，验证所有页面正常显示
10. 测试"跟随系统"模式，修改系统主题验证自动切换

## 影响范围

- ✅ 后端：`backend/provider/dynu/provider.go`, `backend/handler/dns_handler.go`, `backend/main.go`
- ✅ 前端：`frontend/src/pages/Records.jsx`, `frontend/src/components/Layout.jsx`, `frontend/src/api.js`
- ✅ 主题：`frontend/src/ThemeContext.jsx`, `frontend/src/App.jsx`, `frontend/src/index.css`
- ✅ 翻译：`frontend/src/locales/en.js`, `frontend/src/locales/zh.js`
- ✅ 编译测试：通过

## 部署说明

1. 重新编译后端：`cd backend && go build`
2. 重新构建前端：`cd frontend && npm run build`
3. 重启服务
