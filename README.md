# ExchangeRateService

This application provides a currency exchange rate service with asynchronous updates and a simple client for manual testing.

## Project structure

- `server/` � main service with PostgreSQL database
- `client/` � optional Go CLI client for testing requests
- `docker-compose.yml` � launches the database and the server in containers

---

## Features

- Asynchronous update of currency exchange rates
- External exchange API integration
- PostgreSQL for persistent storage
- Simple HTTP JSON API
- Dockerized server with `docker-compose`
- Minimal CLI client for testing (not containerized)

---

## Server setup (Docker)

1. Build and start the server + database:
   ```sh
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

## ������� ������

��� ������ �������� ��������� � ����������� ����������� � ������� ������ ��� ������� ������������.

### ���������

- `server/` � ��������� ���������� � �� PostgreSQL
- `client/` � ��������������� Go-������ ��� ������ ������ (�� � ����������)
- `docker-compose.yml` � ��������� ���� � ������

---

### ������ ������� (Docker)

1. �������� � ���������:
   ```sh
   docker-compose up --build
   ```

2. ������ ����� �������� �� ������:
   ```
   http://localhost:8080
   ```

---

### API

1. **��������� ���������� �����**  
   `POST /rates/update_requests`  
   ���� �������:
   ```json
   { "pair": "EUR/USD" }
   ```

2. **�������� ���� �� ID ����������**  
   `GET /rates/update_requests/<id>`

3. **�������� ��������� ���� �� �������� ����**  
    `GET /rates?pair=EUR/USD`

---

### ������������� �������

1. ������� � ����� `client` � ������:
   ```sh
   cd client
   go build -o client .
   ```

2. ���������:
   ```sh
   ./client
   ```

---

### ����������

- Go 1.24.5+
- Docker � Docker Compose
