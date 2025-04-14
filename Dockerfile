FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/initdb ./cmd/initdb
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/dispatcher ./cmd/dispatcher
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/worker ./cmd/worker

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/initdb /app/initdb
COPY --from=builder /app/dispatcher /app/dispatcher
COPY --from=builder /app/worker /app/worker
