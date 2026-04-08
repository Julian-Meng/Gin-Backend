# Vue Admin Backend (Go)

这是一个基于 **Go + Gin + GORM** 的后端服务，为 Vue Admin 前端提供企业人事管理 API，并支持 JWT 鉴权与角色权限控制（superadmin / admin / staff）。

## 功能概览

已实现模块：

- 账户管理（Account）
- 员工档案管理（Person）（含账户绑定、档案扩展）
- 部门管理（Department）
- 人事变动管理（Personnel）
- 公告管理（Notice）
- 仪表盘数据（Dashboard）
- 考勤管理（Attendance）（签到 / 签退 / 个人查询 / 管理员查询）
- 权限展示接口（Permission）
- AI Chat（对接 Ollama，可选）
- 客服聊天（Support Chat：用户-管理员会话、离线队列、管理员认领、AI 兜底）

## 技术栈

- Go 1.19+
- Gin Web Framework
- GORM ORM
- JWT Authentication
- MySQL / SQLite（自动建表）

## 项目结构

```text
backend/
├── dao/                    # 数据访问层
├── db/                     # 数据库逻辑层
│   └── database.go         # 数据库初始化、自动迁移
├── handlers/               # HTTP 处理逻辑
├── middlewares/            # 中间件
│   └── jwt.go              # JWT 登录与角色验证
├── models/                 # 数据模型
├── static/                 # 测试页面与静态文件
├── tmp/                    # 临时文件
├── .env.example            # 环境变量示例
├── router.go               # 路由注册
└── main.go                 # 程序入口
````

## 快速启动（开发环境）

### 1) 安装依赖

```bash
go mod download
```

### 2) 启动依赖服务（MySQL / Ollama）

如果你使用 docker-compose（推荐开发环境）：

```bash
docker compose up -d
```

> 如果你的 Go 服务跑在 Windows 宿主机上，默认使用 `127.0.0.1` 连接 MySQL / Ollama。

### 3) 启动后端

```bash
go run .
```

## 认证与权限（简版说明）

大多数受保护接口都需要 Header：

```text
Authorization: Bearer <token>
```

角色说明：

* superadmin：最高权限，所有接口可访问
* admin：管理员（人事/部门/账号/考勤）
* staff：普通员工（个人信息/申请/打卡等）

路由权限约定：

* `/api/admin/**`：仅 admin（以及 superadmin）可访问
* `/api/user/**`：任何已登录用户可访问

详细接口说明见：[`docs/API.md`](./API.md)

## 聊天模块说明（新增）

- 入口：`/api/chat/*`
- 实时通道：`GET /api/chat/ws?token=<jwt>`（浏览器场景推荐 query token）
- 用户发消息：`POST /api/chat/user/message`
- 管理员会话：`GET /api/chat/admin/sessions`
- 管理员认领：`POST /api/chat/admin/sessions/claim`
- 会话历史：`GET /api/chat/messages/:id`

模块行为：

- 首次分配按低负载优先再随机，后续会话粘性绑定管理员。
- 管理员离线时消息入库排队，管理员上线后认领并补发。
- 可选 AI 离线兜底，回复会落库，管理员上线可追溯。

## 聊天 AI 兜底环境变量

在 `.env` 中新增：

- `CHAT_AI_FALLBACK_ENABLED=true`
- `CHAT_AI_FALLBACK_MIN_GAP_SECONDS=30`

说明：

- `CHAT_AI_FALLBACK_ENABLED`：是否启用管理员离线时的 AI 自动回复。
- `CHAT_AI_FALLBACK_MIN_GAP_SECONDS`：同一会话最小 AI 触发间隔，避免高频重复回复。

## 联调页面

- 打开：`/bt`
- 页面中的“客服聊天”页签已支持：WS 连接、会话列表、认领会话、拉取历史、用户/管理员发送消息。

## Superadmin（必填配置）

项目包含用于演示/紧急情况的 superadmin 账号，必须通过环境变量配置，禁止硬编码。

在 `.env` 中设置：

* `SUPERADMIN_USERNAME`
* `SUPERADMIN_PASSWORD`
* `SUPERADMIN_ROLE`（默认 superadmin）
* `SUPERADMIN_ENABLED`（true/false）

注意：若启用 `SUPERADMIN_ENABLED=true` 但未设置用户名或密码，服务将拒绝启动。