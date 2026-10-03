.DEFAULT_GOAL := help

.PHONY: help
help: ## Показать доступные цели
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: install
install: ## Установить зависимости интерфейса
	cd web && npm install

.PHONY: api
api: ## Запустить сервер на :8080 (записи берёт из audio/)
	cd server && go run . -audio ../audio

.PHONY: web
web: ## Запустить интерфейс на :5180 (Vite проксирует /api на сервер)
	cd web && npm run dev

.PHONY: missing
missing: ## Показать, какие записи ещё не сделаны и куда их положить
	cd server && go run . -audio ../audio -missing

.PHONY: check
check: ## Проверить сервер и типы интерфейса
	cd server && go vet ./... && go test ./...
	cd web && npm run typecheck

.PHONY: build
build: ## Собрать интерфейс и бинарь сервера в bin/
	cd web && npm run build
	cd server && go build -o ../bin/domashka .

.PHONY: serve
serve: build ## Собрать и запустить всё одним сервером на :8080
	./bin/domashka -audio audio -static web/dist
