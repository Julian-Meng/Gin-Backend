# Vue Admin Backend (Go)

A backend service built with **Go + Gin + GORM**, providing a modern HR management API for a Vue Admin frontend.

It includes **JWT authentication** and role-based access control (**superadmin / admin / staff**), and optionally integrates with **Ollama** for AI chat features.

## Features

- Account management
- Employee profile management (Person)
- Department management
- Personnel change workflow
- Notice management
- Dashboard stats
- Attendance (check-in / check-out / queries)
- Permission matrix API
- AI chat (optional)

## Tech Stack

- Go 1.19+
- Gin
- GORM
- JWT
- MySQL / SQLite (auto-migration)

## Quick Start

### Install dependencies
```bash
go mod download
````

### Start services (MySQL / Ollama)

```bash
docker compose up -d
```

### Run backend

```bash
go run .
```

## Auth & Roles

Most protected APIs require:

```text
Authorization: Bearer <token>
```

Roles:

* **superadmin**: full access
* **admin**: management APIs
* **staff**: employee APIs

For detailed API list: [`docs/API.md`](./API.md)