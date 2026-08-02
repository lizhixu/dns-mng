# Agent.md

本文件用于整理当前项目需求、架构约定与维护注意事项，方便后续由开发者或 AI Agent 继续维护。

## 项目概览

- 项目类型：前后端分离的 DNS 管理系统。
- 后端：Go 1.24、Gin、JWT、SQLite/libSQL。
- 前端：React 19、React Router v7、Vite 7。
- 默认部署形态：Docker Compose，前端 Nginx 反向代理 `/api` 到后端。
- 核心用途：统一管理多个 DNS 服务商账号、域名、DNS 记录，并提供域名续期提醒、DDNS、ACME DNS-01、Cloudflare 优选、DNSHE 管理、WHOIS 查询等能力。

## 目录结构

- `backend/`
  - Go 后端服务。
  - `main.go`：注册服务商、初始化数据库/服务/路由、启动定时任务。
  - `config/`：环境变量配置。
  - `database/`：数据库初始化、表结构、迁移。
  - `models/`：请求/响应/数据库模型。
  - `handler/`：Gin HTTP handler。
  - `service/`：业务逻辑层。
  - `provider/`：各 DNS 服务商适配器。
- `frontend/`
  - React 前端。
  - `src/App.jsx`：路由入口。
  - `src/api.js`：API 封装。
  - `src/pages/`：页面组件。
  - `src/components/`：通用组件和布局。
  - `src/locales/`：中英文文案。
- `doc/`
  - 各服务商/API 的补充文档。
- 根目录文档
  - `README.md`、`DDNS_API.md`、`DOCKER_DEPLOYMENT.md`、`DOMAIN_REFRESH_IMPLEMENTATION.md`、`TROUBLESHOOTING.md` 等。

## 运行与部署需求

### 后端环境变量

- `SERVER_PORT`：服务端口，默认 `8080`。
- `JWT_SECRET`：JWT 密钥，生产环境必须修改，默认值不安全。
- `DB_TYPE`：数据库类型，`sqlite` 或 `libsql`，默认 `sqlite`。
- `DB_PATH`：SQLite 文件路径，默认 `dns-mng.db`，Docker 中通常为 `/data/dns-mng.db`。
- `DB_URL`、`DB_AUTH_TOKEN`：libSQL/Turso 使用。

### Docker 部署

- 根目录 `docker-compose.yaml`
  - 同时启动 backend 和 frontend。
  - backend 暴露 `8080`。
  - frontend 暴露 `80`。
  - 数据卷 `dns-data` 挂载到 `/data`。
- `backend/docker-compose.yaml`
  - 仅后端部署配置，适合 Dokploy 等场景。
  - 支持 `DB_TYPE=libsql`、`DB_URL`、`DB_AUTH_TOKEN`。

### 前端代理与生产 Nginx

- `frontend/vite.config.js` 将开发环境 `/api` 代理到 `http://localhost:8080`。
- `frontend/nginx.conf` 将生产环境 `/api` 代理到 Docker 网络中的 `backend:8080`，并配置 SPA fallback。

## 认证与用户行为

- 登录使用 JWT，token 有效期 7 天。
- 首次登录不存在用户时会自动注册用户。
- 显式注册接口：`POST /api/auth/register`。
- ACME Basic Auth 不会自动创建用户，必须先通过系统登录创建账号。
- 密码使用 bcrypt 存储。
- 敏感信息包括但不限于：服务商 API key、SMTP 密码、DDNS token、WHOIS API key、备份内容。维护时不要写入日志，不要在错误信息中泄露。

## 后端 API 与功能模块

### 服务商管理

服务商统一实现 `DNSProvider` 接口：

- `ListDomains`
- `GetDomain`
- `ListRecords`
- `CreateRecord`
- `UpdateRecord`
- `DeleteRecord`

当前已注册服务商及相关文档：

