FROM golang:1.24.4-alpine AS builder

RUN apk update && apk add build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /migrator ./cmd/migrator/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -o /main ./cmd/main/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /main .
COPY --from=builder /migrator .

COPY db db
COPY .env .env
COPY ./pkg/email/templates ./pkg/email/templates
COPY ./public ./public

RUN DB_FILE=./db.db MIGRATIONS_PATH=./db/migrations ./migrator

EXPOSE 8000

CMD [ "/app/main", "--config=/app/.env" ]