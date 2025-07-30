# Dockerfile

FROM golang:1.24.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server .

FROM ubuntu

RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*
    
WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
