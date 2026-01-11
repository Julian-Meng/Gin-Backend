# Backend (Golang)

一个基于 Go 的后端服务项目，提供账号/权限/业务数据接口，并集成 MySQL 与 Ollama（可选）用于 AI 能力。

> 文档较长已拆分，详见 `docs/` 目录。

## Documentation

- 中文文档（部署/运行/配置）：[`docs/README_CN.md`](./docs/README_CN.md)
- English Documentation: [`docs/README_ENG.md`](./docs/README_ENG.md)
- API 文档（接口列表/统一响应/权限要求）：[`docs/API.md`](./docs/API.md)

## Quick Start

### 1) 准备环境
- Go (建议 1.20+)
- Docker Desktop（Windows 建议使用 WSL2 后端）

### 2) 启动依赖服务（MySQL / Ollama）
```bash
docker compose up -d
````

### 3) 启动后端

```bash
go run .
```

## Notes

* 如果后端运行在宿主机（Windows），默认通过 `127.0.0.1` 访问 MySQL / Ollama。
* 配置项示例见：`.env.example`