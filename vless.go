package vless

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/google/uuid"
	pkgvless "github.com/lyp256/caddy-vless/pkg/vless"
	"golang.org/x/net/proxy"
)

func init() {
	caddy.RegisterModule(&vlessModule{})
	httpcaddyfile.RegisterHandlerDirective("vless", parseCaddyfileHandler)
	httpcaddyfile.RegisterDirectiveOrder("vless", httpcaddyfile.After, "handle")
}

// vlessModule implements an HTTP handler that writes the
// visitor's IP address to a file or stream.
type vlessModule struct {
	UUIDS  []string `json:"uuids"`
	SOCKS5 string   `json:"socks5,omitempty"`

	userIDs map[string]struct{}
	dialer  pkgvless.Dialer
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
	if m == nil {
		return nil
	}

	if m.UUIDS != nil {
		m.userIDs = make(map[string]struct{}, len(m.UUIDS))
		for _, s := range m.UUIDS {
			m.userIDs[s] = struct{}{}
		}
	}

	dialer, err := newSOCKS5Dialer(m.SOCKS5)
	if err != nil {
		return err
	}
	m.dialer = dialer
	return nil
}

// Validate implements caddy.Validator.
func (m *vlessModule) Validate() error {
	if m == nil {
		return nil
	}
	for _, val := range m.UUIDS {
		if _, err := uuid.Parse(val); err != nil {
			return err
		}

	}
	_, _, err := parseSOCKS5(m.SOCKS5)
	return err
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m *vlessModule) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	opts := []pkgvless.Option{pkgvless.WithPreHandle(func(ctx context.Context, request pkgvless.Requester) error {
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
	})}
	if m != nil && m.dialer != nil {
		opts = append(opts, pkgvless.WithDial(m.dialer))
	}
	pkgvless.NewHTTPHandler(opts...).ServeHTTP(w, r)
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
		case "socks5":
			args := d.RemainingArgs()
			if len(args) != 1 {
				return d.ArgErr()
			}
			m.SOCKS5 = args[0]
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

func parseSOCKS5(raw string) (string, *proxy.Auth, error) {
	if raw == "" {
		return "", nil, nil
	}

	u, err := url.Parse("socks5://" + raw)
	if err != nil {
		return "", nil, err
	}
	if u.Host == "" {
		return "", nil, fmt.Errorf("invalid socks5 address")
	}
	if _, _, err := net.SplitHostPort(u.Host); err != nil {
		return "", nil, err
	}
	if u.User == nil {
		return u.Host, nil, nil
	}

	username := u.User.Username()
	password, ok := u.User.Password()
	if username == "" || !ok || password == "" {
		return "", nil, fmt.Errorf("invalid socks5 credentials")
	}
	return u.Host, &proxy.Auth{User: username, Password: password}, nil
}

func newSOCKS5Dialer(raw string) (pkgvless.Dialer, error) {
	addr, auth, err := parseSOCKS5(raw)
	if err != nil || addr == "" {
		return nil, err
	}

	baseDialer, err := proxy.SOCKS5("tcp", addr, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	return socks5Dialer{dialer: baseDialer}, nil
}

type socks5Dialer struct {
	dialer proxy.Dialer
}

func (d socks5Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	type dialResult struct {
		conn net.Conn
		err  error
	}

	resultCh := make(chan dialResult, 1)
	go func() {
		conn, err := d.dialer.Dial(network, address)
		resultCh <- dialResult{conn: conn, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultCh:
		if result.err != nil {
			return nil, result.err
		}
		return result.conn, nil
	}
}

// Interface guards
var (
	_ caddy.Provisioner           = (*vlessModule)(nil)
	_ caddy.Validator             = (*vlessModule)(nil)
	_ caddyhttp.MiddlewareHandler = (*vlessModule)(nil)
	_ caddyfile.Unmarshaler       = (*vlessModule)(nil)
)
