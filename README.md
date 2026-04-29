# caddy-vless

[中文](README_CN.md) | English

A Caddy module that adds VLESS protocol support.

## Build

This project uses `xcaddy` to build a custom Caddy binary and additionally compiles in:

- the `vless` module from this repository
- `github.com/caddy-dns/alidns`
- `github.com/caddy-dns/cloudflare`

Build locally:

```bash
make build
```

The output binary is written to:

- `build/caddy`

If you need to run the underlying command directly, see the equivalent build definition in [Makefile](Makefile).

## Run

Start with the local Caddyfile:

```bash
./build/caddy run --config ./build/caddyfile
```

Check whether the module has been compiled into the binary:

```bash
./build/caddy list-modules | grep vless
```

Show version information:

```bash
./build/caddy version
```

## Test

```bash
make test
```

To run the full checks:

```bash
make all-check
```

## Release

GitHub Releases only publishes Linux `amd64` and `arm64` binaries. Docker images are no longer provided.

CI is defined in [`.github/workflows/build.yml`](.github/workflows/build.yml) and runs in these cases:

- pushes to `main`
- pull requests
- release publishing

## Configuration Example

Supported configuration options:

- `uuids`: list of allowed VLESS user UUIDs
- `socks5`: optional upstream SOCKS5 proxy in the format `[username:password@]host:port`

```caddyfile
https://example.com {
    tls example.crt example.key

    handle /vless/* {
        vless {
            uuids {
                1d74253d-f391-4fef-ac0e-b93bd15f8ecf
                2d74253d-f391-4fef-ac0e-b93bd15f8eca
                3d74253d-f391-4fef-ac0e-b93bd15f8ecc
            }
            socks5 127.0.0.1:1080
        }
    }

    respond "Hello, world!"
}
```

Example SOCKS5 configuration with authentication:

```caddyfile
https://example.com {
    handle /vless/* {
        vless {
            uuids {
                1d74253d-f391-4fef-ac0e-b93bd15f8ecf
            }
            socks5 user:password@127.0.0.1:1080
        }
    }
}
```

If `socks5` is not configured, outbound connections after the VLESS handshake connect directly to the target address. If `socks5` is configured, subsequent upstream connections are established through that proxy.
