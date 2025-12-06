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
├── dao/              # 数据访问层
│   ├── account.go
│   ├── attendance.go       # 新增：考勤管理 DAO
│   ├── dashboard.go
│   ├── department.go
│   ├── notice.go
│   ├── person.go
│   └── personnel.go
│
├── db/
│   └── database.go         # 数据库初始化、自动迁移
│
├── handlers/               # HTTP 处理逻辑
│   ├── account.go
│   ├── attendance.go       # 新增：考勤管理接口
│   ├── auth.go
│   ├── dashboard.go
│   ├── department.go
│   ├── notice.go
│   ├── permission.go       # 新增：权限展示接口
│   ├── person.go
│   └── personnel.go
│
├── middlewares/
│   └── jwt.go              # JWT 登录与角色验证
│
├── models/                 # 数据模型
│   ├── account.go
│   ├── attendance.go       # 新增：考勤模型
│   ├── dashboard.go
│   ├── department.go
│   ├── notice.go
│   ├── person.go
│   └── personnel.go
│
├── static/                 # 测试页面与静态文件
│   └── attendance_test.html（可选）
│
├── router.go               # 路由注册
└── main.go                 # 程序入口
```

---

## API 接口文档（概要）

以下为主要模块的 API 概览，完整参数请参考 handlers 文件。

---

### 1. 认证相关 (/api/login, /api/register)

#### POST /api/login

用户登录，返回 JWT token。

#### POST /api/register

注册新账户（仅管理员场景使用）。

---

### 2. 仪表盘

#### GET /api/admin/dashboard

管理员仪表盘数据。

#### GET /api/user/dashboard

普通用户仪表盘数据。

---

### 3. 账户管理 (/api/admin/account)

* GET /api/admin/accounts
* POST /api/admin/account
* PUT /api/admin/account/:id
* DELETE /api/admin/account/:id

---

### 4. 员工管理 (/api/admin/person)

* GET /api/admin/persons
* POST /api/admin/person
* GET /api/admin/person/:id
* PUT /api/admin/person/:id
* DELETE /api/admin/person/:id
* DELETE /api/admin/person/emp/:emp_id
* PUT /api/admin/person/job
* PUT /api/admin/person/state
* PUT /api/admin/person/change-dept

---

### 5. 员工档案扩展（新功能）

#### GET /api/user/profile

查看当前登录用户完整档案（Account + Person + Department 信息）。

#### GET /api/admin/person/profile/:emp_id

管理员查看任意员工档案。

---

### 6. 部门管理 (/api/admin/department)

* GET /api/admin/departments
* GET /api/admin/department/:id
* POST /api/admin/department
* PUT /api/admin/department/:id
* DELETE /api/admin/department/:id

普通用户也能查看：

* GET /api/user/department/:id

---

### 7. 人事变动 (/api/admin/change, /api/user/change/request)

管理员接口：

* GET /api/admin/changes
* GET /api/admin/change/:id
* POST /api/admin/change
* PUT /api/admin/change/approve

普通用户提交申请：

* POST /api/user/change/request

---

### 8. 公告管理 (/api/admin/notice)

管理员：

* POST /api/admin/notice
* PUT /api/admin/notice/:id
* DELETE /api/admin/notice/:id
* GET /api/admin/notice/:id

公开接口：

* GET /api/notice

---

### 9. 考勤管理（新增模块）

#### 用户接口 (/api/user/attendance)

* POST /api/user/attendance/checkin —— 签到
* POST /api/user/attendance/checkout —— 签退
* GET  /api/user/attendance/my —— 查询自己的考勤记录（支持日期区间 + 分页）

#### 管理员接口 (/api/admin/attendance)

* GET    /api/admin/attendance —— 条件查询（emp_id / dpt_id / 日期区间）
* PUT    /api/admin/attendance/:id —— 修正考勤（状态/备注/时间）
* DELETE /api/admin/attendance/:id —— 删除记录

---

### 10. 权限展示接口（新增模块）

#### GET /api/user/permissions

返回系统全部权限项 + 当前用户是否拥有（HasAccess）。

前端可用此接口自动构建“权限说明”页面，无需手动维护权限表。

---

## 数据模型（新增部分）

### Attendance (考勤)

```go
type Attendance struct {
    ID        uint
    EmpID     string     // 关联 Person.EmpID
    Date      time.Time  // 考勤日期
    CheckIn   *time.Time // 签到
    CheckOut  *time.Time // 签退
    Status    int        // 0缺勤 1正常 2迟到 3早退 4迟到且早退
    Remark    string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

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