| 服务商 | Provider ID | 相关文档 |
| --- | --- | --- |
| Dynu | `dynu` | [`doc/DYNU_OPTIMIZATION.md`](doc/DYNU_OPTIMIZATION.md) |
| 腾讯云 DNSPod | `tencentcloud` | [`doc/TENCENTCLOUD_GUIDE.md`](doc/TENCENTCLOUD_GUIDE.md)、[`doc/TESTING_TENCENTCLOUD.md`](doc/TESTING_TENCENTCLOUD.md) |
| 阿里云 | `aliyun` | 暂无专用文档，维护时参考 `backend/provider/aliyun/` 实现 |
| 华为云 | `huaweicloud` | 暂无专用文档，维护时参考 `backend/provider/huaweicloud/` 实现 |
| Cloudflare | `cloudflare` | 暂无专用供应商文档，Cloudflare 优选逻辑参考 `backend/service/cf_optimize_service.go` |
| NDJP NET | `ndjp` | [`doc/NDJP.json`](doc/NDJP.json) |
| deSEC | `desec` | 暂无专用文档，维护时参考 `backend/provider/desec/` 实现 |
| DNSHE | `dnshe` | [`doc/dnshe_API.md`](doc/dnshe_API.md) |
| Hurricane Electric | `hurricane` | [`doc/hurricane.md`](doc/hurricane.md) |
| IPv64 | `ipv64` | [`doc/IPv64.md`](doc/IPv64.md)、[`doc/ipv64_README.md`](doc/ipv64_README.md) |
| VPS8 | `vps8` | [`doc/VPS8 DNS OpenAPI.md`](doc/VPS8%20DNS%20OpenAPI.md) |

供应商文档索引见 [`doc/README.md`](doc/README.md)。

Cloudflare 维护注意：`backend/provider/cloudflare/client.go` 的 `ListZones` 必须处理 `/zones` 分页，避免只拉取第一页导致“所有域名刷新”误判后续页域名不存在并提示隐藏本地缓存。

新增服务商时：

1. 在 `backend/provider/<name>` 下实现 provider。
2. 确保实现 `backend/provider/provider.go` 的 `DNSProvider` 接口。
3. 在 `backend/main.go` 中 `provider.Register(...)`。
4. 同步更新前端账号创建页面和中英文文案。

### 账号与 DNS 管理

账号 CRUD：

- `GET /api/accounts`
- `POST /api/accounts`
- `PUT /api/accounts/:id`
- `DELETE /api/accounts/:id`

域名/记录：

- `GET /api/domains`
- `GET /api/domains/refresh`
- `GET /api/accounts/:id/domains`
- `GET /api/accounts/:id/domains/refresh`
- `GET /api/accounts/:id/domains/:domainId/records`
- `POST /api/accounts/:id/domains/:domainId/records`
- `PUT /api/accounts/:id/domains/:domainId/records/:recordId`
- `DELETE /api/accounts/:id/domains/:domainId/records/:recordId`

维护要求：

- `DNSService` 优先读取域名缓存；缓存为空或刷新时访问服务商。
- 多账号刷新使用 goroutine 并发拉取。
- 单个账号刷新失败不应阻塞整体结果。

### 域名缓存、续期信息与软删除

`domain_cache` 保存：

- 域名 ID、域名名称、账号 ID。
- 手动维护的续期日期 `renewal_date`。
- 续费链接 `renewal_url`。
- 软删除标记 `deleted_at`。
- `last_sync_at`、`provider_updated_on`。
- `uses_dnshe_dns`：标记 DNSHE 域名是否仍使用 DNSHE 自身解析。

刷新逻辑：

- 如果服务商中不再存在某域名，刷新接口会返回 `domains_to_delete`，前端提示用户确认软删除。
- 已软删除域名不显示在域名列表，也不参与到期通知。
- 如果软删除域名重新出现在服务商数据中，需要支持自动恢复。
- 当前已移除 `renewal_manual` 锁定字段；服务商返回空续期信息时应保留缓存值。

### DDNS

公开 DuckDNS 兼容接口：

- `GET /api/ddns/update`

参数：

- `domains`：必填，逗号分隔。
- `token`：必填，用户级 token。
- `ip`：可选 IPv4，不传则使用客户端 IP。
- `ipv6`：可选 IPv6。

Token 管理：

