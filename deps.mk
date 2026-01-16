## Versions
GOLANGCI_LINT_VERSION ?= v2.7.2
GOCOVER_VERSION ?= v1.4.0
GINKGO_VERSION ?= v2.27.2

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
golangci-lint: ## Download golangci locally if necessary.
	@[ -f $(GOLANGCI_LINT) ] || { \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION) ;\
	}

GOCOVER_COBERTURA ?= $(LOCALBIN)/gocover-cobertura
GINKGO ?= $(LOCALBIN)/ginkgo

.PHONY: gocover-cobertura
gocover-cobertura: ## Download gocover-cobertura locally if necessary.
	test -s $(LOCALBIN)/gocover-cobertura || GOBIN=$(LOCALBIN) go install github.com/boumenot/gocover-cobertura@$(GOCOVER_VERSION)

.PHONY: ginkgo
ginkgo: ## Download ginkgo locally if necessary.
	test -s $(LOCALBIN)/ginkgo || GOBIN=$(LOCALBIN) go install github.com/onsi/ginkgo/v2/ginkgo@$(GINKGO_VERSION)
