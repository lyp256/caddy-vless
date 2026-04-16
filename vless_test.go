package vless

import (
	"net"
	"strings"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestParseSOCKS5(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		wantAddr     string
		wantUser     string
		wantPassword string
		wantErr      string
	}{
		{name: "empty", raw: ""},
		{name: "address only", raw: "127.0.0.1:1080", wantAddr: "127.0.0.1:1080"},
		{name: "with credentials", raw: "user:pass@127.0.0.1:1080", wantAddr: "127.0.0.1:1080", wantUser: "user", wantPassword: "pass"},
		{name: "missing port", raw: "127.0.0.1", wantErr: "missing port in address"},
		{name: "missing password", raw: "user@127.0.0.1:1080", wantErr: "invalid socks5 credentials"},
		{name: "missing username", raw: ":pass@127.0.0.1:1080", wantErr: "invalid socks5 credentials"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAddr, gotAuth, err := parseSOCKS5(tt.raw)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("parseSOCKS5() error = nil, want %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("parseSOCKS5() error = %q, want substring %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSOCKS5() error = %v", err)
			}
			if gotAddr != tt.wantAddr {
				t.Fatalf("parseSOCKS5() addr = %q, want %q", gotAddr, tt.wantAddr)
			}
			if tt.wantUser == "" && gotAuth != nil {
				t.Fatalf("parseSOCKS5() auth = %+v, want nil", gotAuth)
			}
			if tt.wantUser != "" {
				if gotAuth == nil {
					t.Fatal("parseSOCKS5() auth = nil, want credentials")
				}
				if gotAuth.User != tt.wantUser || gotAuth.Password != tt.wantPassword {
					t.Fatalf("parseSOCKS5() auth = %+v, want user=%q password=%q", gotAuth, tt.wantUser, tt.wantPassword)
				}
			}
		})
	}
}

func TestVlessModuleValidate(t *testing.T) {
	tests := []struct {
		name    string
		module  vlessModule
		wantErr string
	}{
		{name: "valid uuid and socks5", module: vlessModule{UUIDS: []string{"123e4567-e89b-12d3-a456-426614174000"}, SOCKS5: "user:pass@127.0.0.1:1080"}},
		{name: "invalid uuid", module: vlessModule{UUIDS: []string{"bad-uuid"}}, wantErr: "invalid UUID length"},
		{name: "invalid socks5", module: vlessModule{SOCKS5: "127.0.0.1"}, wantErr: "missing port in address"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.module.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() error = nil, want %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestUnmarshalCaddyfileSOCKS5(t *testing.T) {
	input := `vless {
		socks5 user:pass@127.0.0.1:1080
		uuids {
			123e4567-e89b-12d3-a456-426614174000
		}
	}`

	var m vlessModule
	if err := m.UnmarshalCaddyfile(caddyfile.NewTestDispenser(input)); err != nil {
		t.Fatalf("UnmarshalCaddyfile() error = %v", err)
	}
	if m.SOCKS5 != "user:pass@127.0.0.1:1080" {
		t.Fatalf("SOCKS5 = %q", m.SOCKS5)
	}
	if len(m.UUIDS) != 1 || m.UUIDS[0] != "123e4567-e89b-12d3-a456-426614174000" {
		t.Fatalf("UUIDS = %v", m.UUIDS)
	}
}

func TestUnmarshalCaddyfileSOCKS5RequiresOneArg(t *testing.T) {
	input := `vless {
		socks5
	}`

	var m vlessModule
	if err := m.UnmarshalCaddyfile(caddyfile.NewTestDispenser(input)); err == nil {
		t.Fatal("expected arg error")
	}
}

func TestNewSOCKS5Dialer(t *testing.T) {
	dialer, err := newSOCKS5Dialer("127.0.0.1:1080")
	if err != nil {
		t.Fatalf("newSOCKS5Dialer() error = %v", err)
	}
	if dialer == nil {
		t.Fatal("newSOCKS5Dialer() = nil")
	}

	dialer, err = newSOCKS5Dialer("")
	if err != nil {
		t.Fatalf("newSOCKS5Dialer(empty) error = %v", err)
	}
	if dialer != nil {
		t.Fatalf("newSOCKS5Dialer(empty) = %v, want nil", dialer)
	}
}

func TestSocks5DialerDialContextCancelled(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer func() { _ = ln.Close() }()
}
