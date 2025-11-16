FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git gcc musl-dev curl

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata curl bash

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

RUN mkdir -p /app/logs /app/migrations
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migrate.sh .
COPY --from=builder /app/migrations ./migrations

RUN chmod +x migrate.sh && \
    chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/cache/health || exit 1

CMD ["./main"]