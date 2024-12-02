run:
	AUTH_CONF_PATH=./config/config.yaml go run ./cmd/main.go

sandbox up:
	docker compose -f dev/sandbox/docker-compose.yaml up

sandbox down:
	docker compose -f dev/sandbox/docker-compose.yaml down