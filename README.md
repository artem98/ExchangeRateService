# ExchangeRateService

This application provides a currency exchange rate service with asynchronous updates and a simple client for manual testing.

## Project structure

- `server/` — main service with PostgreSQL database
- `client/` — optional Go CLI client for testing requests
- `docker-compose.yml` — launches the database and the server in containers

---

## Features

- Asynchronous update of currency exchange rates
- External exchange API integration
- PostgreSQL for persistent storage
- Simple HTTP JSON API
- Dockerized server with `docker-compose`
- Minimal CLI client for manual testing (not containerized)
- Update requests idempotency

---

## Server setup (Docker)

1. Build and start the server + database:
   ```sh
   cd server
   docker-compose up --build
   ```

2. Server will be available at:
   ```
   http://localhost:8080
   ```

---

## API Endpoints

1. **Trigger rate update**  
   `POST /rates/update_requests`  
   JSON body:
   ```json
   { "pair": "EUR/USD" }
   ```

2. **Get rate by update ID**  
   `GET /rates/update_requests/<id>`

3. **Get latest rate by pair**  
   `GET /rates?pair=EUR/USD`

---

## Using the CLI client

The `client` is a simple Go-based CLI tool for testing.

1. Build it manually:
   ```sh
   cd client
   go build -o client .
   ```

2. Run commands:
   ```sh
   ./client
   ```

---

## Requirements

- Go 1.24.5+
- Docker & Docker Compose

---

## Config tips

- Server params can be configured in `server/constants/constants.go` file

## Русская версия

Это сервис валютных котировок с асинхронным обновлением и простой клиент для ручного тестирования.

### Структура

- `server/` — серверное приложение с БД PostgreSQL
- `client/` — вспомогательный Go-клиент для ручных тестов (не в контейнере)
- `docker-compose.yml` — запускает базу и сервер

---

## Особенности

- Асинхронное обновление валютных курсов
- Интеграция внешнего API обменных курсов
- PostgreSQL в качестве базы данных
- HTTP JSON API
- Контейнеризированный сервер и БД с помощью `docker-compose`
- Минимальный CLI клиент для ручного тестирования (не контейнеризированный)
- Идемпотентность запросов обновлений

---

### Запуск сервера (Docker)

1. Соберите и запустите:
   ```sh
   cd server
   docker-compose up --build
   ```

2. Сервер будет доступен по адресу:
   ```
   http://localhost:8080
   ```

---

### API

1. **Запросить обновление курса**  
   `POST /rates/update_requests`  
   Тело запроса:
   ```json
   { "pair": "EUR/USD" }
   ```

2. **Получить курс по ID обновления**  
   `GET /rates/update_requests/<id>`

3. **Получить последний курс по валютной паре**  
    `GET /rates?pair=EUR/USD`

---

### Использование клиента

1. Перейди в папку `client` и собери:
   ```sh
   cd client
   go build -o client .
   ```

2. Используй:
   ```sh
   ./client
   ```

---

### Требования

- Go 1.24.5+
- Docker и Docker Compose

## Управление конфигурацией

-  Параметры серверы могут быть настроены в файле `server/constants/constants.go`

