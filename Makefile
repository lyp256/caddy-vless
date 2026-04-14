VER=develop
COMMIT=$(shell git rev-parse --short HEAD)
GOMODULE=$(shell go list -m)
VERPKG=$(GOMODULE)/version

PKGS=$(shell go list ./... |grep -v vendor |xargs echo)

ifneq ($(REF_TYPE),tag)
TAG:=$(TAG)-$(COMMIT)
endif

.PHONY: fmt
fmt:
	go fmt $(PKGS)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vet
vet:
	go vet $(PKGS)

.PHONY: lint
lint:
	golangci-lint-v2 run ./...

.PHONY: all-check
all-check:tidy fmt vet lint git-check

.PHONY: git-check
git-check:
	git diff --exit-code

.PHONY: test
test:
	go test $(PKGS)

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags "-X $(VERPKG).Version=$(VER) -X $(VERPKG).CommitID=$(COMMIT)" -o build/ ./cmd/...