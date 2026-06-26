# DNS OpenAPI

Use your client API key to manage DNS records via OpenAPI.

##### Base URL

```
https://vps8.zz.cd/api/client/dnsopenapi/*
```

##### Authentication

HTTP Basic auth: username=client, password=YOUR_API_KEY

- **username:** `client`
- **password:** `YOUR_API_KEY`

##### How to get API key

1. 登录客户区域
2. Open Account / Profile settings
3. Find API key section and generate/copy your key

If your account is suspended or API access is disabled by admin, calls will be rejected.

Rate limit is enabled. Excess requests return HTTP 429.

##### Supported record types

- A
- AAAA
- MX
- CNAME
- TXT

##### Endpoints

- `POST /api/client/dnsopenapi/domain_list` — List DNS zones
- `POST /api/client/dnsopenapi/record_list` — List records for a domain
- `POST /api/client/dnsopenapi/record_create` — Create record
- `POST /api/client/dnsopenapi/record_update` — Update record
- `POST /api/client/dnsopenapi/record_delete` — Delete record

##### Example: domain_list

```bash
curl -sS -u "client:YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -X POST "https://vps8.zz.cd/api/client/dnsopenapi/domain_list" \
  -d '{}'
```

```json
Response body: {"result":[{"domain":"example.com","platform_type":"self_platform","source_service":"domain","created_at":"2026-04-08 10:53:37","expires_at":"2027-04-08 10:53:38"}],"error":null}
```

##### Example: record_list

```bash
curl -sS -u "client:YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -X POST "https://vps8.zz.cd/api/client/dnsopenapi/record_list" \
  -d '{"domain":"example.com"}'
```

```json
Response body: { "result": [{ "id": 8801, "host": "@", "type": "A", "value": "154.36.187.27", "ttl": 300, "priority": 0, "provider_record_id": "69ef0e417e6a3" }], "error": null }
```



##### Example: record_create

```bash
curl -sS -u "client:YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -X POST "https://vps8.zz.cd/api/client/dnsopenapi/record_create" \
  -d '{"domain":"example.com","host":"www","type":"A","value":"1.2.3.4","ttl":600}'
```

##### Example: record_update

```bash
curl -sS -u "client:YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -X POST "https://vps8.zz.cd/api/client/dnsopenapi/record_update" \
  -d '{"domain":"example.com","id":12345,"value":"5.6.7.8","ttl":600}'
```

##### Example: record_delete

```bash
curl -sS -u "client:YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -X POST "https://vps8.zz.cd/api/client/dnsopenapi/record_delete" \
  -d '{"domain":"example.com","id":12345}'
```