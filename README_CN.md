# caddy-vless

中文 | [English](README.md)

Caddy 的 VLESS 协议支持模块。

## 构建

本项目使用 `xcaddy` 构建自定义 Caddy 二进制，并额外编译入：

- 当前仓库中的 `vless` 模块
- `github.com/caddy-dns/alidns`
- `github.com/caddy-dns/cloudflare`

本地构建：

```bash
make build
```

产物输出到：

- `build/caddy`

如果你需要直接执行底层命令，等价构建方式见 [Makefile](Makefile)。

## 运行

使用本地 Caddyfile 启动：

```bash
./build/caddy run --config ./build/caddyfile
```

查看模块是否已编译进二进制：

```bash
./build/caddy list-modules | grep vless
```

查看版本信息：

```bash
./build/caddy version
```

## 测试

```bash
make test
```

如需执行完整检查：

```bash
make all-check
```

## 发布

GitHub Releases 只发布 Linux `amd64` 和 `arm64` 二进制文件，不再提供 Docker 镜像。

CI 位于 [`.github/workflows/build.yml`](.github/workflows/build.yml)，会在以下场景运行：

- push 到 `main`
- pull request
- release 发布

## 配置示例

支持的配置项：

- `uuids`：允许访问的 VLESS 用户 UUID 列表
- `socks5`：可选的上游 SOCKS5 代理，格式为 `[username:password@]host:port`

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

带认证信息的 SOCKS5 配置示例：

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

未配置 `socks5` 时，VLESS 握手后的出站连接会继续直接连接目标地址。配置 `socks5` 后，后续上游连接会通过该代理建立。
