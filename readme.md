# Backend (Golang)

一个基于 Go + Gin + GORM 的人事管理后端，提供账号、员工、部门、人事变更、考勤、公告、权限矩阵与聊天能力（含可选 AI）。

## Quick Start

```bash
go mod download
go run .
```

默认监听地址：`http://localhost:2077`  
Swagger：`http://localhost:2077/swagger/index.html`

## Documentation

- 中文部署与配置说明：[`docs/README_CN.md`](./docs/README_CN.md)
- English deployment/config guide: [`docs/README_ENG.md`](./docs/README_ENG.md)
- API 手册（含统一响应与权限说明）：[`docs/API.md`](./docs/API.md)
- 环境变量模板：[`./.env.example`](./.env.example)
- Swagger 生成命令：`swag init -g main.go -o docs`

## TODO

### 近期：

- ~~新增 Swagger 文档~~
- ~~完善错误处理 + 统一响应格式~~
- ~~补充详细 README + 环境变量说明~~
- 添加基础日志 + Request ID

### 中期：

- 增加 Service 层，重构部分逻辑
- 安全加固（密码、JWT、Rate Limit）
- 测试覆盖 + CI

### 长期：

- 更清晰的目录结构 + 依赖注入
- 监控（Prometheus + Grafana）
- 更多自动化（migrations、code generation）
