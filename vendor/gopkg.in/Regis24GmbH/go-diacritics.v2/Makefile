.PHONY: lint test race coverage upgrade help

LIST_ALL := $(shell go list ./... | grep -v /vendor/ | grep -v mocks)

export GOFLAGS=-mod=vendor

lint: ## Lint all files (via golint)
	@golint -set_exit_status ${LIST_ALL}

test: ## Run unittests
	@go test -short -count 1 -v ./...

race: ## Run data race detector
	@go test -race -short -count 1 -v ./...

coverage: ## Generate test coverage
	@go test -coverprofile coverage.txt ./...
	@go tool cover -func=coverage.txt

upgrade: ## Upgrade go dependencies
	@GOFLAGS='' go get -u -t ./...
	@go mod tidy
	@go mod vendor

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
