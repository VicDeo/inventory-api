-include .env

up-dev:
	docker compose up -d
down-dev:
	docker compose down
run:
	go run main.go
build-docs:
	swag init
test:
	@curl -si ${APP_URL}/inventory | head -n 1 | grep "200 OK" && echo "Believe me, it works!" || echo "Oops!"