package utils

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

type closingBuffer struct {
	mu          sync.Mutex
	buf         bytes.Buffer
	writeClosed bool
	readClosed  bool
}

func (b *closingBuffer) Read(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.buf.Len() > 0 {
		return b.buf.Read(p)
	}
	if b.writeClosed || b.readClosed {
		return 0, io.EOF
	}
	return 0, io.EOF
}

func (b *closingBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.writeClosed {
		return 0, net.ErrClosed
	}
	return b.buf.Write(p)
}

func (b *closingBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.writeClosed = true
	b.readClosed = true
	return nil
}

func (b *closingBuffer) CloseWrite() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.writeClosed = true
	return nil
}

func (b *closingBuffer) CloseRead() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.readClosed = true
	return nil
}

func TestTransportClosesPeersAfterCopyCompletes(t *testing.T) {
	upstream := &closingBuffer{}
	downstream := &closingBuffer{}
	if _, err := downstream.Write([]byte("ping")); err != nil {
		t.Fatalf("seed downstream: %v", err)
	}

	done := make(chan struct{})
	var up, down int64
	var err error
	go func() {
		up, down, err = Transport(upstream, downstream, nil)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Transport() did not return")
	}

	if err != nil {
		t.Fatalf("Transport() error = %v", err)
	}
	if up != 4 {
		t.Fatalf("up = %d, want 4", up)
	}
	if down != 0 {
		t.Fatalf("down = %d, want 0", down)
	}
	if !upstream.writeClosed {
		t.Fatal("expected upstream CloseWrite to be called")
	}
	if !downstream.readClosed {
		t.Fatal("expected downstream CloseRead to be called")
	}
}
