BINDIR := $(PWD)/bin
PATH := $(BINDIR):$(PATH)

.PHONY: build
build: test
	@go build .

.PHONY: test
test: generate
	@go test ./...

.PHONY: generate
generate: $(MOCKERY)
	@go generate ./...

MOCKERY_VERSION := v2.42.2
MOCKERY_BINDIR := $(BINDIR)/bin/mockery/$(MOCKERY_VERSION)
MOCKERY := $(MOCKERY_BINDIR)/mockery
$(MOCKERY):
	@GOBIN=$(MOCKERY_BINDIR) go install github.com/vektra/mockery/v2@$(MOCKERY_VERSION)
	@ln -sf $(MOCKERY) $(BINDIR)/mockery
