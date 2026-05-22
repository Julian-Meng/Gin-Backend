# Vue Admin Backend（Go）

基于 **Go + Gin + GORM** 的人事管理后端，提供账号、员工、部门、人事变更、公告、考勤、权限矩阵与客服聊天能力（含可选 AI 兜底）。

## 1. 快速启动

### 1.1 路径 A：SQLite 最小启动（推荐本地开发）

1. 安装依赖：

```bash
go mod download
```

2. 复制环境变量模板并按需调整：

```bash
cp .env.example .env
```

3. 保证至少配置了 `JWT_SECRET`（必填），并使用 SQLite：

```env
DB_DRIVER=sqlite
DB_DSN=./db/hr.db
```

4. 启动服务：

```bash
go run .
```

默认地址：`http://localhost:2077`  
Swagger：`http://localhost:2077/swagger/index.html`

### 1.2 路径 B：MySQL + Ollama（docker-compose）

1. 启动依赖服务：

```bash
docker compose up -d
```

2. 在 `.env` 切换为 MySQL DSN（示例）：

```env
DB_DRIVER=mysql
DB_DSN=user:112233@tcp(127.0.0.1:3306)/hrdb?charset=utf8mb4&parseTime=True&loc=Local
```

3. 启动后端：

```bash
go run .
```

---

## 2. 文档入口

- API 文档：[`docs/API.md`](./API.md)
- Swagger：`/swagger/index.html`
- 英文文档：[`docs/README_ENG.md`](./README_ENG.md)
- 环境变量模板：[`../.env.example`](../.env.example)

---

## 3. 环境变量说明

以代码实际读取为准，分为**应用运行层**和**基础设施层**。

## 3.1 应用运行层（Go 服务直接读取）

| 变量 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `GIN_MODE` | 否 | `release` | Gin 运行模式：`debug/release/test` |
| `SERVER_ADDR` | 否 | `:2077` | HTTP 服务监听地址 |
| `SHUTDOWN_TIMEOUT_SECONDS` | 否 | `5` | 优雅关闭超时时间（秒） |
| `JWT_SECRET` | **是** | 无 | JWT 签名密钥；未配置会启动失败 |
| `JWT_ISSUER` | 否 | `gin-backend` | JWT issuer |
| `JWT_EXPIRE_HOURS` | 否 | `24` | JWT 过期时间（小时） |
| `JWT_REFRESH_THRESHOLD_HOURS` | 否 | `6` | 自动续签阈值（小时） |
| `DB_DRIVER` | 否 | `sqlite` | 数据库驱动：`sqlite/mysql` |
| `DB_DSN` | 否 | `./db/hr.db` | 数据库连接串 |
| `DB_DEBUG` | 否 | `false` | GORM debug 日志开关 |
| `SUPERADMIN_ENABLED` | **是** | 无 | superadmin 开关，必须显式配置 `true/false` |
| `SUPERADMIN_USERNAME` | 条件必填 | 无 | `SUPERADMIN_ENABLED=true` 时必填 |
| `SUPERADMIN_PASSWORD` | 条件必填 | 无 | `SUPERADMIN_ENABLED=true` 时必填 |
| `SUPERADMIN_ROLE` | 条件必填 | 无 | `SUPERADMIN_ENABLED=true` 时必填，支持 `superadmin/admin`（不区分大小写） |
| `ACCOUNT_ROLE_CHANGE_REQUIRED_ROLE` | 否 | `superadmin` | 变更账号角色所需权限，支持 `admin/superadmin`（不区分大小写） |
| `LOGIN_CAPTCHA_THRESHOLD` | 否 | `3` | 登录连续失败多少次后启用验证码 |
| `CAPTCHA_TTL_SECONDS` | 否 | `180` | 验证码有效期（秒） |
| `AI_PROVIDER` | 否 | `ollama` | AI 提供方标识 |
| `AI_BASE_URL` | 条件必填 | 无 | 启用 AI 调用时必填（OpenAI 兼容基础地址） |
| `AI_MODEL_ID` | 条件必填 | 无 | 启用 AI 调用时必填 |
| `AI_API_KEY` | 否 | 空 | 某些 provider 需要 |
| `AI_SYSTEM_PROMPT` | 否 | 内置默认 Prompt | AI system prompt |
| `CHAT_AI_FALLBACK_ENABLED` | 否 | `true` | 管理员离线时是否启用 AI 兜底 |
| `CHAT_AI_FALLBACK_MIN_GAP_SECONDS` | 否 | `30` | 同会话 AI 兜底最小触发间隔 |
| `ERROR_DETAIL_ENABLED` | 否 | `false` | 是否在错误响应中暴露 `detail`；`GIN_MODE=debug` 时也会暴露 |

## 3.2 基础设施层（docker-compose 使用）

| 变量 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `MYSQL_ROOT_PASSWORD` | 建议 | 无 | MySQL root 密码 |
| `MYSQL_DATABASE` | 建议 | 无 | 初始化数据库名 |
| `MYSQL_USER` | 建议 | 无 | 业务用户名 |
| `MYSQL_PASSWORD` | 建议 | 无 | 业务用户密码 |
| `MYSQL_PORT` | 否 | `3306` | MySQL 映射端口 |
| `OLLAMA_PORT` | 否 | `11434` | Ollama 映射端口 |

---

## 4. 推荐配置模板

### 4.1 本地最小开发（SQLite）

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

### 4.2 MySQL + Ollama 开发

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

## 5. 常见问题

1. **启动即退出，提示 `JWT_SECRET 未设置`**：补充 `JWT_SECRET`。
2. **启动失败，提示缺少 `SUPERADMIN_ENABLED`**：该项必须显式设置 `true/false`。
3. **AI 调用失败**：检查 `AI_BASE_URL`、`AI_MODEL_ID` 以及对应服务是否可达。
4. **错误响应没看到 detail**：`ERROR_DETAIL_ENABLED=true` 或 `GIN_MODE=debug` 时才会返回 detail。
