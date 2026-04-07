GO ?= go
GOFMT ?= gofmt
GOFILES := $(shell find . -type f -name '*.go' -not -path './vendor/*' -not -path './.git/*')

.PHONY: test test-race vet fmt fmt-check tidy check examples

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

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

check: fmt-check vet test

examples:
	$(GO) test ./examples/...
