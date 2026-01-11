# API 文档

## 通用规则

- 认证方式：大多数接口需要在 Header 中携带 `Authorization: Bearer <token>`（登录后获取）
- `/api/admin/*`：仅管理员（role=admin）可访问（superadmin 也可访问）
- `/api/user/*`：普通员工（role=staff）可访问（superadmin 也可访问）
- 请求/响应内容类型：`application/json`

## 统一响应格式

所有接口统一返回 JSON：

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
````

字段说明：

* `code`：业务状态码，0 表示成功
* `msg`：提示信息
* `data`：返回数据内容（不同接口含义不同）

---

## 目录

1. 登录与注册
2. Dashboard 概览
3. 员工管理（Person）
4. 部门管理（Department）
5. 人事变更（Personnel Change）
6. 公告管理（Notice）
7. 账号管理（Account）
8. 个人档案（Profile）
9. 考勤管理（Attendance）
10. 权限矩阵（Permissions）
11. AI聊天（AIChat）

---

## 1. 登录与注册

| 方法   | 路径              | 权限  | 请求体 (JSON)                                                            | 说明                        |
| ---- | --------------- | --- | --------------------------------------------------------------------- | ------------------------- |
| POST | `/api/login`    | 公开  | `{ "username": string, "password": string }`                          | 登录，返回 `{ token: string }` |
| POST | `/api/register` | 公开？ | `{ "username": string, "password": string, "role": "staff"/"admin" }` | 注册新账号（需确认是否需要管理员权限）       |

---

## 2. Dashboard 概览

| 方法  | 路径                     | 权限    | 请求参数 | 请求体 | 说明         |
| --- | ---------------------- | ----- | ---- | --- | ---------- |
| GET | `/api/admin/dashboard` | admin | 无    | 无   | 管理员全局数据概览  |
| GET | `/api/user/dashboard`  | staff | 无    | 无   | 普通用户个人数据概览 |

---

## 3. 员工管理 (Person)

| 方法     | 路径                                  | 权限    | 请求参数 / 路径参数           | 请求体 (JSON)                                                                   | 说明                |
| ------ | ----------------------------------- | ----- | --------------------- | ---------------------------------------------------------------------------- | ----------------- |
| GET    | `/api/admin/persons`                | admin | 无（分页由后端默认）            | 无                                                                            | 获取所有员工列表（分页）      |
| GET    | `/api/admin/person/:id`             | admin | `:id` (数据库 Person ID) | 无                                                                            | 查询单个员工详情          |
| GET    | `/api/admin/person/profile/:emp_id` | admin | `:emp_id` (工号)        | 无                                                                            | 管理员查询员工档案         |
| DELETE | `/api/admin/person/:id`             | admin | `:id` (数据库 ID)        | 无                                                                            | 删除员工（按数据库 ID）     |
| DELETE | `/api/admin/person/emp/:emp_id`     | admin | `:emp_id` (工号)        | 无                                                                            | 删除员工（按工号）         |
| POST   | `/api/admin/person`                 | admin | 无                     | `{ "name": string, "emp_id": string (可选), "dpt_id": number, "job": string }` | 创建新员工             |
| PUT    | `/api/admin/person/change-dept`     | admin | 无                     | `{ "emp_id": string, "dept": string (新部门名) }`                                | 修改员工部门            |
| PUT    | `/api/admin/person/state`           | admin | 无                     | `{ "emp_id": string, "state": 0/1 }`                                         | 修改员工在职状态（0离职，1在职） |
| PUT    | `/api/admin/person/job`             | admin | 无                     | `{ "emp_id": string, "job": string }`                                        | 修改员工岗位            |

---

## 4. 部门管理 (Department)

| 方法     | 路径                          | 权限    | 查询参数                               | 请求体 (JSON)                                         | 说明              |
| ------ | --------------------------- | ----- | ---------------------------------- | -------------------------------------------------- | --------------- |
| GET    | `/api/admin/departments`    | admin | `page`, `pageSize`, `keyword` (可选) | 无                                                  | 部门列表（支持分页+搜索）   |
| GET    | `/api/admin/department/:id` | admin | `:id` (部门 ID)                      | 无                                                  | 查询单个部门详情        |
| GET    | `/api/user/department/:id`  | staff | `:id` (部门 ID)                      | 无                                                  | 普通用户查询部门（权限受限？） |
| DELETE | `/api/admin/department/:id` | admin | `:id` (部门 ID)                      | 无                                                  | 删除部门            |
| POST   | `/api/admin/department`     | admin | 无                                  | `{ "name": string, "full_num": number (可选，默认20) }` | 创建新部门           |

---

## 5. 人事变更 (Personnel Change)

