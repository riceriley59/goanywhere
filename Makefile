# include depdency targets
include deps.mk


# Global Variables


GIT_SHA 		:= $(shell git rev-parse --short HEAD)
GIT_TAG_LAST 	:= $(shell git tag --list 'operator*' --sort=-v:refname | head -n 1 | cut -d/ -f2)

## GO Flags
GO_LDFLAGS  := -ldflags "-X github.com/riceriley59/goanywhere/internal/version.GIT_SHA=$(GIT_SHA) \
	-X github.com/riceriley59/goanywhere/internal/version.VERSION=$(VERSION)"
GOFLAGS 	:= -mod=vendor


REPORTING ?= $(shell pwd)/reporting
.PHONY: reporting
reporting: $(REPORTING)
$(REPORTING):
	mkdir -p $@

# Default target, clean,  and help


.PHONY: all clean help
all: build

clean:
	rm -rf bin/

help:
	printf "hello"


# Build Targets


.PHONY: build build-goanywhere
build: build-goanywhere

build-goanywhere:
	go build $(GOFLAGS) $(GO_LDFLAGS) -o bin/goanywhere cmd/goanywhere/main.go


# Lint Targets


.PHONY: format fmt vet lint lint-fix

format:
	@echo "Formatting Go code..."
	gofmt -w .
	@echo "Done."

fmt:
	@echo "Checking formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not formatted correctly:"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "All files are properly formatted."

vet:
	@echo "Running go vet..."
	go vet $(GOFLAGS) ./...

lint: golangci-lint
	@echo "Running golangci-lint..."
	$(GOLANGCI_LINT) run ./...

lint-fix: golangci-lint
	@echo "Running golangci-lint with auto-fix..."
	$(GOLANGCI_LINT) run --fix ./...


# Test Targets


.PHONY: test unit-tests coverage coverage-html

test: unit-tests coverage

unit-tests: reporting ginkgo
	go test $(GOFLAGS) -coverprofile=$(REPORTING)/unit.coverprofile -covermode=atomic -coverpkg=./internal/... ./internal/... ./tests/... -v

coverage: reporting
	@echo ""
	@echo "=== Coverage Summary ==="
	@go tool cover -func=$(REPORTING)/unit.coverprofile | tail -1
	@echo ""

coverage-html: reporting
	@echo "Generating HTML coverage report..."
	go tool cover -html=$(REPORTING)/unit.coverprofile -o $(REPORTING)/coverage.html
	@echo "Coverage report generated at $(REPORTING)/coverage.html"


# CI Target


.PHONY: ci

ci: fmt vet lint test
	@echo "CI checks passed!"