- `GET /api/ddns-token`
- `PUT /api/ddns-token`
- `DELETE /api/ddns-token`

行为要求：

- 每个用户一个 DDNS token。
- token 可更新该用户所有账号下的所有域名。
- 只自动更新已有 A/AAAA 记录，不创建新记录。
- 返回 DuckDNS 风格纯文本：成功 `OK`，失败 `KO`。

### ACME DNS-01

对外接口使用 HTTP Basic Auth，账号密码为系统用户。

路由：

- `POST /api/acme/dns01/present`
- `POST /api/acme/dns01/cleanup`

用途与要求：

- 用于 lego 或脚本自动签发证书时创建/清理 TXT 记录。
- ACME Basic Auth 使用 `VerifyCredentials`，不会自动注册用户。

### 到期通知与邮件

需求：

- 每个域名可配置提前通知天数和是否启用。
- 邮件配置为用户级 SMTP 配置。
- 定时任务每天 09:00 执行：
  - 域名到期提醒。
  - DNSHE 自动续期。
- 定时任务日志写入 `scheduler_logs`。

到期通知应跳过：

- 未启用通知的域名。
- 未设置续期日期的域名。
- 已软删除的域名。
- 已过期域名。
- 当天已经通知过的域名。

注意：当前 `scheduler_service.go` 实现是计算下一个 09:00 并等待，后端启动不会立即执行定时检查；维护文档时需保持一致。

### 备份与恢复

路由：

- `POST /api/backup/export`：前端使用的导出接口，请求体 `{ "password": "..." }`，避免通过 query 传递备份密码。
- `GET /api/backup/export`：兼容旧用法。
- `POST /api/backup/import`

导出内容包括：

- 账号。
- 域名缓存、续期信息、软删除状态、同步时间与通知设置。
- DDNS token。
- 邮件配置。
- WHOIS 配置。
- DNSHE 自动续期配置。
- CF 优选配置。

要求：

- 支持可选密码加密备份；明文备份允许导出，但前端必须强提醒其包含敏感信息。
- 导入支持 `overwrite` 控制覆盖或跳过。
- `/api/backup/export` 与 `/api/backup/import` 不写入 `api_call_logs`，避免备份内容、备份密码、API key、SMTP 密码、DDNS token、WHOIS API key 等敏感信息落库。
- 备份中包含敏感信息，下载、保存、日志处理要谨慎。

### Cloudflare 优选

页面：`/cf-optimize`

后端服务：`CFOptimizeService`

目标：一键配置 Cloudflare for SaaS/Custom Hostnames 相关 DNS 记录与回源。

要求：

- 账号必须是 Cloudflare provider。
- API Token 需要 DNS 编辑权限，以及 SSL/证书/Custom Hostnames 相关权限。
- 账户需开通 Cloudflare for SaaS。

会创建/更新：

- origin A 记录。
- intermediate CNAME。
- 业务 CNAME。
- custom hostname。
- 可能的验证记录。

维护注意：部分失败时有回滚新建记录逻辑；修改此模块时要格外注意清理/回滚路径。

### DNSHE 管理与自动续期

页面：`/dnshe`

功能：

- DNSHE 账号列表。
- 额度查询。
- 注册子域名。
- 删除子域名。
- 手动/自动续期。
- 配置域名是否使用 DNSHE 自身解析。
- 一键解析到 Cloudflare。

自动续期：

- 用户级配置：`enabled`、`days_before`、`last_run_at`。
- 定时任务每天 09:00 运行所有启用用户。
- 手动触发接口也存在。
- 永久域名或空续期日期会跳过。
- 当前注释明确“续期不检查额度”。

DNSHE 第三方解析域名处理：

- `uses_dnshe_dns=false` 的 DNSHE 域名在“所有域名”中会被过滤，避免与第三方服务商账号重复。
- DNSHE 缓存中的续期日期/续费链接可用于回填同名第三方域名。

### WHOIS 查询

页面：`/whois`

后端路由：

- `GET /api/whois/config`
- `PUT /api/whois/config`
- `GET /api/whois/query?domain=...`

第三方服务：WhoisJSON.com

