# TodoApp — Task Management System

A production-ready REST API for task management with AI-powered priority classification, built with Go and PostgreSQL.

## Overview

TodoApp is a backend application that helps users manage tasks intelligently. The system automatically calculates task priority using a weighted scoring algorithm based on deadline urgency, importance, stagnation detection, and dependencies.

**Key Features:**
- Full CRUD for users and tasks
- AI Priority Engine — automatic priority scoring with human-readable explanations
- Stagnation Detection — identifies important tasks that haven't been updated in 5+ days
- Statistics & Analytics — completion rates, overdue tasks, average completion time
- Domain-Based Automation — category-aware priority behavior
- Web UI — built-in frontend dashboard served by the Go server
- Structured logging with request tracing
- Graceful shutdown

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.22 |
| Database | PostgreSQL 18 (Docker) |
| DB Driver | pgx/v5 |
| Logger | go.uber.org/zap |
| Config | kelseyhightower/envconfig |
| Migrations | golang-migrate |
| Container | Docker + Docker Compose |

---

## Architecture

The project follows **Clean Architecture** with strict layer separation:

```
cmd/
  todoapp/
    main.go                          # Entry point

internal/
  core/
    domain/                          # Domain entities: User, Task, Nullable
    errors/                          # Sentinel errors
    logger/                          # Structured zap logger
    repository/postgres/conn/        # PostgreSQL connection pool
    transport/http/
      middleware/                    # RequestID, Logger, Panic, Trace
      response/                      # HTTP response helpers
      server/                        # HTTPServer, APIVersionRouter

  features/
    users/
      transport/http/                # HTTP handlers
      service/                       # Business logic
      repository/postgres/           # SQL queries
    tasks/
      transport/http/                # HTTP handlers
      service/                       # Business logic
      repository/postgres/           # SQL queries
      priority/                      # AI Priority Engine
    statistics/
      transport/http/                # HTTP handlers
      service/                       # Business logic
      repository/postgres/           # SQL queries

migrations/                          # SQL migration files
web/
  index.html                         # Frontend dashboard
```

**Request flow:**
```
HTTP Request
  → RequestID middleware  (assign unique request ID)
  → Logger middleware     (inject logger into context)
  → Panic middleware      (recover from panics)
  → Trace middleware      (log latency and status)
  → Handler              (decode, validate, call service)
  → Service              (business logic, validation)
  → Repository           (SQL query)
  → PostgreSQL
```

---

## AI Priority Engine

Each task receives a `priority_score` (0.0–1.0) calculated by:

```
score = 0.35 × deadline + 0.30 × importance + 0.20 × dependencies + 0.15 × stagnation
```

| Factor | Weight | Description |
|--------|--------|-------------|
| deadline | 35% | Overdue=1.0, <1 day=0.9, <3 days=0.7, <7 days=0.5, <14 days=0.3 |
| importance | 30% | Normalized from 1–5 scale |
| dependencies | 20% | Placeholder 0.5 (future feature) |
| stagnation | 15% | Not updated 7+ days=1.0, 3+ days=0.5 |

**Result:**
- `score >= 0.6` → `priority_level: "high"`
- `score < 0.6` → `priority_level: "normal"`
- `priority_reason` — human-readable explanation in Russian

Weight justification: 78% of survey respondents (N=50) named deadline as the primary factor.

---

## Database Schema

```sql
-- Users table
CREATE TABLE todoapp.users (
    id              BIGSERIAL PRIMARY KEY,
    version         INT NOT NULL DEFAULT 1,
    full_name       VARCHAR(100) NOT NULL,
    phone_number    VARCHAR(15),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tasks table
CREATE TABLE todoapp.tasks (
    id              BIGSERIAL PRIMARY KEY,
    version         INT NOT NULL DEFAULT 1,
    user_id         BIGINT NOT NULL REFERENCES todoapp.users(id) ON DELETE CASCADE,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    deadline        TIMESTAMPTZ,
    importance      INT NOT NULL DEFAULT 1,      -- 1 to 5
    category        TEXT NOT NULL DEFAULT 'personal', -- work / study / personal
    completed       BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at    TIMESTAMPTZ,
    priority_score  FLOAT,
    priority_level  TEXT,                        -- high / normal
    priority_reason TEXT,
    last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop) with WSL2 integration enabled
- Go 1.22+
- make

---

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/aaaaarsen/golang-todoapp.git
cd golang-todoapp
```

### 2. Configure environment

Copy the example env file and fill in your values:

```bash
cp .env.example .env
```

Required variables in `.env`:

```env
# HTTP Server
HTTP_ADDR=:5050
HTTP_SHUTDOWN_TIMEOUT=30s

# PostgreSQL
POSTGRES_USER=test-user-123
POSTGRES_PASSWORD=test-postgres-password-456
POSTGRES_DB=test-db
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
POSTGRES_TIMEOUT=10s

# Logger
LOGGER_LEVEL=DEBUG
LOGGER_FOLDER=./out/logs
```

