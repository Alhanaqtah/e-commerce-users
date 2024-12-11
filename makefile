run:
	AUTH_CONF_PATH=config/config.yaml go run ./cmd/main.go

sandbox-up:
	docker compose -f dev/sandbox/docker-compose.yaml up -d 

sandbox-down:
	docker compose -f dev/sandbox/docker-compose.yaml down

test.unit:
	go test ./...

lint:
	golangci-lint run ./...