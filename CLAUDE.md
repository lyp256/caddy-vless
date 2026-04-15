# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- `make build` — build the custom Caddy binary into `build/caddy` with `xcaddy`.
- `make test` — run all Go tests across non-vendor packages.
- `go test ./...` — run the full Go test suite directly.
- `go test ./pkg/...` — run tests only for library packages.
- `make all-check` — run `go mod tidy`, `go fmt`, `go vet`, then fail if that changed tracked files.
- `make fmt` / `make vet` / `make tidy` — run individual maintenance steps.
- `go generate ./...` — rebuild generated artifacts declared by `generate.go` using `xcaddy`.

## Architecture

This repo builds a custom Caddy binary with the VLESS module compiled in via `xcaddy`.

- `vless.go` is the Caddy module entrypoint at the module root. It registers the `vless` Caddyfile directive and the `http.handlers.vless` module, parses the `uuids` block from Caddyfile, validates configured UUIDs, and wraps requests with a pre-check that only allows configured users.
- `pkg/vless/` is the protocol implementation. `handshake.go` parses the inbound VLESS request header and destination address; `vless.go` performs the handshake response, selects TCP vs UDP dialing, opens the upstream connection, and relays traffic bidirectionally.
- `pkg/utils/http.go` adapts Caddy HTTP requests into an `io.ReadWriteCloser`. For HTTP/2 it uses a flushed response/body stream rather than a raw socket hijack.
- `pkg/utils/stream.go` contains the bidirectional transport loop used after the upstream dial succeeds.
- `Makefile` defines the local `xcaddy` build with `--with github.com/lyp256/caddy-vless=.` and `--with github.com/caddy-dns/alidns`.

## Request flow

1. Caddy routes a matching HTTP request to the `vless` handler.
2. `vless.go` converts configured UUID strings into an in-memory allowlist during provisioning.
3. `pkg/vless.NewHTTPHandler` turns the HTTP request/response pair into a stream, reads the VLESS handshake, and parses the destination.
4. The Caddy-layer pre-handler checks the client UUID against the allowlist.
5. The protocol handler sends the VLESS response header, dials the requested upstream destination, then pipes bytes in both directions until one side closes.

## Config model

The documented user-facing config is a Caddyfile handler block like:

```caddyfile
handle /vless/* {
    vless {
        uuids {
            <uuid>
        }
    }
}
```

Only the `uuids` sub-block is currently parsed by the module.

## Build and CI notes

- The primary local binary artifact is `build/caddy`.
- CI is defined in `.github/workflows/build.yml` and runs checks/tests on pushes to `main` and pull requests.
- Release workflows build Linux `amd64` and `arm64` binaries with `xcaddy` and upload them as GitHub release assets.
