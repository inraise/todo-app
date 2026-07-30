# TODO-APP

Backend-сервис для управления задачами (TODO): пользователи, задачи и статистика по ним. Go + PostgreSQL, HTTP JSON API.

## Стек

Go 1.26 · PostgreSQL 17 (`pgx/v5`) · `net/http` (`http.ServeMux`, без фреймворка) · `zap` (логи) · `go-playground/validator` (валидация) · `golang-migrate` (миграции) · `envconfig` (конфиг из env) · `swaggo` (Swagger-документация) · Docker / Docker Compose

## Архитектура

Feature-based структура: каждая фича (`users`, `tasks`, `statistics`) содержит слои `repository/postgres → service → transport/http`, связанные интерфейсами.

```
cmd/todoapp/          main.go (DI) + Dockerfile
internal/
  core/                общие компоненты: domain-модели, ошибки, логгер,
                        пул соединений с БД, HTTP-сервер/роутер/middleware
  features/
    users/ tasks/ statistics/ web/   repository → service → transport/http
migrations/           SQL-миграции (golang-migrate)
public/                index.html — фронтенд-страничка (см. ниже)
```

**Ключевые решения:**
- **Optimistic locking** через поле `version` у `User`/`Task` (`UPDATE ... WHERE id = $1 AND version = $2`); конфликт версии → `404`.
- **Nullable[T]** — различает в PATCH "поле не передано" / "передано null" / "передано значение".
- API версионируется через префикс `/api/{version}` (сейчас `v1`).
- Middleware-цепочка: `RequestID → Logger → Trace → Panic`.
- Доменные ошибки маппятся в HTTP-коды: `ErrInvalidArgument→400`, `ErrNotFound→404`, `ErrConflict→409`, прочие/паника→`500`.

## Домены

- **User**: `id`, `version`, `full_name` (3–100 симв.), `phone_number` (опц., формат `+<цифры>`, 10–15 симв.)
- **Task**: `id`, `version`, `title` (1–100 симв.), `description` (опц., 1–1000 симв.), `completed`, `created_at`, `completed_at` (обязателен при `completed=true`), `author_user_id` (FK → `users`)
- **Statistics**: `tasks_created`, `tasks_completed`, `tasks_completed_rate`, `tasks_average_completion_time`

## HTTP API

Прод-сервер: `http://45.131.43.74:5050`, базовый префикс `/api/v1`.
Swagger UI: [http://45.131.43.74:5050/swagger/](http://45.131.43.74:5050/swagger/) (спецификация — `/swagger/doc.json`).

| Ресурс | Метод/путь | Описание |
|---|---|---|
| Users | `POST /users` | создать пользователя |
| | `GET /users?limit=&offset=` | список пользователей |
| | `GET /users/{id}` | получить пользователя |
| | `PATCH /users/{id}` | частично обновить |
| | `DELETE /users/{id}` | удалить |
| Tasks | `POST /tasks` | создать задачу |
| | `GET /tasks?user_id=&limit=&offset=` | список задач |
| | `GET /tasks/{id}` | получить задачу |
| | `PATCH /tasks/{id}` | частично обновить (в т.ч. `completed`) |
| | `DELETE /tasks/{id}` | удалить |
| Statistics | `GET /statistics?user_id=&from=&to=` | статистика по задачам (`from`/`to` в формате `YYYY-MM-DD`) |

Пример `POST /tasks`:
```json
{ "title": "Купить молоко", "description": "2 литра", "author_user_id": 1 }
```

Пример `PATCH /tasks/{id}` (отметить выполненной, `completed_at` проставится автоматически):
```json
{ "completed": true }
```

Формат ошибки:
```json
{ "message": "human-readable контекст", "error": "исходная ошибка" }
```

## Frontend

На `/` отдаётся `public/index.html` (см. `internal/features/web`) — простая HTML-страничка для наглядной работы с API (карточки задач и т.п.). Написана Claude.ai как демонстрационный пример использования API, самостоятельной ценности как фронтенд не несёт.

## Запуск

### Требования
Go 1.26+, Docker/Docker Compose, `make`.

### 1. Окружение
```bash
cp .env.example .env
```

| Переменная | Назначение |
|---|---|
| `POSTGRES_USER/PASSWORD/DB` | доступ к БД |
| `POSTGRES_TIMEOUT` | таймаут операций с БД |
| `LOGGER_LEVEL` | уровень логирования |
| `HTTP_ADDR` | адрес HTTP-сервера |
| `HTTP_SHUTDOWN_TIMEOUT` | таймаут graceful shutdown |
| `TIME_ZONE` | тайм-зона приложения |

### 2. База данных и миграции
```bash
make env-up       # поднять Postgres (данные — в ./out/pgdata)
make migrate-up   # применить миграции (make migrate-down — откатить)
```

### 3. Запуск приложения

**Локально:**
```bash
make todoapp-run   # go mod tidy + go run cmd/todoapp/main.go, логи в ./out/logs
```

**В Docker** (сборка образа из `cmd/todoapp/Dockerfile` и запуск сервиса `todoapp` из `docker-compose.yaml`, порт `5050`):
```bash
make todoapp-deploy
make ps            # статус контейнеров
```

### Прочие команды
`make env-down` / `make env-cleanup` — остановить БД / очистить `./out/pgdata` · `make env-port-forward` / `env-port-close` — проброс порта Postgres · `make migrate-create seq=<имя>` — новая миграция · `make logs-cleanup` — очистить логи.

## База данных

Схема `todoapp`:
- **`users`** — `id`, `version`, `full_name` (CHECK 3–100), `phone_number` (CHECK на формат и длину)
- **`tasks`** — `id`, `version`, `title` (CHECK 1–100), `description` (CHECK 1–1000), `completed`, `created_at`, `completed_at` (CHECK согласованности с `completed`), `author_user_id` (FK → `users.id`)
