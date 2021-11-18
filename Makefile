# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

export PATH := $(abspath bin/):${PATH}

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
	gotestsum --no-summary=skipped --junitfile ${BUILD_DIR}/coverage.xml --jsonfile ${BUILD_DIR}/test.json --format ${TEST_FORMAT} -- -race -coverprofile=${BUILD_DIR}/coverage.txt -covermode=atomic ./...

.PHONY: lint
lint: ## Run linter
	golangci-lint run ${LINT_ARGS}

.PHONY: fix
fix: ## Fix lint violations
	golangci-lint run --fix

# Dependency versions
GOTESTSUM_VERSION ?= 1.7.0
GOLANGCI_VERSION ?= 1.43.0

deps: bin/gotestsum bin/golangci-lint

bin/gotestsum:
	@mkdir -p bin
	curl -L https://github.com/gotestyourself/gotestsum/releases/download/v${GOTESTSUM_VERSION}/gotestsum_${GOTESTSUM_VERSION}_$(shell uname | tr A-Z a-z)_amd64.tar.gz | tar -zOxf - gotestsum > ./bin/gotestsum
	@chmod +x ./bin/gotestsum

bin/golangci-lint:
	@mkdir -p bin
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'
