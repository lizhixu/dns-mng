# DNS Manager

[中文](README.md) | [English](README_EN.md)

A modern, multi-provider DNS record management system with a clean web UI. Manage domains and records across 11 DNS providers from a single interface.

## Features

- 🌐 **Multi-provider support** — Cloudflare, Tencent Cloud DNSPod, Alibaba Cloud DNS, Huawei Cloud DNS, Dynu, NDJP NET, deSEC, Hurricane Electric, IPv64, DNSHE, VPS8
- 🔐 **JWT authentication** — secure login with auto-registration on first use
- 🔄 **DDNS** — DuckDNS-compatible dynamic DNS API for routers and clients
- 🔒 **ACME DNS-01** — HTTP Basic Auth endpoints for automated SSL/TLS certificate issuance
- 📧 **Domain expiry notifications** — scheduled daily email alerts for domains approaching renewal
- 💾 **Backup & restore** — JSON export/import with optional AES encryption
- 📝 **Logging** — API call logs, login logs with IP geolocation, scheduler task logs
- 🎨 **Modern UI** — clean interface with light / dark / system theme
- 🌍 **i18n** — Chinese and English
- 📱 **Responsive** — works on all screen sizes
- 🐳 **Docker** — one-command deployment

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.24+, Gin, SQLite / libSQL, JWT |
| Frontend | React 19, React Router v7, Vite 7, Lucide Icons |
| Deploy | Docker Compose, Nginx, GitHub Actions CI/CD |

## Quick Start

### Docker (Recommended)

```bash
git clone https://github.com/lizhixu/dns-mng.git
cd dns-mng
chmod +x start.sh
./start.sh
```

- **Frontend**: http://localhost
- **Backend API**: http://localhost:8080

On first visit, enter any username and password — the account is created automatically.

### Manual Setup

**Backend:**

```bash
cd backend
go mod download
go run main.go
```

**Frontend:**

```bash
cd frontend
npm install
npm run dev        # development
npm run build      # production
```

## Configuration

Create a `.env` file in the `backend/` directory (see `backend/.env.example`):

```bash
# JWT secret (MUST change in production)
JWT_SECRET=your-secure-random-secret-key

# Database: "sqlite" (default) or "libsql" (Turso / remote)
DB_TYPE=sqlite
DB_PATH=dns-mng.db

# For Turso / remote libSQL:
# DB_URL=libsql://your-db.turso.io
# DB_AUTH_TOKEN=your-turso-auth-token

# Server port
SERVER_PORT=8080
```

### Using Turso

Set `DB_TYPE=libsql` and provide your Turso connection URL and auth token. Tables are created automatically on startup.

> **Note:** The libSQL driver requires CGO and is only available on linux/darwin (amd64/arm64). Use the Docker image to connect to Turso from Windows.

## Usage

### 1. Login

Enter any username and password. If the account doesn't exist, it is created automatically.

### 2. Add a DNS Provider Account

Go to **Accounts** and add your DNS provider credentials:

| Provider | Auth Format |
|----------|-------------|
| Cloudflare | API Token (recommended) or Global API Key |
| Tencent Cloud DNSPod | `SecretId,SecretKey` |
| Alibaba Cloud DNS | `AccessKeyId,AccessKeySecret` |
| Huawei Cloud DNS | `AccessKeyId,SecretAccessKey[,ProjectId][,RegionId]` |
| Dynu | API Key |
| NDJP NET | Bearer Token |
| deSEC | Token |
| Hurricane Electric | `email,password` |
| IPv64 | API Key |
| DNSHE | `APIKey,APISecret` |
| VPS8 | API Key |

### 3. Manage Domains & Records

- **All Domains** — view domains across all accounts, with search and filters
- **Records** — create, update, delete DNS records (A, AAAA, CNAME, MX, TXT, SPF, SRV)

## APIs

### ACME DNS-01

For automated SSL/TLS certificate issuance (compatible with `lego`, Certbot hooks, etc.).

