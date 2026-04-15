package vless

import (
	"bytes"
	"io"
	"testing"

	xrayuuid "github.com/xtls/xray-core/common/uuid"
)

type chunkedReader struct {
	data []byte
	step int
}

func (r *chunkedReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	if r.step < 1 {
		r.step = 1
	}
	n := r.step
	if n > len(r.data) {
		n = len(r.data)
	}
	if n > len(p) {
		n = len(p)
	}
	copy(p, r.data[:n])
	r.data = r.data[n:]
	return n, nil
}

func buildHandshake(t *testing.T) []byte {
	t.Helper()

	id, err := xrayuuid.ParseString("123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteByte(Version)
	buf.Write(id.Bytes())
	buf.WriteByte(0)
	buf.WriteByte(byte(TCP))
	buf.Write([]byte{0x01, 0xbb})
	buf.WriteByte(AddrTypeDomain)
	buf.WriteByte(byte(len("example.com")))
	buf.WriteString("example.com")
	return buf.Bytes()
}

func TestRequestInfoFromReaderHandlesFragmentedInput(t *testing.T) {
	reader := &chunkedReader{data: buildHandshake(t), step: 1}

	var req requestInfo
	if err := req.FromReader(reader); err != nil {
		t.Fatalf("FromReader() error = %v", err)
	}

	if got := req.Version(); got != Version {
		t.Fatalf("Version() = %d, want %d", got, Version)
	}
	if got := req.Command(); got != TCP {
		t.Fatalf("Command() = %d, want %d", got, TCP)
	}
	if got := req.DestAddr(); got != "example.com:443" {
		t.Fatalf("DestAddr() = %q, want %q", got, "example.com:443")
	}
	uid := req.UUID()
	if got := uid.String(); got != "123e4567-e89b-12d3-a456-426614174000" {
		t.Fatalf("UUID() = %q", got)
	}
}

func TestRequestInfoFromReaderResetsStateOnError(t *testing.T) {
	var req requestInfo
	if err := req.FromReader(bytes.NewReader(buildHandshake(t))); err != nil {
		t.Fatalf("initial FromReader() error = %v", err)
	}

	if err := req.FromReader(bytes.NewReader([]byte{Version})); err == nil {
		t.Fatal("expected truncated read error")
	}
	if req.effective {
		t.Fatal("expected request to be invalid after failed parse")
	}
	if req.destAddr != "" {
		t.Fatalf("destAddr = %q, want empty", req.destAddr)
	}
	if req.addons != nil {
		t.Fatalf("addons = %v, want nil", req.addons)
	}
}
