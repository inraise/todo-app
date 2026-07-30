# TODO-APP

Backend-сервис для управления задачами (TODO) с пользователями, задачами и статистикой по ним. Написан на Go, хранит данные в PostgreSQL, отдаёт HTTP JSON API.

## Стек технологий

- **Go 1.26**
- **PostgreSQL 17** (драйвер [pgx/v5](https://github.com/jackc/pgx))
- **net/http** (`http.ServeMux`) — без внешнего HTTP-фреймворка
- **go.uber.org/zap** — структурированное логирование
- **go-playground/validator** — валидация входящих запросов
- **golang-migrate** — миграции БД
- **kelseyhightower/envconfig** — конфигурация через переменные окружения
- **Docker Compose** — окружение для разработки (Postgres + инструмент миграций)

## Архитектура

Проект организован по принципу **feature-based** структуры с разделением на слои внутри каждой фичи:

```
cmd/todoapp/            точка входа (main.go), сборка зависимостей (DI)
internal/
  core/                 переиспользуемые компоненты, не привязанные к конкретной фиче
    config/             конфигурация приложения (тайм-зона)
    domain/             доменные модели: User, Task, Statistics, Nullable[T]
    errors/             общие sentinel-ошибки (ErrNotFound, ErrInvalidArgument, ErrConflict)
    logger/             обёртка над zap
    repository/postgres/pool/  абстракция над пулом соединений (pgx-реализация)
    transport/http/     HTTP-сервер, роутер с версионированием API, middleware,
                         разбор запросов (path/query params), единый формат ответов
  features/
    users/               репозиторий → сервис → HTTP-хендлеры пользователей
    tasks/                репозиторий → сервис → HTTP-хендлеры задач
    statistics/           репозиторий → сервис → HTTP-хендлеры статистики
migrations/              SQL-миграции (golang-migrate)
```

Каждая фича (`users`, `tasks`, `statistics`) имеет одинаковую внутреннюю структуру:

```
repository/postgres/   работа с БД (SQL-запросы, маппинг в модели)
service/               бизнес-логика, валидация, оркестрация репозитория
transport/http/        HTTP-хендлеры, DTO для запросов/ответов, регистрация роутов
```

Слои взаимодействуют через интерфейсы (например, `TasksService` описан в `transport/http`, а `TasksRepository` — в `service`), что позволяет независимо тестировать и заменять реализации.

### Ключевые архитектурные решения

- **Optimistic locking**: у `User` и `Task` есть поле `version`. При `PATCH`-обновлении запрос `UPDATE ... WHERE id = $1 AND version = $2` гарантирует, что запись не была изменена параллельно; при конфликте возвращается `404 Not Found` (запись "не найдена" в ожидаемой версии).
- **Nullable[T]**: обёртка для различения в PATCH-запросах трёх состояний поля — "не передано", "передано как null" и "передано значение". Используется в JSON DTO (`internal/core/transport/http/types`) и в доменных патчах (`internal/core/domain`).
- **Версионирование API**: роуты регистрируются под префиксом `/api/{version}` (`internal/core/transport/http/server`), сейчас доступна версия `v1`.
- **Middleware-цепочка**: `RequestID → Logger → Trace → Panic`, применяется как ко всему серверу, так и (при необходимости) к отдельным роутам.
- **Единый формат ошибок**: доменные ошибки (`ErrNotFound`, `ErrInvalidArgument`, `ErrConflict`) автоматически транслируются в HTTP-коды (`404`, `400`, `409`), любая иная ошибка — в `500`.

## Домены

### User (пользователь)
- `id`, `version`
- `full_name` — строка, 3–100 символов
- `phone_number` — опционально, формат `+<цифры>`, 10–15 символов

### Task (задача)
- `id`, `version`
- `title` — строка, 1–100 символов
- `description` — опционально, 1–1000 символов
- `completed` — флаг выполнения
- `created_at`, `completed_at` — `completed_at` обязателен, если `completed = true`, и не может предшествовать `created_at`
- `author_user_id` — ссылка на автора (`FOREIGN KEY` на `users`)

### Statistics (статистика по задачам)
- `tasks_created` — количество созданных задач
- `tasks_completed` — количество выполненных задач
- `tasks_completed_rate` — доля выполненных задач
- `tasks_average_completion_time` — среднее время выполнения задачи

## HTTP API

Базовый префикс: `/api/v1`.

### Пользователи (`/users`)

| Метод  | Путь          | Описание                                  |
|--------|---------------|--------------------------------------------|
| POST   | `/users`      | Создать пользователя                       |
| GET    | `/users`      | Список пользователей (`?limit=&offset=`)   |
| GET    | `/users/{id}` | Получить пользователя по ID                |
| PATCH  | `/users/{id}` | Частично обновить пользователя             |
| DELETE | `/users/{id}` | Удалить пользователя                       |

Пример тела запроса `POST /users`:
```json
{
  "full_name": "Иван Иванов",
  "phone_number": "+79991234567"
}
```

### Задачи (`/tasks`)

| Метод  | Путь          | Описание                                                |
|--------|---------------|------------------------------------------------------------|
| POST   | `/tasks`      | Создать задачу                                             |
| GET    | `/tasks`      | Список задач (`?user_id=&limit=&offset=`)                 |
| GET    | `/tasks/{id}` | Получить задачу по ID                                       |
| PATCH  | `/tasks/{id}` | Частично обновить задачу (в т.ч. отметить выполненной)     |
| DELETE | `/tasks/{id}` | Удалить задачу                                              |

Пример тела запроса `POST /tasks`:
```json
{
  "title": "Купить молоко",
  "description": "2 литра, обезжиренное",
  "author_user_id": 1
}
```

Пример тела запроса `PATCH /tasks/{id}` (отметить выполненной — `completed_at` проставится автоматически):
```json
{
  "completed": true
}
```

### Статистика (`/statistics`)

| Метод | Путь          | Описание                                                              |
|-------|---------------|--------------------------------------------------------------------------|
| GET   | `/statistics` | Статистика по задачам (`?user_id=&from=&to=`, формат дат `YYYY-MM-DD`)   |

Все параметры (`user_id`, `from`, `to`) опциональны и сужают выборку; без них считается статистика по всем задачам.

### Формат ошибок

```json
{
  "message": "human-readable контекст",
  "error": "исходная ошибка"
}
```

| Доменная ошибка       | HTTP-код |
|-------------------------|----------|
| `ErrInvalidArgument`    | 400      |
| `ErrNotFound`           | 404      |
| `ErrConflict`           | 409      |
| прочие (в т.ч. паника)  | 500      |

## Запуск проекта

### Требования
- Go 1.26+
- Docker и Docker Compose (для Postgres)
- `make`

### 1. Настройка окружения

Скопируйте `.env.example` в `.env` и заполните значения:

```bash
cp .env.example .env
```

Переменные окружения:

| Переменная              | Назначение                                                 |
|--------------------------|-------------------------------------------------------------|
| `POSTGRES_USER`          | пользователь БД                                              |
| `POSTGRES_PASSWORD`      | пароль БД                                                    |
| `POSTGRES_DB`             | имя базы данных                                              |
| `POSTGRES_TIMEOUT`       | таймаут операций с БД (например, `5s`)                       |
| `LOGGER_LEVEL`            | уровень логирования (`DEBUG`, `INFO`, ...)                   |
| `HTTP_ADDR`                | адрес, на котором слушает HTTP-сервер (например, `:8080`)    |
| `HTTP_SHUTDOWN_TIMEOUT`  | таймаут graceful shutdown (например, `30s`)                  |
| `TIME_ZONE`                | тайм-зона приложения (например, `UTC`)                        |

### 2. Поднять базу данных

```bash
make env-up
```

Поднимает контейнер `todoapp-postgres` (Postgres 17), данные сохраняются в `./out/pgdata`.

### 3. Применить миграции

```bash
make migrate-up
```

Откатить миграции: `make migrate-down`.
Создать новую миграцию: `make migrate-create seq=<имя>`.

### 4. Запустить приложение

```bash
make todoapp-run
```

Команда выполняет `go mod tidy` и запускает `cmd/todoapp/main.go` (подключение к Postgres через `localhost`, логи пишутся в `./out/logs`).

### Дополнительные команды

| Команда                | Описание                                                        |
|--------------------------|--------------------------------------------------------------------|
| `make env-down`          | остановить контейнер Postgres                                       |
| `make env-cleanup`       | остановить окружение и удалить файлы данных (`./out/pgdata`)        |
| `make env-port-forward`  | пробросить порт Postgres на `127.0.0.1:5432` через `socat`          |
| `make env-port-close`    | закрыть проброшенный порт                                            |
| `make logs-cleanup`      | очистить логи (`./out/logs`)                                         |

## Структура базы данных

Схема `todoapp`:

- **`users`** — `id`, `version`, `full_name` (`CHECK` 3–100 символов), `phone_number` (`CHECK` на формат `+<цифры>`, 10–15 символов)
- **`tasks`** — `id`, `version`, `title` (`CHECK` 1–100), `description` (`CHECK` 1–1000), `completed`, `created_at`, `completed_at` (согласованность с `completed` проверяется констрейнтом), `author_user_id` (`FOREIGN KEY → users.id`)

Миграции лежат в `migrations/` и применяются через `golang-migrate` (сервис `todoapp-postgres-migrate` в `docker-compose.yaml`).