Auth: **HTTP Basic Auth** (same credentials as the web login).

> The account must have been created by logging in at least once — the ACME API does not auto-create accounts.

```bash
# Present challenge
curl -u "user:pass" \
  -H "Content-Type: application/json" \
  -d '{"fqdn":"_acme-challenge.example.com.","value":"txt-value","ttl":300}' \
  http://localhost:8080/api/acme/dns01/present

# Cleanup challenge
curl -u "user:pass" \
  -H "Content-Type: application/json" \
  -d '{"fqdn":"_acme-challenge.example.com.","value":"txt-value"}' \
  http://localhost:8080/api/acme/dns01/cleanup
```

### DDNS (Dynamic DNS)

DuckDNS-compatible API for routers and dynamic IP clients.

1. Go to **DDNS Settings** in the web UI and create a token.
2. Use the token to update records:

```bash
# Auto-detect client IP
curl "http://localhost:8080/api/ddns/update?domains=example.com&token=your-token"

# Specify IP
curl "http://localhost:8080/api/ddns/update?domains=example.com&token=your-token&ip=1.2.3.4"

# Multiple domains
curl "http://localhost:8080/api/ddns/update?domains=example.com,sub.example.com&token=your-token"
```

**Router configuration** — select DuckDNS or custom URL:

```
Domain: your-domain-name
Token:  your-ddns-token
Update URL: https://your-domain.com/api/ddns/update?domains=%s&token=%s&ip=%s
```

See [DDNS_API.md](DDNS_API.md) for full API documentation.

### Domain Expiry Notifications

The scheduler runs daily at 9:00 AM (and once on startup) to check for domains approaching expiration. Configure per-domain notification settings and SMTP email in the web UI.

To change the schedule, edit `backend/service/scheduler_service.go`:

```go
nextRun := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
// Change 9 to your preferred hour (0-23)
```

## Project Structure

```
.
├── backend/                 # Go backend
│   ├── config/              # Environment config
│   ├── database/            # SQLite / libSQL setup
│   ├── handler/             # HTTP handlers
│   ├── middleware/           # Auth, CORS, logging
│   ├── models/              # Data models
│   ├── provider/            # DNS provider implementations
│   │   ├── cloudflare/
│   │   ├── aliyun/
│   │   ├── huaweicloud/
│   │   ├── tencentcloud/
│   │   ├── dynu/
│   │   ├── ndjp/
│   │   ├── desec/
│   │   ├── hurricane/
│   │   ├── ipv64/
│   │   ├── dnshe/
│   │   └── vps8/
│   └── service/             # Business logic
├── frontend/                # React frontend
│   └── src/
│       ├── components/      # Shared components
│       ├── pages/           # Page views
│       └── locales/         # i18n translations
├── docker-compose.yaml      # Docker Compose config
├── start.sh / stop.sh       # Convenience scripts
└── .github/workflows/       # CI/CD
```

## Adding a New DNS Provider

See [PROVIDER_GUIDE.md](PROVIDER_GUIDE.md) for a step-by-step guide.

## Contributing

Contributions are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -m 'Add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

## Changelog

### v0.0.2 (2026-04-23)

- ✨ Added 3 DNS providers: Hurricane Electric, IPv64, DNSHE
- 🔄 Added DDNS (DuckDNS-compatible dynamic DNS)
- ⬆️ Upgraded React to v19, React Router to v7, Vite to v7
- 📝 Improved documentation

### v0.0.1 (2026-03-27)

- ✨ Initial release with 5 DNS providers
- 🎨 Modern UI with theme switching
- 🌍 Chinese / English i18n
- 📝 API call & scheduler logging
- 🐳 Docker one-click deployment

## Links

- GitHub: https://github.com/lizhixu/dns-mng
- Docker Hub: https://hub.docker.com/r/jacyli/dns-mng

## License

[MIT](LICENSE)
