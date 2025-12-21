# Vue Admin Backend - README

## 项目概述

这是一个基于 **Go + Gin + GORM** 的后端服务，为 Vue Admin 前端提供现代化的企业人事管理 API。

项目目前已经实现：

* **账户管理（Account）**
* **员工档案管理（Person）**（含账户绑定、档案扩展）
* **部门管理（Department）**
* **人事变动管理（Personnel）**
* **公告管理（Notice）**
* **仪表盘数据（Dashboard）**
* **考勤管理（Attendance）**（签到 / 签退 / 个人查询 / 管理员查询）
* **权限展示接口（Permission）**

并通过 JWT 完成权限认证和角色控制（superadmin / admin / staff）。

---

## 技术栈

* **Go 1.19+**
* **Gin Web Framework**
* **GORM ORM**
* **JWT Authentication**
* **MySQL / SQLite**（自动建表）

---

## 项目结构

```
backend/
├── dao/                    # 数据访问层
├── db/                     # 数据库逻辑层
│   └── database.go         # 数据库初始化、自动迁移
├── handlers/               # HTTP 处理逻辑
├── middlewares/            # 中间件
│   └── jwt.go              # JWT 登录与角色验证
├── models/                 # 数据模型
├── static/                 # 测试页面与静态文件
├──.env                     # 环境配置文件
├── router.go               # 路由注册
└── main.go                 # 程序入口
```

---

## API 接口文档（概要）

**通用规则**：
- 认证方式：大多数接口需要在 Header 中携带 `Authorization: Bearer <token>`（登录后获取）
- `/api/admin/*`：仅管理员（role=admin）可访问
- `/api/user/*`：普通员工（role=staff）可访问
- 请求/响应内容类型：`application/json`
- 成功响应通常包含 `{ code: 0, msg: "success", data: ... }`（具体以实际返回为准）

---

### 1. 登录与注册

| 方法 | 路径                  | 权限     | 请求体 (JSON)                                      | 说明                          |
|------|-----------------------|----------|----------------------------------------------------|-------------------------------|
| POST | `/api/login`          | 公开     | `{ "username": string, "password": string }`       | 登录，返回 `{ token: string }` |
| POST | `/api/register`       | 公开？   | `{ "username": string, "password": string, "role": "staff"/"admin" }` | 注册新账号（需确认是否需要管理员权限） |

---

### 2. Dashboard 概览

| 方法 | 路径                        | 权限   | 请求参数 | 请求体 | 说明                  |
|------|-----------------------------|--------|----------|--------|-----------------------|
| GET  | `/api/admin/dashboard`      | admin  | 无       | 无     | 管理员全局数据概览    |
| GET  | `/api/user/dashboard`       | staff  | 无       | 无     | 普通用户个人数据概览  |

---

### 3. 员工管理 (Person)

| 方法   | 路径                                 | 权限   | 请求参数 / 路径参数          | 请求体 (JSON)                                      | 说明                          |
|--------|--------------------------------------|--------|------------------------------|----------------------------------------------------|-------------------------------|
| GET    | `/api/admin/persons`                 | admin  | 无（分页由后端默认）         | 无                                                 | 获取所有员工列表（分页）      |
| GET    | `/api/admin/person/:id`              | admin  | `:id` (数据库 Person ID)     | 无                                                 | 查询单个员工详情              |
| GET    | `/api/admin/person/profile/:emp_id`  | admin  | `:emp_id` (工号)             | 无                                                 | 管理员查询员工档案            |
| DELETE | `/api/admin/person/:id`              | admin  | `:id` (数据库 ID)            | 无                                                 | 删除员工（按数据库 ID）       |
| DELETE | `/api/admin/person/emp/:emp_id`      | admin  | `:emp_id` (工号)             | 无                                                 | 删除员工（按工号）            |
| POST   | `/api/admin/person`                  | admin  | 无                           | `{ "name": string, "emp_id": string (可选), "dpt_id": number, "job": string }` | 创建新员工                    |
| PUT    | `/api/admin/person/change-dept`       | admin  | 无                           | `{ "emp_id": string, "dept": string (新部门名) }` | 修改员工部门                  |
| PUT    | `/api/admin/person/state`            | admin  | 无                           | `{ "emp_id": string, "state": 0/1 }`               | 修改员工在职状态（0离职，1在职）|
| PUT    | `/api/admin/person/job`              | admin  | 无                           | `{ "emp_id": string, "job": string }`             | 修改员工岗位                  |

