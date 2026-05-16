# Backend (Golang)

一个基于 Go 的后端服务项目，提供账号/权限/业务数据接口，并集成 MySQL 与 Ollama（可选）用于 AI 能力。

> 文档较长已拆分，详见 `docs/` 目录。

## Documentation

- 中文文档（部署/运行/配置）：[`docs/README_CN.md`](./docs/README_CN.md)
- English Documentation: [`docs/README_ENG.md`](./docs/README_ENG.md)
- API 文档（接口列表/统一响应/权限要求）：[`docs/API.md`](./docs/API.md)
- Swagger 在线文档：`/`（重定向到 `/swagger/index.html`）
- Swagger 重新生成命令：`swag init -g main.go -o docs`

## TODO

- 完善 error handling
- 扩充/重写 README
