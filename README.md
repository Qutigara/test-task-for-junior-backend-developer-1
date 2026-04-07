# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

## Периодичность задач (Recurrence)

Добавлена возможность задавать периодичность выполнения задач.

### Типы периодичности

| Тип | Описание | Параметры |
|-----|----------|-----------|
| `none` | Без периодичности (по умолчанию) | - |
| `daily` | Ежедневные задачи | `interval` — каждый n-й день |
| `monthly` | Ежемесячные задачи | `day_of_month` — день месяца (1-30) |
| `specific_dates` | На конкретные даты | `specific_dates` — массив дат |
| `even_days` | Чётные дни месяца | - |
| `odd_days` | Нечётные дни месяца | - |

### Примеры запросов

#### Ежедневная задача (каждые 3 дня)

```json
POST /api/v1/tasks
{
  "title": "Ежедневная проверка",
  "description": "Проверка системы",
  "status": "new",
  "recurrence": {
    "type": "daily",
    "interval": 3
  }
}
```

#### Ежемесячная задача (15 числа)

```json
POST /api/v1/tasks
{
  "title": "Зарплата",
  "description": "Выплата зарплаты",
  "status": "new",
  "recurrence": {
    "type": "monthly",
    "day_of_month": 15
  }
}
```

#### Задача на конкретные даты

```json
POST /api/v1/tasks
{
  "title": "Отчёт",
  "description": "Ежемесячный отчёт",
  "status": "new",
  "recurrence": {
    "type": "specific_dates",
    "specific_dates": ["2026-04-30", "2026-05-31", "2026-06-30"]
  }
}
```

#### Задача на чётные дни

```json
POST /api/v1/tasks
{
  "title": "Чётные дни",
  "description": "Выполняется по чётным числам",
  "status": "new",
  "recurrence": {
    "type": "even_days"
  }
}
```

#### Задача на нечётные дни

```json
POST /api/v1/tasks
{
  "title": "Нечётные дни",
  "description": "Выполняется по нечётным числам",
  "status": "new",
  "recurrence": {
    "type": "odd_days"
  }
}
```

### Валидация

- Для `daily` — обязательный `interval` > 0
- Для `monthly` — обязательный `day_of_month` от 1 до 30
- Для `specific_dates` — обязательный массив `specific_dates` с хотя бы одной датой
- Для `even_days` и `odd_days` — дополнительная валидация не требуется

### Структура БД

Добавлены поля в таблицу `tasks`:

- `recurrence_type` — тип периодичности (TEXT)
- `recurrence_interval` — интервал для daily (INTEGER)
- `recurrence_day_of_month` — день месяца для monthly (INTEGER)
- `recurrence_specific_dates` — конкретные даты (JSONB)

Миграция: `migrations/0002_add_recurrence.up.sql`

## Изменения в коде

### Domain

- Добавлен тип `RecurrenceType` с константами периодичности
- Добавлена структура `Recurrence` с полями `Type`, `Interval`, `DayOfMonth`, `SpecificDates`
- Добавлено поле `Recurrence` в структуру `Task`

### UseCase

- Добавлено поле `Recurrence` в `CreateInput` и `UpdateInput`
- Добавлена валидация для всех типов периодичности в `validateRecurrence()`

### Repository

- Обновлены SQL-запросы (Create, GetByID, Update, List) для работы с новыми полями
- Добавлен маппинг полей периодичности в `scanTask()`

### Transport/HTTP

- Обновлены DTO: добавлена структура `recurrenceDTO`
- Обновлены handlers для передачи данных о периодичности

### OpenAPI

- Обновлена спецификация `openapi.json` с описанием схемы `Recurrence`

## Валидация

### Валидация recurrence

- Для `daily` — обязательный `interval` > 0
- Для `monthly` — обязательный `day_of_month` от 1 до 30
- Для `specific_dates` — обязательный массив с хотя бы одной датой
  - Формат дат: `YYYY-MM-DD`
  - Даты не могут быть в прошлом
- Для `even_days` и `odd_days` — дополнительная валидация не требуется

## Воркер для периодических задач

Автоматический воркер проверяет каждый час все задачи с настроенной периодичностью и создаёт новые задачи на текущий день.

### Логика работы

| Тип периодичности | Поведение |
|-------------------|-----------|
| `daily` | Создаёт задачу каждый день |
| `monthly` | Создаёт задачу в указанный день месяца |
| `specific_dates` | Создаёт задачу только на указанные даты |
| `even_days` | Создаёт задачу по чётным числам |
| `odd_days` | Создаёт задачу по нечётным числам |

### Название создаваемых задач

Новая задача создаётся с названием: `{оригинальное название} ({дата})`

Например: `Ежедневная проверка (2026-04-07)`

### Защита от дубликатов

Воркер проверяет, что задача с таким названием ещё не существует (и не завершена).
