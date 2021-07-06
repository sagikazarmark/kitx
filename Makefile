# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

# Build variables
BUILD_DIR ?= build
export CGO_ENABLED ?= 0

.PHONY: check
check: test lint ## Run checks (tests and linters)

.PHONY: test
test: TEST_FORMAT ?= short
test: export CGO_ENABLED=1
test: ## Run tests
	@mkdir -p ${BUILD_DIR}
	gotestsum --no-summary=skipped --junitfile ${BUILD_DIR}/coverage.xml --format ${TEST_FORMAT} -- -race -coverprofile=${BUILD_DIR}/coverage.txt -covermode=atomic ./...

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: fix
fix: ## Fix lint violations
	golangci-lint run --fix

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'
