GOLANG_VERSION ?= $(shell grep '^go ' go.mod | sed 's/go //')
GOSUMDB = sum.golang.org
GOTOOLCHAIN = go$(GOLANG_VERSION)

SRCROOT = $(shell pwd)

MODULE_NAME = github.com/aws-neuron/connected-device-maps-over-efa-for-neuron

.PHONY: all
all: build

.PHONY: fmt
fmt:: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet:: ## Run go vet against code.
	go vet ./...

##@ Development
.PHONY: mod-tidy
mod-tidy:: ## Run go mod tidy against code.
	go mod tidy

.PHONY: clean-modcache
clean-modcache:: ## Run go clean -modcache against code.
	go clean -modcache

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.4.0
GOSEC = $(shell pwd)/bin/gosec
GOSEC_VERSION ?= v2.22.11

golangci-lint::
	@[ -f $(GOLANGCI_LINT) ] || { \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION) ;\
	}

.PHONY: lint
lint:: golangci-lint ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix:: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: gosec
gosec::
	@[ -f $(GOSEC) ] || GOBIN=$(shell dirname $(GOSEC)) go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)

.PHONY: security-scan
security-scan:: gosec ## Run gosec security scanner
	$(GOSEC) -fmt=text -confidence=medium -severity=medium ./...

##@ Build
.PHONY: build
build:: fmt vet security-scan

##@ Unit test
.PHONY: prepare-coverage
prepare-coverage::
	mkdir -p $(SRCROOT)/build/brazil-documentation/coverage/$(MODULE_NAME)

.PHONY: unit-test
unit-test:: prepare-coverage
	go test -v -tags unit `go list ./... | grep -v pkg` -count 1 -covermode=atomic -coverprofile $(SRCROOT)/build/brazil-documentation/coverage/$(MODULE_NAME)/coverage.out
	sed 's|$(MODULE_NAME)/||g' $(SRCROOT)/build/brazil-documentation/coverage/$(MODULE_NAME)/coverage.out > $(SRCROOT)/build/brazil-documentation/coverage/coverage.out

.PHONY: generate-coverage
generate-coverage:: unit-test
	go tool cover -html=$(SRCROOT)/build/brazil-documentation/coverage/$(MODULE_NAME)/coverage.out -o $(SRCROOT)/build/brazil-documentation/coverage/$(MODULE_NAME)/index.html
	cp $(SRCROOT)/build/brazil-documentation/coverage/$(MODULE_NAME)/index.html $(SRCROOT)/build/brazil-documentation/coverage/

##@ Release
.PHONY: release
release:: build generate-coverage
