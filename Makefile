GO ?= go
GOFMT ?= gofmt
GOLANGCI_LINT ?= golangci-lint
GOFILES := $(shell find . -type f -name '*.go' -not -path './vendor/*' -not -path './.git/*')

.PHONY: test test-race vet golint lint-fix fmt fmt-check tidy check examples

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

golint:
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || { \
		echo "$(GOLANGCI_LINT) is not installed. Install from https://golangci-lint.run/welcome/install/"; \
		exit 1; \
	}
	$(GOLANGCI_LINT) run --config .golangci.yml ./...

lint-fix:
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || { \
		echo "$(GOLANGCI_LINT) is not installed. Install from https://golangci-lint.run/welcome/install/"; \
		exit 1; \
	}
	$(GOLANGCI_LINT) run --fix --config .golangci.yml ./...

fmt:
	$(GOFMT) -w $(GOFILES)

fmt-check:
	@out="$$($(GOFMT) -l $(GOFILES))"; \
	if [ -n "$$out" ]; then \
		echo "Unformatted files:"; \
		echo "$$out"; \
		exit 1; \
	fi

tidy:
	$(GO) mod tidy

check: fmt-check vet golint test

examples:
	$(GO) test ./examples/...