| 方法   | 路径                          | 权限    | 请求体 (JSON)                                                                                     | 说明          |
| ---- | --------------------------- | ----- | ---------------------------------------------------------------------------------------------- | ----------- |
| GET  | `/api/admin/changes`        | admin | 无                                                                                              | 获取所有变更记录列表  |
| GET  | `/api/admin/change/:id`     | admin | `:id` (变更 ID)                                                                                  | 查询单条变更详情    |
| POST | `/api/admin/change`         | admin | `{ "emp_id": string, "change_type": 1/2/3, "target_dpt": number (可选), "description": string }` | 管理员直接发起变更   |
| PUT  | `/api/admin/change/approve` | admin | `{ "id": number, "approver": string, "approve": true/false }`                                  | 审批变更（通过/驳回） |
| POST | `/api/user/change/request`  | staff | `{ "emp_id": string, "change_type": 1/2/3, "target_dpt": number (可选), "description": string }` | 员工提交变更申请    |

change_type 值：

* 1：调部门
* 2：调岗
* 3：离职

---

## 6. 公告管理 (Notice)

| 方法     | 路径                      | 权限    | 请求体 (JSON)                                                    | 说明       |
| ------ | ----------------------- | ----- | ------------------------------------------------------------- | -------- |
| GET    | `/api/notice`           | 公开？   | 无                                                             | 获取公开公告列表 |
| GET    | `/api/admin/notice/:id` | admin | `:id`                                                         | 查询公告详情   |
| DELETE | `/api/admin/notice/:id` | admin | `:id`                                                         | 删除公告     |
| POST   | `/api/admin/notice`     | admin | `{ "title": string, "content": string, "publisher": string }` | 发布新公告    |

---

## 7. 账号管理 (Account)

| 方法     | 路径                       | 权限    | 请求体 (JSON)                                                            | 说明        |
| ------ | ------------------------ | ----- | --------------------------------------------------------------------- | --------- |
| GET    | `/api/admin/accounts`    | admin | 无                                                                     | 获取所有账号列表  |
| POST   | `/api/admin/account`     | admin | `{ "username": string, "password": string, "role": "staff"/"admin" }` | 创建新账号     |
| PUT    | `/api/admin/account/:id` | admin | `:id` + `{ "role": "staff"/"admin", "status": 0/1 }`                  | 更新账号角色或状态 |
| DELETE | `/api/admin/account/:id` | admin | `:id`                                                                 | 删除账号      |

---

## 8. 个人档案 (Profile)

| 方法  | 路径                             | 权限    | 请求体 (JSON)                                   | 说明            |                    |
| --- | ------------------------------ | ----- | -------------------------------------------- | ------------- | ------------------ |
| GET | `/api/user/profile`            | staff | 无                                            | 查看当前登录用户的完整档案 |                    |
| GET | `/api/user/profile/:person_id` | staff | `:person_id` (Person 数据库 ID)                 | 无             | 查询指定 Person ID 的档案 |
| PUT | `/api/user/profile/:person_id` | staff | 任意可更新字段（如 `{ "name": "...", "job": "..." }`） | 更新档案（部分字段）    |                    |

---

## 9. 考勤管理 (Attendance)

| 方法     | 路径                              | 权限    | 请求参数 / 请求体                                                                                             | 说明        |
| ------ | ------------------------------- | ----- | ------------------------------------------------------------------------------------------------------ | --------- |
| POST   | `/api/user/attendance/checkin`  | staff | 无（空体）                                                                                                  | 上班签到      |
| POST   | `/api/user/attendance/checkout` | staff | 无（空体）                                                                                                  | 下班签退      |
| GET    | `/api/user/attendance/my`       | staff | Query: `start` (YYYY-MM-DD), `end`, `page`, `pageSize` (可选)                                            | 查询我的考勤记录  |
| GET    | `/api/admin/attendance`         | admin | Query: `emp_id`, `dpt_id`, `start`, `end` 等                                                            | 管理员查询考勤列表 |
| PUT    | `/api/admin/attendance/:id`     | admin | `:id` + 可选字段 `{ "status": number, "remark": string, "check_in": ISO string, "check_out": ISO string }` | 修改单条考勤记录  |
| DELETE | `/api/admin/attendance/:id`     | admin | `:id`                                                                                                  | 删除考勤记录    |

---

## 10. 权限矩阵 (Permissions)

| 方法  | 路径                      | 权限    | 请求体 | 说明               |
| --- | ----------------------- | ----- | --- | ---------------- |
| GET | `/api/user/permissions` | staff | 无   | 获取当前用户对所有接口的权限状态 |

---

## 11. AI聊天 (AIChat)

| 方法   | 路径          | 权限    | 请求体                                                     | 说明              |
| ---- | ----------- | ----- | ------------------------------------------------------- | --------------- |
| POST | `/api/chat` | staff | `{ "message": "Message", "session_id": "web-session" }` | 返回结果中包含 AI 回复内容 |
