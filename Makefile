BINPATH ?= build

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

LDFLAGS = -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

JAVA_SDK_DIR="./sdk/java"

.PHONY: all
all: delimiter-AUDIT audit delimiter-LINTERS lint delimiter-UNIT-TESTS test delimiter-COMPONENT_TESTS test-component delimiter-FINISH ## Runs multiple targets, audit, lint, test and test-component

.PHONY: audit
audit: audit-go audit-java ## Runs checks for security vulnerabilities on dependencies (including transient ones)
	
.PHONY: audit-go
audit-go:
	dis-vulncheck

.PHONY: audit-java
audit-java: 
	mvn -f $(JAVA_SDK_DIR) ossindex:audit

.PHONY: build
build: build-go build-java

.PHONY: build-go
build-go: ## Builds binary of application code and stores in bin directory as dis-redirect-api
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dis-redirect-api

.PHONY: build-java
build-java:	
	mvn -f $(JAVA_SDK_DIR) clean package -Dmaven.test.skip -Dossindex.skip=true

.PHONY: convey
convey: ## Runs unit test suite and outputs results on http://127.0.0.1:8080/
	goconvey ./...

.PHONY: debug
debug: ## Used to run code locally in debug mode
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dis-redirect-api
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dis-redirect-api

.PHONY: delimiter-%
delimiter-%:
	@echo '===================${GREEN} $* ${RESET}==================='

.PHONY: fmt
fmt: ## Run Go formatting on code
	go fmt ./...

.PHONY: lint
lint: lint-go lint-java

.PHONY: lint-go
lint-go: ## Used in ci to run linters against Go code
	golangci-lint run ./...

.PHONY: lint-java
lint-java:
	mvn -f $(JAVA_SDK_DIR) clean checkstyle:check test-compile

.PHONY: lint-local
lint-local: ## Use locally to run linters against Go code
	golangci-lint run ./...

.PHONY: validate-specification
validate-specification: # Validate swagger spec
	redocly lint swagger.yaml

.PHONY: test
test: test-go test-java

.PHONY: test-go
test-go: ## Runs unit tests including checks for race conditions and returns coverage
	go test -race -cover ./...

.PHONY: test-java
test-java:
	mvn -f $(JAVA_SDK_DIR) -Dossindex.skip=true test

.PHONY: test-component
test-component: ## Runs component test suite
	go test -cover -coverpkg=github.com/ONSdigital/dis-redirect-api/... -component

.PHONY: help
help: ## Show help page for list of make targets
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)
