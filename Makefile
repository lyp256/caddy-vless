VER=develop
COMMIT=$(shell git rev-parse --short HEAD)
TAG?=$(VER)

PKGS=$(shell go list ./... |grep -v vendor |xargs echo)

ifneq ($(REF_TYPE),tag)
TAG:=$(TAG)-$(COMMIT)
endif

XCADDY?=$(shell go env GOPATH)/bin/xcaddy
MODULE=$(shell go list -m)
XCADDY_BUILD_ARGS=build --with $(MODULE)=. --with github.com/caddy-dns/alidns --with github.com/caddy-dns/cloudflare

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
	golangci-lint run ./...

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
	mkdir -p build
	CGO_ENABLED=0 $(XCADDY) $(XCADDY_BUILD_ARGS) --output ./build/caddy
