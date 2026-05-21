.PHONY: run build test tidy lint docker-build docker-up docker-down clean

# ── Development ───────────────────────────────────────────────────────────────
run:
	go run ./cmd/server

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/schulbot ./cmd/server

test:
	go test -v -race ./...

tidy:
	go mod tidy

lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "golangci-lint not found – install via https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

# ── Docker ────────────────────────────────────────────────────────────────────
docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# ── Cleanup ───────────────────────────────────────────────────────────────────
clean:
	rm -rf bin/ data/
