# Wishlist API

REST-сервис для вишлистов: регистрация и вход по JWT, CRUD вишлистов и позиций, публичная ссылка по токену и резервирование подарка без авторизации. Помимо всех основных требований, реализованы тесты и graceful shutdown.

## Стек

- Go
- chi
- PostgreSQL
- pgx
- JWT
- bcrypt
- validator
- golang-migrate

## Запуск

```bash
cp .env.example .env   
docker compose up --build
```

Перед стартом API выполняется сервис **migrate** (миграции из каталога `migrations/`). База поднимается с healthcheck; API ждёт готовности БД и успешного завершения миграций.

## Переменные окружения

| Переменная | Описание | Пример |
|------------|----------|--------|
| `SERVER_PORT` | Порт HTTP-сервера | `8080` |
| `SERVER_READ_TIMEOUT` | Таймаут чтения (секунды) | `10` |
| `SERVER_WRITE_TIMEOUT` | Таймаут записи (секунды) | `10` |
| `SERVER_IDLE_TIMEOUT` | Idle-таймаут (секунды) | `120` |
| `DB_HOST` | Хост PostgreSQL | `db` / `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER`, `DB_PASSWORD`, `DB_NAME` | Учётные данные БД | см. `.env.example` |
| `DB_SSL_MODE` | `sslmode` для DSN | `disable` |
| `JWT_SECRET` | Секрет подписи JWT | длинная случайная строка |
| `JWT_EXPIRATION` | Срок жизни токена в **секундах** | `86400` |

## API

Базовый префикс: **`/api/v1`**.

Ошибки в формате JSON: `{"error":"..."}`.

### Авторизация (без JWT)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/auth/register` | Регистрация; ответ `201`, тело `{"token":"..."}` |
| `POST` | `/api/v1/auth/login` | Вход; ответ `200`, тело `{"token":"..."}` |

Используется заголовок: **`Authorization: Bearer <token>`**.

### Вишлисты (JWT)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/wishlists/` | Создать вишлист; в ответе есть уникальный **`token`** для публичной ссылки |
| `GET` | `/api/v1/wishlists/` | Список вишлистов текущего пользователя |
| `GET` | `/api/v1/wishlists/{id}` | Вишлист с позициями |
| `PUT` | `/api/v1/wishlists/{id}` | Обновить |
| `DELETE` | `/api/v1/wishlists/{id}` | Удалить |

Тело создания (JSON):

- `title` (обязательно, до 255 символов)
- `description` (опционально, до 1000)
- `event_date` (опционально, формат `YYYY-MM-DD`, например `"2026-12-25"`)

Обновление: те же поля опционально через указатели (можно передать только изменяемые).

### Позиции вишлиста (JWT)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/wishlists/{wishlistID}/items/` | Создать позицию |
| `GET` | `/api/v1/wishlists/{wishlistID}/items/` | Список (сортировка: приоритет по убыванию, затем `created_at`) |
| `GET` | `/api/v1/wishlists/{wishlistID}/items/{itemID}` | Одна позиция |
| `PUT` | `/api/v1/wishlists/{wishlistID}/items/{itemID}` | Обновить |
| `DELETE` | `/api/v1/wishlists/{wishlistID}/items/{itemID}` | Удалить |

Тело создания позиции:

- `name` (обязательно)
- `description`, `url` (опционально; для `url` — валидный URL)
- `priority` — целое **0–10**; если поле не передать, используется **0**

### Публичные маршруты (без JWT)

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/api/v1/shared/{token}` | Вишлист по публичному токену с позициями |
| `POST` | `/api/v1/shared/{token}/items/{itemID}/reserve` | Зарезервировать подарок; если уже зарезервирован — **`409 Conflict`** |

## Тест через браузер (сайт сгенерирован через ии для удобства проверки работы апи, признаюсь)

Откройте файл `api-test.html` в браузере.
Отправляйте запросы через кнопки на странице, слева есть поле с ответами от сервера.

## Примеры `curl`

**Регистрация и сохранение токена:**

```bash
TOKEN=$(curl -sS -X POST "http://localhost:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password12"}' | jq -r .token)
```

**Создать вишлист:**

```bash
curl -sS -X POST "http://localhost:8080/api/v1/wishlists/" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"День рождения","description":"Идеи подарков","event_date":"2026-04-18"}'
```

В ответе возьмите `id` вишлиста, `token` для гостей и при необходимости `items[].id`.

**Добавить позицию:**

```bash
curl -sS -X POST "http://localhost:8080/api/v1/wishlists/1/items/" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Книга","description":"Про Go","url":"https://example.com/book","priority":5}'
```

**Публичный просмотр (подставьте `PUBLIC_TOKEN` из ответа создания вишлиста):**

```bash
curl -sS "http://localhost:8080/api/v1/shared/$PUBLIC_TOKEN"
```

**Резерв без JWT:**

```bash
curl -sS -X POST "http://localhost:8080/api/v1/shared/$PUBLIC_TOKEN/items/1/reserve"
```

## Тесты

```bash
go test ./...
```

## Структура репозитория (кратко)

- `cmd/api` — точка входа
- `internal/config` — конфигурация из окружения
- `internal/model` — модели и DTO запросов
- `internal/repository` — доступ к БД
- `internal/service` — бизнес-логика
- `internal/handler` — HTTP-хендлеры и роутер
- `internal/middleware` — JWT
- `migrations/` — SQL-миграции
