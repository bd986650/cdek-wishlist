# Wishlist API

REST API сервис для создания вишлистов к праздникам и событиям. Пользователь может зарегистрироваться, создать вишлист, наполнить его подарками и поделиться по уникальной ссылке.

## Технологии

- **Go 1.23** — язык
- **PostgreSQL 16** — база данных
- **chi** — HTTP-роутер
- **pgx** — драйвер PostgreSQL
- **golang-migrate** — миграции БД
- **golang-jwt** — JWT-авторизация
- **bcrypt** — хэширование паролей
- **Docker Compose** — оркестрация

## Архитектура

```
cmd/api/          — точка входа
internal/
  config/         — конфигурация из ENV
  model/          — доменные модели и DTO
  repository/     — слой доступа к данным (PostgreSQL)
  service/        — бизнес-логика
  handler/        — HTTP-хендлеры (транспортный слой)
  middleware/     — JWT-авторизация
migrations/       — SQL-миграции
```

## Запуск

```bash
cp .env.example .env
docker-compose up --build
```

API доступен на `http://localhost:8080`.

## API эндпоинты

### Авторизация

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/auth/register` | Регистрация |
| POST | `/api/v1/auth/login` | Вход |

### Вишлисты (требуется авторизация)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/wishlists` | Создать вишлист |
| GET | `/api/v1/wishlists` | Получить все вишлисты |
| GET | `/api/v1/wishlists/:id` | Получить вишлист по ID |
| PUT | `/api/v1/wishlists/:id` | Обновить вишлист |
| DELETE | `/api/v1/wishlists/:id` | Удалить вишлист |

### Позиции (требуется авторизация)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/wishlists/:id/items` | Добавить позицию |
| GET | `/api/v1/wishlists/:id/items` | Получить все позиции |
| GET | `/api/v1/wishlists/:id/items/:itemID` | Получить позицию |
| PUT | `/api/v1/wishlists/:id/items/:itemID` | Обновить позицию |
| DELETE | `/api/v1/wishlists/:id/items/:itemID` | Удалить позицию |

### Публичный доступ (без авторизации)

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/shared/:token` | Получить вишлист по токену |
| POST | `/api/v1/shared/:token/items/:itemID/reserve` | Забронировать подарок |

## Примеры запросов

### Регистрация

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "secret123"}'
```

### Создание вишлиста

```bash
curl -X POST http://localhost:8080/api/v1/wishlists \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title": "День рождения", "description": "Мой вишлист на ДР", "event_date": "2025-06-15"}'
```

### Добавление подарка

```bash
curl -X POST http://localhost:8080/api/v1/wishlists/1/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name": "Книга", "description": "Война и мир", "url": "https://example.com/book", "priority": 5}'
```

### Публичный просмотр

```bash
curl http://localhost:8080/api/v1/shared/<wishlist-token>
```

### Бронирование подарка

```bash
curl -X POST http://localhost:8080/api/v1/shared/<wishlist-token>/items/1/reserve
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|-------------|----------|
| `SERVER_PORT` | `8080` | Порт сервера |
| `DB_HOST` | `db` | Хост PostgreSQL |
| `DB_PORT` | `5432` | Порт PostgreSQL |
| `DB_USER` | `wishlist` | Пользователь БД |
| `DB_PASSWORD` | `wishlist` | Пароль БД |
| `DB_NAME` | `wishlist` | Имя БД |
| `DB_SSL_MODE` | `disable` | SSL-режим |
| `JWT_SECRET` | — | Секрет для JWT |
| `JWT_EXPIRATION` | `86400` | Время жизни токена (сек) |
