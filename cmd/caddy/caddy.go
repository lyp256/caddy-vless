package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/lyp256/caddy-vless/pkg/caddy/vless"
)

func main() {
	caddycmd.Main()
}
