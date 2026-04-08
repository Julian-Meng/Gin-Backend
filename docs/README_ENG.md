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
- Support chat (user-admin sessions, offline queue, admin claim, optional AI fallback)

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

## Support Chat Module (New)

- Base path: `/api/chat/*`
- Realtime channel: `GET /api/chat/ws?token=<jwt>`
- User send message: `POST /api/chat/user/message`
- Admin sessions list: `GET /api/chat/admin/sessions`
- Admin claim waiting sessions: `POST /api/chat/admin/sessions/claim`
- Session messages: `GET /api/chat/messages/:id`

Behavior summary:

- First assignment uses least-load-first, then random among candidates.
- Session stays sticky to one admin after assignment.
- If no admin is online, messages are queued and dispatched after admin claim.
- Optional AI fallback can reply while admins are offline; all messages are persisted for admin review.

## Chat AI Fallback Env Vars

Add in `.env`:

- `CHAT_AI_FALLBACK_ENABLED=true`
- `CHAT_AI_FALLBACK_MIN_GAP_SECONDS=30`

## Test Console

- Open `/bt` for the built-in backend test page.
- The Support Chat tab now supports WS connect/disconnect, session refresh, claim actions, history loading, and message sending in both user/admin modes.