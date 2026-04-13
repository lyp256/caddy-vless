package vless

import (
	"context"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/google/uuid"
	"github.com/lyp256/caddy-vless/pkg/vless"
	"net/http"
)

func init() {
	caddy.RegisterModule(&vlessModule{})
	httpcaddyfile.RegisterHandlerDirective("vless", parseCaddyfileHandler)
	httpcaddyfile.RegisterDirectiveOrder("vless", httpcaddyfile.After, "handle")
}

// vlessModule implements an HTTP handler that writes the
// visitor's IP address to a file or stream.
type vlessModule struct {
	UUIDS []string `json:"uuids"`

	userIDs map[string]struct{}
}

// CaddyModule returns the Caddy module information.
func (*vlessModule) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.vless",
		New: func() caddy.Module { return new(vlessModule) },
	}
}

// Provision implements caddy.Provisioner.
func (m *vlessModule) Provision(ctx caddy.Context) error {
	if m == nil || m.UUIDS == nil {
		return nil
	}
	m.userIDs = make(map[string]struct{}, len(m.UUIDS))
	for _, s := range m.UUIDS {
		m.userIDs[s] = struct{}{}
	}
	return nil
}

// Validate implements caddy.Validator.
func (m *vlessModule) Validate() error {
	if m == nil || m.UUIDS == nil {
		return nil
	}
	for _, val := range m.UUIDS {
		if _, err := uuid.Parse(val); err != nil {
			return err
		}

	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m *vlessModule) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	vless.NewHTTPHandler(vless.WithPreHandle(func(ctx context.Context, request vless.Requester) error {
		if m == nil || m.userIDs == nil {
			return nil
		}
		u := request.UUID()
		uid := u.String()
		_, ok := m.userIDs[uid]
		if !ok {
			return fmt.Errorf("uuid %s not found", u.String())
		}
		return nil
	})).ServeHTTP(w, r)
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *vlessModule) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		opt := d.Val()
		switch opt {
		case "uuids":
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				val := d.Val()
				m.UUIDS = append(m.UUIDS, val)
			}
		default:
			return d.ArgErr()
		}
	}
	return nil
}

func parseCaddyfileHandler(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	i := new(vlessModule)
	err := i.UnmarshalCaddyfile(h.Dispenser)
	return i, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*vlessModule)(nil)
	_ caddy.Validator             = (*vlessModule)(nil)
	_ caddyhttp.MiddlewareHandler = (*vlessModule)(nil)
	_ caddyfile.Unmarshaler       = (*vlessModule)(nil)
	_ caddyfile.Unmarshaler       = (*vlessModule)(nil)
)
