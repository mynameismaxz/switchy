BINARY_NAME := swy
MODULE      := github.com/mynameismaxz/switchy
VERSION     ?= dev
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE        := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build install test lint clean help

help: ## Show available make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

build: ## Build binary to ./bin/swy
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/swy

install: ## Install binary to $GOPATH/bin
	go install -ldflags "$(LDFLAGS)" ./cmd/swy

test: ## Run all tests
	go test ./...

lint: ## Run go vet
	go vet ./...

clean: ## Remove ./bin directory
	rm -rf bin/
