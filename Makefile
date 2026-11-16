.PHONY: dev dev-build prod prod-build prod-deploy prod-down prod-logs prod-ps down logs ps generate wire-gen

dev:
	docker compose up --build

dev-build:
	docker compose up -d --build

prod: prod-deploy

prod-build:
	bash ./scripts/prod.sh build

prod-deploy:
	bash ./scripts/prod.sh deploy

prod-down:
	bash ./scripts/prod.sh down

prod-logs:
	bash ./scripts/prod.sh logs

prod-ps:
	bash ./scripts/prod.sh ps

down:
	docker compose down

logs:
	docker compose logs -f api web db

ps:
	docker compose ps

generate: wire-gen

wire-gen:
	docker compose exec api sh -c "go install github.com/google/wire/cmd/wire@v0.7.0 && wire ./ioc"
