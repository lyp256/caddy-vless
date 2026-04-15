package vless

import (
	"fmt"
	"net/http"
	"testing"
)

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil", err: nil, want: http.StatusOK},
		{name: "pre verify", err: fmt.Errorf("preVerify:%w", fmt.Errorf("denied")), want: http.StatusForbidden},
		{name: "handshake", err: fmt.Errorf("handshake:%w", fmt.Errorf("eof")), want: http.StatusBadRequest},
		{name: "unknown command", err: fmt.Errorf("unknown command:%d", 3), want: http.StatusBadRequest},
		{name: "reply", err: fmt.Errorf("reply :%w", fmt.Errorf("broken pipe")), want: http.StatusBadRequest},
		{name: "dial", err: fmt.Errorf("dial tcp example.com:443 fail:%s", "refused"), want: http.StatusBadGateway},
		{name: "forward", err: fmt.Errorf("traffic forward:%w", fmt.Errorf("reset")), want: http.StatusBadGateway},
		{name: "fallback", err: fmt.Errorf("other"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := httpStatus(tt.err); got != tt.want {
				t.Fatalf("httpStatus() = %d, want %d", got, tt.want)
			}
		})
	}
}