---

### 4. 部门管理 (Department)

| 方法   | 路径                                 | 权限   | 查询参数                              | 请求体 (JSON)                               | 说明                          |
|--------|--------------------------------------|--------|---------------------------------------|---------------------------------------------|-------------------------------|
| GET    | `/api/admin/departments`             | admin  | `page`, `pageSize`, `keyword` (可选)  | 无                                          | 部门列表（支持分页+搜索）     |
| GET    | `/api/admin/department/:id`          | admin  | `:id` (部门 ID)                       | 无                                          | 查询单个部门详情              |
| GET    | `/api/user/department/:id`           | staff  | `:id` (部门 ID)                       | 无                                          | 普通用户查询部门（权限受限？）|
| DELETE | `/api/admin/department/:id`          | admin  | `:id` (部门 ID)                       | 无                                          | 删除部门                      |
| POST   | `/api/admin/department`              | admin  | 无                                    | `{ "name": string, "full_num": number (可选，默认20) }` | 创建新部门                    |

---

### 5. 人事变更 (Personnel Change)

| 方法   | 路径                                 | 权限   | 请求体 (JSON)                                                                 | 说明                              |
|--------|--------------------------------------|--------|-------------------------------------------------------------------------------|-----------------------------------|
| GET    | `/api/admin/changes`                 | admin  | 无                                                                            | 获取所有变更记录列表              |
| GET    | `/api/admin/change/:id`              | admin  | `:id` (变更 ID)                                                               | 查询单条变更详情                  |
| POST   | `/api/admin/change`                  | admin  | `{ "emp_id": string, "change_type": 1/2/3, "target_dpt": number (可选), "description": string }` | 管理员直接发起变更（调部门/调岗/离职）|
| PUT    | `/api/admin/change/approve`          | admin  | `{ "id": number, "approver": string, "approve": true/false }`                | 审批变更（通过/驳回）             |
| POST   | `/api/user/change/request`           | staff  | `{ "emp_id": string, "change_type": 1/2/3, "target_dpt": number (可选), "description": string }` | 员工提交变更申请                  |

**change_type 值**：
- 1：调部门
- 2：调岗
- 3：离职

---

### 6. 公告管理 (Notice)

| 方法   | 路径                          | 权限   | 请求体 (JSON)                                      | 说明                     |
|--------|-------------------------------|--------|----------------------------------------------------|--------------------------|
| GET    | `/api/notice`                 | 公开？ | 无                                                 | 获取公开公告列表         |
| GET    | `/api/admin/notice/:id`       | admin  | `:id`                                              | 查询公告详情             |
| DELETE | `/api/admin/notice/:id`       | admin  | `:id`                                              | 删除公告                 |
| POST   | `/api/admin/notice`           | admin  | `{ "title": string, "content": string, "publisher": string }` | 发布新公告               |

---

### 7. 账号管理 (Account)

| 方法   | 路径                          | 权限   | 请求体 (JSON)                                      | 说明                     |
|--------|-------------------------------|--------|----------------------------------------------------|--------------------------|
| GET    | `/api/admin/accounts`         | admin  | 无                                                 | 获取所有账号列表         |
| POST   | `/api/admin/account`          | admin  | `{ "username": string, "password": string, "role": "staff"/"admin" }` | 创建新账号               |
| PUT    | `/api/admin/account/:id`      | admin  | `:id` + `{ "role": "staff"/"admin", "status": 0/1 }` | 更新账号角色或状态       |
| DELETE | `/api/admin/account/:id`      | admin  | `:id`                                              | 删除账号                 |