- Endpoint：`https://whoisjson.com/api/v1/whois`
- Authorization header：`TOKEN=<api_key>`

需求：

- API key 为用户级配置。
- 已移除 WHOIS enable/disable 开关；当前只要配置 API key 即可查询。
- 后端需兼容 WhoisJSON.com 返回字段类型不稳定的情况，例如 string、数组、bool、数字等。

注意：handler 注释与 service 注释存在轻微不一致：handler 注释说 `GetConfig` 返回 key cleared，service 实际说明和代码会返回明文 API key 给所属用户。后续维护应统一安全语义和注释。

### 日志

- API 调用日志：
  - 中间件 `APILogger` 记录已认证 API 请求。
  - 未登录/公开接口请求没有有效用户 ID，不写入 `api_call_logs`，避免 `user_id=0` 触发外键约束失败。
  - 表：`api_call_logs`。
- 登录日志：
  - 表：`login_logs`。
  - 字段包括 IP、UA、设备、状态、IP 地理位置等。
- 定时任务日志：
  - 表：`scheduler_logs`。
- 前端页面：`/logs`。
- SQLite 模式下数据库连接限制为单连接，主要用于避免异步 API 日志写入和页面读日志时出现 `database locked`。

## 数据库维护注意事项

- 数据库表在 `backend/database/database.go` 中通过 `CREATE TABLE IF NOT EXISTS` 和若干 `ALTER TABLE ADD COLUMN` 自动初始化/迁移。
- 已存在列的 `ALTER TABLE ADD COLUMN` 报错会被忽略。
- 其他建表/迁移错误会 `log.Fatalf` 终止服务。
- SQLite DSN 使用 WAL 和 busy timeout：`?_journal_mode=WAL&_busy_timeout=5000`。
- SQLite 模式：
  - `SetMaxOpenConns(1)`
  - `SetMaxIdleConns(1)`
- libSQL 模式：
  - 最大打开连接数 10。
  - 最大空闲连接数 5。
- `.env.example` 提到 Windows host build 会 fallback 到 sqlite；Turso/libSQL 建议用 Linux/Docker 镜像运行。
- 仓库中存在 `dns-mng.db`、`backend/dns-mng.db` 和若干 `.exe` 构建产物，应避免误提交新的二进制或数据库文件。

## 前端页面与路由

主要路由：

- `/login`
- `/domains`
- `/dnshe`
- `/accounts`
- `/accounts/:accountId/domains`
- `/accounts/:accountId/domains/:domainId/records`
- `/profile`
- `/logs`
- `/email-settings`
- `/backup`
- `/cf-optimize`
- `/whois`

前端维护要求：

- `Layout.jsx` 管理侧边栏、主题切换、语言切换、用户设置、密码修改、移动端适配。
- 语言支持中文/英文，新增功能需同步更新：
  - `frontend/src/locales/zh.js`
  - `frontend/src/locales/en.js`
- 维护多语言文案时，需要检查是否已有可复用 key，避免新增重复语义或未使用的 locale 字段。
- 主题支持亮色、暗色、跟随系统。
- 多个页面使用 `fetchedRef` 防止 React StrictMode 下重复请求。
- 账户下域名管理页 `/accounts/:accountId/domains` 的顶部返回按钮应返回浏览器上一页，而不是固定跳转账户列表。
- DNS 记录页 `/accounts/:accountId/domains/:domainId/records` 的顶部返回按钮应返回浏览器上一页，而不是固定跳转域名列表。
- 新增页面需考虑移动端布局、暗色主题、401 自动跳转登录逻辑。

## 最近功能/变更线索

- `c2d40e3 Remove WHOIS enable/disable toggle`
  - 移除 WHOIS 启用/禁用开关。
  - 涉及 `backend/database/database.go`、`backend/models/whois.go`、`backend/service/whois_service.go`、前端 WHOIS 页和中英文文案。
- `4a91353 Add WHOIS lookup feature with per-user API key config`
  - 新增 WHOIS 查询功能、用户级 API key 配置、前端页面和路由。
- `8ca6c36 Remove renewal_manual lock flag from domain cache`
  - 移除续期信息手动锁定字段。
