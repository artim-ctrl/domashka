.DEFAULT_GOAL := help

.PHONY: help
help: ## Показать доступные цели
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: install
install: ## Установить зависимости интерфейса
	cd web && npm install

.PHONY: api
api: ## Запустить сервер на :8080 с админкой (is_admin=true в localStorage)
	cd server && go run . -audio ../audio -admin

.PHONY: web
web: ## Запустить интерфейс на :5180 (Vite проксирует /api на сервер)
	cd web && npm run dev

.PHONY: missing
missing: ## Показать, каких записей ещё нет
	cd server && go run . -audio ../audio -missing

.PHONY: voices
voices: ## Заранее сделать варианты озвучки всеми голосами, где их ещё нет
	cd server && go run . -voices

.PHONY: voices-all
voices-all: ## Сделать все варианты озвучки заново (например, после правок speech.json)
	cd server && go run . -voices-all

.PHONY: check
check: ## Проверить сервер и типы интерфейса
	cd server && go vet ./... && go test ./...
	cd web && npm run typecheck

.PHONY: build
build: ## Собрать интерфейс и бинарь сервера в bin/
	cd web && npm run build
	cd server && go build -o ../bin/domashka .

.PHONY: serve
serve: build ## Собрать и запустить всё одним сервером на :8080 (без админки, как для учеников)
	./bin/domashka -audio audio -static web/dist

.PHONY: docker-build
docker-build: ## Собрать прод-образ domashka:latest
	docker build -t domashka:latest .

.PHONY: docker-up
docker-up: ## Поднять прод-контейнер (compose.yaml) на :8080
	docker compose up -d --build

.PHONY: docker-down
docker-down: ## Остановить прод-контейнер (штатно, через SIGTERM)
	docker compose down

.PHONY: deploy
deploy: ## Собрать и выкатить на сервер по ssh: make deploy SERVER=user@host (PORT=…)
	./deploy/deploy.sh "$(SERVER)"
