.DEFAULT_GOAL := help
.PHONY: test lint check install-linters help

test: ## Run tests
	GO111MODULE=on go test -v ./cmd/... -race -timeout=1m -cover
	GO111MODULE=on go test -v ./src/... -race -timeout=1m -cover

lint: ## Run linters. Use make install-linters first.
	GO111MODULE=on golangci-lint run -c .golangci.yml ./...
	go vet -all ./...

check: lint test ## Run tests and linters

format: ## Formats the code. Must have goimports installed (use make install-linters).
	GO111MODULE=on goimports -w -local github.com/kittycash/wallet ./cmd
	GO111MODULE=on goimports -w -local github.com/kittycash/wallet ./src

install-linters: ## Install linters
	GO111MODULE=on go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
