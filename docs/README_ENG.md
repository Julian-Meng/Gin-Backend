# Vue Admin Backend (Go)

This project is an HR backend built with **Go + Gin + GORM**, including account, employee, department, personnel-change, notice, attendance, permission, and support-chat modules (with optional AI fallback).

## 1. Quick Start

### 1.1 Path A: SQLite minimal setup (recommended for local development)

1. Install dependencies:

```bash
go mod download
```

2. Copy env template:

```bash
cp .env.example .env
```

3. Make sure `JWT_SECRET` is set, and keep SQLite:

```env
DB_DRIVER=sqlite
DB_DSN=./db/hr.db
```

4. Run server:

```bash
go run .
```

Default URL: `http://localhost:2077`  
Swagger: `http://localhost:2077/swagger/index.html`

### 1.2 Path B: MySQL + Ollama with docker-compose

1. Start dependencies:

```bash
docker compose up -d
```

2. Switch DB settings in `.env` (example):

```env
DB_DRIVER=mysql
DB_DSN=user:112233@tcp(127.0.0.1:3306)/hrdb?charset=utf8mb4&parseTime=True&loc=Local
```

3. Run backend:

```bash
go run .
```

---

## 2. Documentation Entry Points

- API manual: [`docs/API.md`](./API.md)
- Swagger UI: `/swagger/index.html`
- Chinese guide: [`docs/README_CN.md`](./README_CN.md)
- Env template: [`../.env.example`](../.env.example)

---

## 3. Environment Variables

The source of truth is what the code actually reads. Variables are grouped into **application runtime** and **infrastructure**.

## 3.1 Application runtime (read by Go service)

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `GIN_MODE` | No | `release` | Gin mode: `debug/release/test` |
| `SERVER_ADDR` | No | `:2077` | HTTP bind address |
| `SHUTDOWN_TIMEOUT_SECONDS` | No | `5` | Graceful shutdown timeout (seconds) |
| `JWT_SECRET` | **Yes** | None | JWT signing secret; server exits if missing |
| `JWT_ISSUER` | No | `gin-backend` | JWT issuer |
| `JWT_EXPIRE_HOURS` | No | `24` | JWT expiration in hours |
| `JWT_REFRESH_THRESHOLD_HOURS` | No | `6` | Auto-refresh threshold in hours |
| `DB_DRIVER` | No | `sqlite` | DB driver: `sqlite/mysql` |
| `DB_DSN` | No | `./db/hr.db` | Database DSN |
| `DB_DEBUG` | No | `false` | GORM debug logging |
| `SUPERADMIN_ENABLED` | **Yes** | None | Must be explicitly set to `true/false` |
| `SUPERADMIN_USERNAME` | Conditionally | None | Required when `SUPERADMIN_ENABLED=true` |
| `SUPERADMIN_PASSWORD` | Conditionally | None | Required when `SUPERADMIN_ENABLED=true` |
| `SUPERADMIN_ROLE` | Conditionally | None | Required when `SUPERADMIN_ENABLED=true`, supports `superadmin/admin` (case-insensitive) |
| `ACCOUNT_ROLE_CHANGE_REQUIRED_ROLE` | No | `superadmin` | Required role to change account roles; supports `admin/superadmin` (case-insensitive) |
| `LOGIN_CAPTCHA_THRESHOLD` | No | `3` | Captcha required after N failed logins |
| `CAPTCHA_TTL_SECONDS` | No | `180` | Captcha TTL in seconds |
| `AI_PROVIDER` | No | `ollama` | AI provider identifier |
| `AI_BASE_URL` | Conditionally | None | Required when AI is used (OpenAI-compatible base URL) |
| `AI_MODEL_ID` | Conditionally | None | Required when AI is used |
| `AI_API_KEY` | No | Empty | Required by some providers |
| `AI_SYSTEM_PROMPT` | No | Built-in default | AI system prompt |
| `CHAT_AI_FALLBACK_ENABLED` | No | `true` | Enable AI fallback when no admin is online |
| `CHAT_AI_FALLBACK_MIN_GAP_SECONDS` | No | `30` | Min interval between fallback replies in same session |
| `ERROR_DETAIL_ENABLED` | No | `false` | Include `detail` in error response; also enabled in `GIN_MODE=debug` |

## 3.2 Infrastructure (used by docker-compose)

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `MYSQL_ROOT_PASSWORD` | Recommended | None | MySQL root password |
| `MYSQL_DATABASE` | Recommended | None | Initial database name |
| `MYSQL_USER` | Recommended | None | App database user |
| `MYSQL_PASSWORD` | Recommended | None | App database user password |
| `MYSQL_PORT` | No | `3306` | Exposed MySQL port |
| `OLLAMA_PORT` | No | `11434` | Exposed Ollama port |

---

## 4. Recommended Config Presets

### 4.1 Minimal local dev (SQLite)

```env
GIN_MODE=debug
SERVER_ADDR=:2077
JWT_SECRET=HereIsYourStrongSecret
SUPERADMIN_ENABLED=true
SUPERADMIN_USERNAME=root
SUPERADMIN_PASSWORD=123
SUPERADMIN_ROLE=superadmin
ACCOUNT_ROLE_CHANGE_REQUIRED_ROLE=superadmin
DB_DRIVER=sqlite
DB_DSN=./db/hr.db
DB_DEBUG=false
```

### 4.2 MySQL + Ollama dev

```env
DB_DRIVER=mysql
DB_DSN=user:112233@tcp(127.0.0.1:3306)/hrdb?charset=utf8mb4&parseTime=True&loc=Local
AI_PROVIDER=ollama
AI_BASE_URL=http://127.0.0.1:11434/v1
AI_MODEL_ID=qwen3:4b
CHAT_AI_FALLBACK_ENABLED=true
CHAT_AI_FALLBACK_MIN_GAP_SECONDS=30
```

---

## 5. Troubleshooting

1. **Server exits with `JWT_SECRET` missing**: set `JWT_SECRET`.
2. **Server exits with missing `SUPERADMIN_ENABLED`**: this variable must be explicitly set.
3. **AI request failures**: verify `AI_BASE_URL`, `AI_MODEL_ID`, and provider availability.
4. **No `detail` field in error response**: set `ERROR_DETAIL_ENABLED=true` or run in `GIN_MODE=debug`.