---

### 8. 个人档案 (Profile)

| 方法   | 路径                              | 权限   | 请求体 (JSON)                                      | 说明                          |
|--------|-----------------------------------|--------|----------------------------------------------------|-------------------------------|
| GET    | `/api/user/profile`               | staff  | 无                                                 | 查看当前登录用户的完整档案    |
| GET    | `/api/user/profile/:person_id`    | staff  | `:person_id` (Person 数据库 ID)                    | 查询指定 Person ID 的档案     |
| PUT    | `/api/user/profile/:person_id`    | staff  | 任意可更新字段（如 `{ "name": "...", "job": "..." }`） | 更新档案（部分字段）          |

---

### 9. 考勤管理 (Attendance)

| 方法   | 路径                                   | 权限   | 请求参数 / 请求体                                  | 说明                          |
|--------|----------------------------------------|--------|----------------------------------------------------|-------------------------------|
| POST   | `/api/user/attendance/checkin`         | staff  | 无（空体）                                         | 上班签到                      |
| POST   | `/api/user/attendance/checkout`        | staff  | 无（空体）                                         | 下班签退                      |
| GET    | `/api/user/attendance/my`              | staff  | Query: `start` (YYYY-MM-DD), `end`, `page`, `pageSize` (可选) | 查询我的考勤记录              |
| GET    | `/api/admin/attendance`                | admin  | Query: `emp_id`, `dpt_id`, `start`, `end` 等       | 管理员查询考勤列表             |
| PUT    | `/api/admin/attendance/:id`            | admin  | `:id` + 可选字段 `{ "status": number, "remark": string, "check_in": ISO string, "check_out": ISO string }` | 修改单条考勤记录              |
| DELETE | `/api/admin/attendance/:id`            | admin  | `:id`                                              | 删除考勤记录                  |

---

### 10. 权限矩阵

| 方法 | 路径                      | 权限   | 请求体 | 说明                        |
|------|---------------------------|--------|--------|-----------------------------|
| GET  | `/api/user/permissions`   | staff  | 无     | 获取当前用户对所有接口的权限状态 |

---

## 权限说明

系统角色：

* **superadmin**：最高权限，所有接口可访问
* **admin**：管理员，可管理人事、部门、账号、考勤
* **staff**：普通员工，可查看自己的信息、提交申请、个人打卡

权限控制方式：

* JWT 中写入 role
* `/api/admin/**` 由 `AdminOnly()` 中间件保护
* `/api/user/**` 任何登录用户可访问
* 权限展示接口 `/api/user/permissions` 返回每个接口可访问角色明细

---

## JWT 认证方式

所有受保护接口需携带 Header：

```
Authorization: Bearer <token>
```

Token 中包含：

```json
{
  "username": "user1",
  "emp_id": "EMP0002",
  "role": "staff"
}
```

---

## 开发环境配置

### 1. 依赖安装

```bash
go mod download
```

### 2. 数据库配置

在 `db/database.go` 配置 MySQL 或 SQLite，启动时自动建表。

### 3. 运行项目

```bash
go run .
```

---

## 注意事项

1. 密码全部采用 bcrypt 加密
2. JWT 密钥建议使用环境变量
3. 所有数据表由 GORM 自动迁移创建
4. 考勤表建议在数据库中为 `(emp_id, date)` 添加唯一索引
5. 权限展示无需数据库，直接由后端维护静态权限表即可

---

## 常见问题

### 1. 如何修改 JWT 密钥？

编辑 `middlewares/jwt.go` 内的 `getSecret()` 函数。

### 2. 如何新增业务模块？

创建 `models/xxx.go` → `dao/xxx.go` → `handlers/xxx.go` → 在 `router.go` 注册路由。

### 3. 如何初始化 SQLite？

确保 `db/` 目录存在，或在 `database.go` 内配置 SQLite 文件路径。
