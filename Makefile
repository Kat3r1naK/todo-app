.PHONY: dev dev-build prod prod-build down down-prod logs logs-prod ps ps-prod generate wire-gen

PROD_ENV_FILE := $(if $(wildcard .env.prod),--env-file .env.prod,)

dev:
	docker compose up --build

dev-build:
	docker compose up -d --build

prod:
	docker compose $(PROD_ENV_FILE) -f docker-compose.prod.yml up -d --build

prod-build:
	docker compose $(PROD_ENV_FILE) -f docker-compose.prod.yml build

down:
	docker compose down

down-prod:
	docker compose $(PROD_ENV_FILE) -f docker-compose.prod.yml down

logs:
	docker compose logs -f api web db

logs-prod:
	docker compose $(PROD_ENV_FILE) -f docker-compose.prod.yml logs -f api web db

ps:
	docker compose ps

ps-prod:
	docker compose $(PROD_ENV_FILE) -f docker-compose.prod.yml ps

generate: wire-gen

wire-gen:
	docker compose exec api sh -c "go install github.com/google/wire/cmd/wire@v0.7.0 && wire ./ioc"
