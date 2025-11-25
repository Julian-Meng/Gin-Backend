
# Vue Admin Backend - README

## 项目概述
这是一个基于 Go + Gin + GORM 的后端项目，为 Vue Admin 管理系统提供 API 支持。项目实现了完整的权限管理、人事管理、部门管理等功能。

## 技术栈
- Go 1.19+
- Gin Web Framework
- GORM (ORM框架)
- JWT (身份验证)
- MySQL/SQLite (数据库)

## 项目结构
```
backend/
├── dao/          # 数据访问层
├── db/           # 数据库配置
├── handlers/     # HTTP处理函数
├── middlewares/  # 中间件
├── models/       # 数据模型
├── static/       # 静态文件
├── main.go       # 入口文件
└── router.go     # 路由配置
```

## API 接口文档

### 1. 认证相关 (/auth)
- `POST /auth/login` - 用户登录
  ```json
  {
    "username": "string",
    "password": "string"
  }
  ```
- `POST /auth/register` - 用户注册
  ```json
  {
    "username": "string",
    "password": "string",
    "role": "string"
  }
  ```

### 2. 仪表盘 (/dashboard)
- `GET /dashboard/admin` - 获取管理员仪表盘数据
- `GET /dashboard/user` - 获取普通用户仪表盘数据

### 3. 账户管理 (/accounts)
- `GET /accounts` - 获取账户列表
- `POST /accounts` - 创建新账户
- `PUT /accounts/:id` - 更新账户信息
- `DELETE /accounts/:id` - 删除账户

### 4. 部门管理 (/departments)
- `GET /departments` - 获取部门列表
- `GET /departments/:id` - 获取部门详情
- `POST /departments` - 创建新部门
- `PUT /departments/:id` - 更新部门信息
- `DELETE /departments/:id` - 删除部门

### 5. 员工管理 (/persons)
- `GET /persons` - 获取员工列表
- `GET /persons/:id` - 获取员工详情
- `POST /persons` - 创建新员工
- `PUT /persons/:id` - 更新员工信息
- `DELETE /persons/:id` - 删除员工
- `PUT /persons/:id/department` - 调整员工部门
- `PUT /persons/:id/state` - 更改员工状态
- `PUT /persons/:id/job` - 更改员工职位

### 6. 人事变动 (/personnel)
- `GET /personnel` - 获取人事变动列表
- `GET /personnel/:id` - 获取人事变动详情
- `POST /personnel` - 提交人事变动申请
- `PUT /personnel/:id/approve` - 审批人事变动

### 7. 公告管理 (/notices)
- `GET /notices` - 获取公告列表
- `GET /notices/:id` - 获取公告详情
- `POST /notices` - 创建新公告
- `PUT /notices/:id` - 更新公告
- `DELETE /notices/:id` - 删除公告

## 数据模型

### Account (账户)
```go
type Account struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Password  string    `json:"-"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### Department (部门)
```go
type Department struct {
    ID          uint   `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Count       int    `json:"count"`
}
```

### Person (员工)
```go
type Person struct {
    ID         uint      `json:"id"`
    EmpID      string    `json:"emp_id"`
    Name       string    `json:"name"`
    Department string    `json:"department"`
    Job        string    `json:"job"`
    State      int       `json:"state"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```

### Personnel (人事变动)
```go
type Personnel struct {
    ID         uint      `json:"id"`
    EmpID      string    `json:"emp_id"`
    Type       string    `json:"type"`
    Content    string    `json:"content"`
    Status     int       `json:"status"`
    Approver   string    `json:"approver"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```

### Notice (公告)
```go
type Notice struct {
    ID         uint      `json:"id"`
    Title      string    `json:"title"`
    Content    string    `json:"content"`
    Publisher  string    `json:"publisher"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```

## 权限说明
- `superadmin`: 超级管理员，拥有所有权限
- `admin`: 普通管理员，拥有部门管理权限
- `user`: 普通用户，只能查看个人信息

## JWT 认证
所有需要认证的接口都需要在请求头中携带 JWT token：
```
Authorization: Bearer <token>
```

## 开发环境配置

### 1. 环境要求
- Go 1.19+
- MySQL 5.7+ 或 SQLite 3+

### 2. 安装依赖
```bash
go mod download
```

### 3. 配置数据库
在 `db/database.go` 中配置数据库连接信息：
```go
type Config struct {
    Driver   string // "mysql" 或 "sqlite"
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SQLitePath string // SQLite 数据库文件路径
}
```

### 4. 运行项目
```bash
go run .
```

## 注意事项
1. 所有密码都使用 bcrypt 加密存储
2. 敏感信息（如 JWT 密钥）建议使用环境变量配置
3. 数据库表会在项目启动时自动创建
4. 建议在生产环境中使用 MySQL 而不是 SQLite

## 常见问题
1. 如何修改 JWT 密钥？
   - 在 `middlewares/jwt.go` 中修改 `getSecret()` 函数

2. 如何添加新的 API 接口？
   - 在 `handlers` 目录下创建新的处理函数
   - 在 `router.go` 中注册路由

3. 如何修改数据库表结构？
   - 在 `models` 目录下修改对应的结构体
   - 项目启动时会自动执行数据库迁移