# proxy

caddy 添加 vless 协议支持模块

GitHub Releases 只发布 Linux `amd64` 和 `arm64` 二进制文件，不再提供 Docker 镜像。

Example:
```
https://example.com {
    tls example.crt example.key
    handle /vless/*  {
       vless {
        uuids {
            1d74253d-f391-4fef-ac0e-b93bd15f8ecf
            2d74253d-f391-4fef-ac0e-b93bd15f8eca
            3d74253d-f391-4fef-ac0e-b93bd15f8ecc
        }
       }
    }
    respond "Hello, world!"
}
```
