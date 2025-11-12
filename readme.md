
# Vue Admin Backend API 文档

> 作者：CodeGeeX,Julian M,ChatGPT  

## 项目简介

这是一个基于 Gin 框架开发的后端管理系统，提供完整的员工管理、部门管理、人事变更、公告系统等功能。系统采用 JWT 进行身份认证，实现了基于角色的权限控制。

## 技术栈

- Go 1.19+
- Gin Web Framework
- SQLite 数据库
- JWT 认证
- RESTful API 设计

## 项目结构

```
backend/
├── db/
│   └── database.go          # 数据库初始化和连接管理
├── handlers/                # 接口处理层
│   ├── account.go          # 账号相关接口
│   ├── auth.go             # 认证相关接口
│   ├── dashboard.go        # 仪表盘接口
│   ├── department.go       # 部门管理接口
│   ├── notice.go           # 公告管理接口
│   ├── person.go           # 员工管理接口
│   └── personnel.go        # 人事变更接口
├── middlewares/
│   └── jwt.go              # JWT认证中间件
├── models/                 # 数据模型层
│   ├── account.go          # 账号数据模型
│   ├── common.go           # 通用数据库操作
│   ├── dashboard.go        # 仪表盘数据模型
│   ├── department.go       # 部门数据模型
│   ├── notice.go           # 公告数据模型
│   ├── person.go           # 员工数据模型
│   └── personnel.go        # 人事变更数据模型
├── static                   # 静态文件目录
├── main.go                 # 程序入口
└── router.go               # 路由配置
```

## API 接口文档

### 1. 认证接口

#### 登录
```http
POST /api/login
Content-Type: application/json

{
    "username": "string",
    "password": "string"
}
```

#### 注册
```http
POST /api/register
Content-Type: application/json

{
    "username": "string",
    "password": "string",
    "role": "string"
}
```

### 2. 员工管理接口

#### 获取员工列表
```http
GET /api/admin/employees?page=1&pageSize=10&keyword=xxx
Authorization: Bearer <token>
```

#### 获取员工详情
```http
GET /api/admin/employee/:id
Authorization: Bearer <token>
```

#### 创建员工
```http
POST /api/admin/employee
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "string",
    "dptID": "number",
    "job": "string",
    "state": "number"
}
```

#### 更新员工信息
```http
PUT /api/admin/employee/:id
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "string",
    "dptID": "number",
    "job": "string",
    "state": "number"
}
```

#### 删除员工
```http
DELETE /api/admin/employee/:id
Authorization: Bearer <token>
```

### 3. 部门管理接口

#### 获取部门列表
```http
GET /api/admin/departments?page=1&pageSize=10&keyword=xxx
Authorization: Bearer <token>
```

#### 获取部门详情
```http
GET /api/admin/department/:id
Authorization: Bearer <token>
```

#### 创建部门
```http
POST /api/admin/department
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "string",
    "description": "string"
}
```

#### 更新部门信息
```http
PUT /api/admin/department/:id
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "string",
    "description": "string"
}
```

#### 删除部门
```http
DELETE /api/admin/department/:id
Authorization: Bearer <token>
```

### 4. 人事变更接口

#### 获取所有人事变更
```http
GET /api/admin/changes?page=1&pageSize=10
Authorization: Bearer <token>
```

#### 创建人事变更
```http
POST /api/admin/change
Authorization: Bearer <token>
Content-Type: application/json

{
    "empID": "string",
    "type": "string",
    "oldValue": "string",
    "newValue": "string",
    "reason": "string"
}
```

#### 审批人事变更
```http
PUT /api/admin/changes/:id
Authorization: Bearer <token>
Content-Type: application/json

{
    "approve": "boolean",
    "comment": "string"
}
```

### 5. 公告管理接口

#### 获取公告列表
```http
GET /api/notice?page=1&pageSize=10
```

#### 创建公告
```http
POST /api/admin/notice
Authorization: Bearer <token>
Content-Type: application/json

{
    "title": "string",
    "content": "string",
    "type": "string"
}
```

#### 更新公告
```http
PUT /api/admin/notice/:id
Authorization: Bearer <token>
Content-Type: application/json

{
    "title": "string",
    "content": "string",
    "type": "string"
}
```

#### 删除公告
```http
DELETE /api/admin/notice/:id
Authorization: Bearer <token>
```

### 6. 仪表盘接口

#### 管理员仪表盘
```http
GET /api/admin/dashboard
Authorization: Bearer <token>
```

#### 用户仪表盘
```http
GET /api/user/dashboard
Authorization: Bearer <token>
```

## 权限说明

系统采用基于角色的权限控制，包含以下角色：

1. **superadmin**: 超级管理员，拥有所有权限
2. **admin**: 普通管理员，可以管理部门和员工
3. **staff**: 普通员工，只能查看和修改自己的信息

## 数据库结构

主要数据表：

1. **accounts**: 账号表
   - id
   - username
   - password
   - role
   - emp_id

2. **persons**: 员工表
   - emp_id
   - name
   - dpt_id
   - job
   - state

3. **departments**: 部门表
   - id
   - name
   - description
   - count

4. **personnel**: 人事变更表
   - id
   - emp_id
   - type
   - old_value
   - new_value
   - reason
   - status

5. **notices**: 公告表
   - id
   - title
   - content
   - type
   - create_time

## 错误码说明

- 0: 成功
- 1: 业务错误
- 401: 未授权
- 403: 权限不足
- 404: 资源不存在
- 500: 服务器错误

## 开发指南

### 环境要求

- Go 1.19+
- SQLite 3

### 安装步骤

1. 克隆项目
```bash
git clone <repository-url>
cd backend
```

2. 安装依赖
```bash
go mod download
```

3. 初始化数据库
```bash
go run main.go
```

### 开发规范

1. API 接口遵循 RESTful 规范
2. 统一使用 JSON 格式进行数据交换
3. 所有接口都需要进行身份验证（除登录和注册接口）
4. 使用 JWT 进行身份认证
5. 错误处理统一返回格式：
```json
{
    "code": 0,
    "msg": "success",
    "data": {}
}
```

## 部署说明

1. 编译项目
```bash
go build -o vue-admin-backend
```

2. 运行项目
```bash
./vue-admin-backend
```

默认服务端口：8080

## 注意事项

1. 确保数据库文件有正确的读写权限
2. 生产环境请修改 JWT 密钥
3. 建议使用 HTTPS 协议
4. 定期备份数据库文件