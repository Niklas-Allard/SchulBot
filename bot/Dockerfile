FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /schulbot ./cmd/server

# ── Runtime image ─────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /schulbot /usr/local/bin/schulbot

RUN addgroup -S schulbot && adduser -S schulbot -G schulbot
RUN mkdir -p /app/data && chown schulbot:schulbot /app/data

USER schulbot

VOLUME ["/app/data"]

ENTRYPOINT ["schulbot"]
