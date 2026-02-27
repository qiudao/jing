.DEFAULT_GOAL := help

.PHONY: help build run test clean lint

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	go build -o bin/tictactoe .

run: build ## Build and run the server
	./bin/tictactoe

test: ## Run all tests
	go test ./... -v

clean: ## Remove build artifacts
	rm -rf bin/

lint: ## Run go vet
	go vet ./...
