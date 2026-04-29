package vless

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type failingHandler struct {
	err error
}

func (h failingHandler) Handle(_ context.Context, _ io.ReadWriteCloser) error {
	return h.err
}

func TestHTTPStatus(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/vless", strings.NewReader("payload"))

	t.Run("success keeps continue status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		httpHandler{Handler: failingHandler{err: nil}}.ServeHTTP(rec, req.Clone(req.Context()))
		if rec.Code != http.StatusContinue {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusContinue)
		}
	})

	t.Run("failure still returns continue after hijack", func(t *testing.T) {
		rec := httptest.NewRecorder()
		httpHandler{Handler: failingHandler{err: assertiveErr("boom")}}.ServeHTTP(rec, req.Clone(req.Context()))
		if rec.Code != http.StatusContinue {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusContinue)
		}
	})
}

type assertiveErr string

func (e assertiveErr) Error() string { return string(e) }