### 3. Start PostgreSQL

```bash
make env-up
```

### 4. Run migrations

```bash
make migrate-up
```

### 5. Forward PostgreSQL port (WSL/Docker Desktop)

```bash
make env-port-forward
```

### 6. Run the application

```bash
make todoapp-run
```

The server starts at `http://localhost:5050`.

Open the web dashboard at `http://localhost:5050` in your browser.

---

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make env-up` | Start PostgreSQL container |
| `make env-down` | Stop PostgreSQL container |
| `make env-cleanup` | Remove PostgreSQL data (with confirmation) |
| `make env-port-forward` | Forward PostgreSQL port via socat |
| `make env-port-close` | Stop port forwarder |
| `make migrate-up` | Apply all migrations |
| `make migrate-down` | Roll back last migration |
| `make migrate-create seq=<name>` | Create new migration files |
| `make todoapp-run` | Run the application |

---

## API Reference

Base URL: `http://localhost:5050/api/v1`

### Users

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/users` | Create user |
| `GET` | `/users` | List users (paginated) |
| `DELETE` | `/users/{id}` | Delete user |

**POST /users**
```json
{
  "full_name": "Arsen Raymbek",
  "phone_number": "+77001234567"
}
```

**GET /users**
```
?page=1&page_size=20
```

---

### Tasks

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/tasks` | Create task (auto-calculates priority) |
| `GET` | `/tasks` | List tasks with filters |
| `GET` | `/tasks/{id}` | Get task by ID |
| `PATCH` | `/tasks/{id}` | Update task (recalculates priority) |
| `DELETE` | `/tasks/{id}` | Delete task |
| `GET` | `/tasks/stagnant` | Get stagnant tasks |

**POST /tasks**
```json
{
  "user_id": 1,
  "title": "Finish SIS Project",
  "description": "Complete the university project",
  "deadline": "2026-05-10T00:00:00Z",
  "importance": 5,
  "category": "study"
}
```

**Response:**
```json
{
  "id": 1,
  "version": 1,
  "user_id": 1,
  "title": "Finish SIS Project",
  "importance": 5,
  "category": "study",
  "completed": false,
  "priority_score": 0.87,
  "priority_level": "high",
  "priority_reason": "Высокий приоритет: дедлайн через 4 дней",
  "created_at": "2026-05-05T16:00:00Z"
}
```

**GET /tasks filters:**
```
?user_id=1&category=study&completed=false&page=1&page_size=10
```

**PATCH /tasks/{id}**
```json
{
  "version": 1,
  "completed": true,
  "importance": 4
}
```

**GET /tasks/stagnant** — returns tasks with `importance >= 3` not updated in 5+ days:
```
?user_id=1
```

---

### Statistics

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/statistics` | Get task statistics |

**GET /statistics**
```
?user_id=1&from=2026-01-01T00:00:00Z&to=2026-12-31T00:00:00Z&category=study
```

**Response:**
```json
{
  "total_tasks": 10,
  "completed_tasks": 7,
  "completion_rate": 0.70,
  "overdue_tasks": 2,
  "high_priority_tasks": 3,
  "avg_completion_time_hours": 24.5
}
```

---

## Validation Rules

**Users:**
- `full_name`: 3–100 characters, required
- `phone_number`: optional, must start with `+`, 10–15 characters

**Tasks:**
- `title`: 1–255 characters, required
- `importance`: 1–5, default 1
- `category`: `work`, `study`, or `personal`, default `personal`
- `deadline`: optional, RFC3339 format

---

## Project Structure Notes

**Concurrency & Database:**
- Uses optimistic locking (`version` field) on PATCH operations — returns 409 Conflict on version mismatch
- All timestamps stored and returned as UTC
- PostgreSQL schema `todoapp` separates app tables from system tables

**Middleware Chain:**
```
RequestID → Logger → Panic Recovery → Trace → Handler
```
Every request gets a unique `X-Request-ID` header. The logger middleware injects a request-scoped logger with `request_id` and `url` fields into the context, making all downstream logs traceable.

---

## Logs

Application logs are written to both stdout and `./out/logs/` directory.

Log format:
```
2026-05-05T16:00:10.363222    DEBUG    middleware/common.go:82    >>> Incoming HTTP Request    {"request_id": "816da061", "url": "/api/v1/tasks"}
2026-05-05T16:00:10.412164    ERROR    http/create_task.go:121    create task error    {"request_id": "816da061", "error": "..."}
```

---

## Research Background

This project was built as a SIS (Student Independent Study) project at KBTU. The feature set was derived from:

- **3 qualitative interviews** — identified "decision paralysis" as the main pain point
- **Survey of N=50 users** — Likert scale 1–5, Pearson correlation analysis
  - M=3.84 — users want automation
  - r=0.593 — override capability is the strongest predictor of adoption
  - 78% chose deadline-based prioritization as the most important feature
- **Model evaluation on 50 tasks** — Precision = Recall = F1 = 89.47%

---

## License

MIT
