# URL Shortener

Сервис сокращения длинных URL, написанный на Go.

Проект реализует создание коротких ссылок, хранение URL в SQLite, кеширование через Redis и автоматический редирект по короткому адресу.

## Возможности

- Создание короткой ссылки
- Проверка существования URL перед созданием новой ссылки
- Перенаправление пользователя на исходный URL
- Хранение данных в SQLite
- Кеширование ссылок в Redis
- Docker Compose запуск приложения и Redis
- Graceful shutdown HTTP сервера

## Запуск проекта
Необходимо установить:
- Docker
- Docker Compose
Запуск по команде:
docker compose up --build
После запуска приложение доступно:
http://localhost:8080
Настройка через переменные окружения
PORT=8080

REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_TTL=24h

DB_PATH=/app/data/urls.db


## Пример работы 
Создание ссылки
curl -X POST http://localhost:8080/shorten \
-H "Content-Type: application/json" \
-d '{"url":"https://google.com"}'
Ответ:
{
  "short-url":"1"
}
Переход:
curl -I http://localhost:8080/1
Ответ:
HTTP/1.1 301 Moved Permanently
Location: https://google.com
---

## Архитектура

Проект построен по слоистой архитектуре:
1. Слой Handler
2. слой Service
3. слой Repository
### Handler

Отвечает за:
- обработку HTTP запросов;
- валидацию входных данных;
- формирование HTTP ответов.

### Service

Содержит бизнес-логику:
- проверка существования ссылки;
- генерация короткого кода;
- работа с кешем.

### Repository

Абстрагирует работу с хранилищами:

- SQLite repository — постоянное хранение данных;
- Redis repository — быстрый доступ к часто используемым ссылкам.

---

## Технологии

- Go 1.25
- net/http
- SQLite
- Redis
- Docker
- Docker Compose
## API

### Создание короткой ссылки

**POST**
/shorten
### Request

{
  "url": "https://google.com"
}
### Response
{
  "short-url": "1"
}
### Переход по короткой ссылке
**GET**
### Request
/{short-code}
Например: GET /1
### Response
HTTP/1.1 301 Moved Permanently
Location: https://google.com
## Схема БД
CREATE TABLE urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    long_url TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    clicks INTEGER DEFAULT 0
);
