.PHONY: all build deps dev-deps image migrate test vet sec format unused static generate dev down docker-test docker-build docker-clean kapply-dev kustomize-dev kustomize-prod
CHECK_FILES?=./...

ifdef RELEASE_VERSION
	VERSION=$(RELEASE_VERSION)
else
	VERSION=$(shell git describe --tags)
endif

FLAGS=-ldflags "-X github.com/supabase/auth/internal/utilities.Version=$(VERSION)" -buildvcs=false

ifneq ($(shell docker compose version 2>/dev/null),)
  DOCKER_COMPOSE=docker compose
else
  DOCKER_COMPOSE=docker-compose
endif

DEV_DOCKER_COMPOSE:=docker-compose.yml

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: vet sec static build ## Run the tests and build the binary.

build: deps ## Build the binary.
	CGO_ENABLED=0 go build $(FLAGS)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(FLAGS) -o auth-arm64

build-strip: deps ## Build a stripped binary, for which the version file needs to be rewritten.
	echo "package utilities" > internal/utilities/version.go
	echo "const Version = \"$(VERSION)\"" >> internal/utilities/version.go

	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
		$(FLAGS) -ldflags "-s -w" -o auth-arm64-strip

deps: ## Install dependencies.
	@go mod download
	@go mod verify

migrate_dev: ## Run database migrations for development.
	hack/migrate.sh postgres

migrate_test: ## Run database migrations for test.
	hack/migrate.sh postgres

test: build ## Run tests.
	go test $(CHECK_FILES) -coverprofile=coverage.out -coverpkg ./... -p 1 -race -v -count=1
	./hack/coverage.sh

vet: # Vet the code
	go vet $(CHECK_FILES)

sec: # Check for security vulnerabilities
	go tool gosec -quiet -exclude-generated $(CHECK_FILES)
	go tool gosec -quiet -tests -exclude-generated -exclude=G104 $(CHECK_FILES)

unused: # Look for unused code
	@echo "Unused code:"
	go tool staticcheck -checks U1000 $(CHECK_FILES)
	@echo
	@echo "Code used only in _test.go (do move it in those files):"
	go tool staticcheck -checks U1000 -tests=false $(CHECK_FILES)

static:
	go tool staticcheck ./...
	go tool exhaustive ./...

generate:
	go tool oapi-codegen ./...
	go generate ./...

dev: ## Run the development containers
	${DOCKER_COMPOSE} -f $(DEV_DOCKER_COMPOSE) up

down: ## Shutdown the development containers
	# Start postgres first and apply migrations
	${DOCKER_COMPOSE} -f $(DEV_DOCKER_COMPOSE) down -v

docker-test: ## Run the tests using the development containers
	${DOCKER_COMPOSE} -f $(DEV_DOCKER_COMPOSE) up -d postgres
	${DOCKER_COMPOSE} -f $(DEV_DOCKER_COMPOSE) run auth sh -c "make migrate_test"
	${DOCKER_COMPOSE} -f $(DEV_DOCKER_COMPOSE) run auth sh -c "make test"
	${DOCKER_COMPOSE} -f $(DEV_DOCKER_COMPOSE) down -v

docker-build: ## Force a full rebuild of the development containers
	${DOCKER_COMPOSE} -f $(DEV_DOCKER_COMPOSE) build --no-cache
	${DOCKER_COMPOSE} -f $(DEV_DOCKER_COMPOSE) up -d postgres
	${DOCKER_COMPOSE} -f $(DEV_DOCKER_COMPOSE) run auth sh -c "make migrate_dev"
	${DOCKER_COMPOSE} -f $(DEV_DOCKER_COMPOSE) down

docker-clean: ## Remove the development containers and volumes
	${DOCKER_COMPOSE} -f $(DEV_DOCKER_COMPOSE) rm -fsv

format:
	gofmt -s -w .

kapply: ## Apply all YAML files in k8s directory recursively.
	kubectl apply -f k8s --recursive

kustomize-dev: ## Apply kustomization in kustomize/overlays/dev
	kubectl apply -k kustomize/overlays/dev

kustomize-prod: ## Apply kustomization in kustomize/overlays/prod
	kubectl apply -k kustomize/overlays/prod