- `2d5dd76 fix: don't use VipEndAt as domain renewal date`
  - 修正腾讯云域名续期日期字段来源。
- 更早提交集中在 DNSHE：
  - DNSHE 管理。
  - 自动续期。
  - Cloudflare 解析。
  - 第三方解析域名过滤。
  - 续期日期从 DNSHE 缓存回填。

## 通用开发维护约定

修改后端路由时，检查三处是否同步：

1. `backend/main.go`
2. `frontend/src/api.js`
3. 对应页面/文案

修改数据模型时，检查：

1. `backend/models/`
2. `backend/database/database.go`
3. service/handler 读写逻辑
4. 前端字段名

接口与安全约定：

- 新增需要登录的接口必须挂在 `protected` group 并使用 JWT middleware。
- 公开接口仅保留必要能力：
  - `/api/providers`
  - `/api/auth/*`
  - `/api/ddns/update`
  - `/api/acme/*`，但 ACME 有 Basic Auth。
- 所有敏感值不要输出到日志，包括：
  - provider API key
  - JWT
  - DDNS token
  - SMTP password
  - WHOIS API key
  - 备份明文内容

服务商适配注意：

- 涉及服务商 API 的改动要注意各 provider 的字段差异：
  - 域名 ID。
  - 根域/子域格式。
  - TTL 下限。
  - 记录启用状态。
  - TXT 值引号。
  - 续期日期字段。

DNSHE 特殊逻辑：

- 维护所有域名列表、缓存、续期时要保留：
  - `uses_dnshe_dns=false` 过滤。
  - DNSHE 续期信息向第三方同名域名回填。
  - 第三方解析域名不触发误删除。

SQLite 并发：

- SQLite 并发写入敏感。
- 新增异步写日志、批量更新或定时任务时，需要考虑锁竞争。

发布：

- Docker 发布由 GitHub Actions 构建并推送：
  - `jacyli/dns-mng:backend`
  - `jacyli/dns-mng:frontend`
  - tag 版本会额外推送 `backend-<version>`、`frontend-<version>`。

## 建议后续维护方式

- 新增需求时，优先补充到本文件对应模块。
- 后续如有业务逻辑、接口行为、数据结构、定时任务、缓存/同步策略、权限安全策略等逻辑改动，必须同步维护本文件，确保 Agent.md 与实际代码行为一致。
- 如果新增服务商、页面、数据库字段或公开接口，必须同时更新“维护约定”中提到的关联文件。
- 如果实际代码行为与历史文档不一致，以代码为准，并同步修正文档。
- 对可能造成外部变更的功能，例如删除 DNS 记录、Cloudflare 优选回滚、DNSHE 删除/续期，修改前应先确认交互和失败恢复路径。

## 本次整理参考的关键文件

- `README.md`
- `DDNS_API.md`
- `DOMAIN_REFRESH_IMPLEMENTATION.md`
- `backend/NOTIFICATION_GUIDE.md`
- `docker-compose.yaml`
- `backend/docker-compose.yaml`
- `backend/.env.example`
- `backend/go.mod`
- `backend/main.go`
- `backend/config/config.go`
- `backend/database/database.go`
- `backend/provider/provider.go`
- `backend/provider/registry.go`
- `backend/models/domain.go`
- `backend/models/account.go`
- `backend/service/user_service.go`
- `backend/service/dns_service.go`
- `backend/service/scheduler_service.go`
- `backend/service/dnshe_auto_renew_service.go`
- `backend/service/whois_service.go`
- `backend/service/backup_service.go`
- `backend/service/cf_optimize_service.go`
- `backend/handler/ddns_handler.go`
- `backend/handler/whois_handler.go`
- `frontend/package.json`
- `frontend/vite.config.js`
- `frontend/nginx.conf`
- `frontend/src/App.jsx`
- `frontend/src/api.js`
- `frontend/src/components/Layout.jsx`
- `frontend/src/pages/AllDomains.jsx`
- `frontend/src/pages/DNSHE.jsx`
- `.github/workflows/docker-publish.yml`
