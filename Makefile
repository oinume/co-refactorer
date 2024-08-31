NAME = co-refactorer
GO_TEST ?= go test -v -race -p=1
GOLANGCI_LINT_VERSION = v1.60.3

.PHONY: all
all: build

help: ## Show this help
	@perl -nle 'BEGIN {printf "Usage:\n  make \033[33m<target>\033[0m\n\nTargets:\n"} printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 if /^([a-zA-Z_-].+):.*\s+## (.*)/' $(MAKEFILE_LIST)
.PHONY: help

build:
	go build -o bin/$(NAME) ./cmd/$(NAME)/main.go
.PHONY: build

test: ## Run go test
	$(GO_TEST) ./...
.PHONY: test

lint: ## Run golangci-lint
	docker run --rm -v ${GOPATH}/pkg/mod:/go/pkg/mod -v $(shell pwd):/app -v $(shell go env GOCACHE):/cache/go -e GOCACHE=/cache/go -e GOLANGCI_LINT_CACHE=/cache/go -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run --modules-download-mode=readonly /app/...
.PHONY: lint

lint/fix: ## Run golangci-lint with --fix
	docker run --rm -v ${GOPATH}/pkg/mod:/go/pkg/mod -v $(shell pwd):/app -v $(shell go env GOCACHE):/cache/go -e GOCACHE=/cache/go -e GOLANGCI_LINT_CACHE=/cache/go -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run --fix --modules-download-mode=readonly /app/...
.PHONY: lint/fix

lint/version: ## Show golangci-lint version
	@echo $(GOLANGCI_LINT_VERSION)

.PHONY: clean
clean:
	${RM} bin/$(NAME)
	${RM} -fr dist/*
