.DEFAULT_GOAL := help

CMD := ./cmd/domashka
BASE_CONFIG := -config config/base.yaml

.PHONY: help
help: ## Show available targets
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: install
install: ## Install frontend dependencies
	cd web && npm install

.PHONY: api
api: ## Run the server on :8080 with the admin mode (is_admin=true in localStorage)
	go run $(CMD) $(BASE_CONFIG) -admin

.PHONY: web
web: ## Run the frontend on :5180 (Vite proxies /api to the server)
	cd web && npm run dev

.PHONY: missing
missing: ## Show which recordings are missing
	go run $(CMD) $(BASE_CONFIG) -missing

.PHONY: voices
voices: ## Generate voiceover candidates with every voice where they are missing
	go run $(CMD) $(BASE_CONFIG) -voices

.PHONY: voices-all
voices-all: ## Regenerate all voiceover candidates (e.g. after editing speech.json)
	go run $(CMD) $(BASE_CONFIG) -voices-all

.PHONY: check
check: ## Vet and test the server, typecheck and lint the frontend
	go vet ./...
	go test ./...
	cd web && npm run typecheck && npm run lint

.PHONY: build
build: ## Build the frontend and the server binary into bin/
	cd web && npm run build
	go build -o bin/domashka $(CMD)

.PHONY: serve
serve: build ## Build and run everything as one server on :8080 (no admin mode, as for students)
	./bin/domashka $(BASE_CONFIG) -config config/serve.yaml

.PHONY: docker-build
docker-build: ## Build the production image domashka:latest
	docker build -t domashka:latest .

.PHONY: docker-up
docker-up: ## Start the production container (compose.yaml) on :8080
	docker compose up -d --build

.PHONY: docker-down
docker-down: ## Stop the production container (gracefully, via SIGTERM)
	docker compose down

.PHONY: deploy
deploy: ## Build and deploy to a server over ssh: make deploy SERVER=user@host (PORT=…)
	./deploy/deploy.sh "$(SERVER)"
