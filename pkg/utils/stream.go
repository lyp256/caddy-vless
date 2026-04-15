package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

func Transport(upstream, downstream io.ReadWriter) (up, down int64, err error) {
	var upErr, downErr error
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		up, upErr = io.Copy(upstream, downstream)
		closeWrite(upstream)
		closeRead(downstream)
	}()
	go func() {
		defer wg.Done()
		down, downErr = io.Copy(downstream, upstream)
		closeWrite(downstream)
		closeRead(upstream)
	}()

	wg.Wait()

	errs := make([]string, 0, 2)
	if downErr != nil && !errors.Is(downErr, io.EOF) && !errors.Is(downErr, net.ErrClosed) {
		errs = append(errs, fmt.Sprintf("[upstream=>downstream]:%s", downErr))
	}
	if upErr != nil && !errors.Is(upErr, io.EOF) && !errors.Is(upErr, net.ErrClosed) {
		errs = append(errs, fmt.Sprintf("[downstream=>upstream]:%s", upErr))
	}
	if len(errs) > 0 {
		err = errors.New(strings.Join(errs, " "))
	}
	return
}

type closeWriter interface {
	CloseWrite() error
}

type closeReader interface {
	CloseRead() error
}

func closeWrite(rw io.ReadWriter) {
	if cw, ok := rw.(closeWriter); ok {
		_ = cw.CloseWrite()
		return
	}
	if c, ok := rw.(io.Closer); ok {
		_ = c.Close()
	}
}

func closeRead(rw io.ReadWriter) {
	if cr, ok := rw.(closeReader); ok {
		_ = cr.CloseRead()
		return
	}
	if c, ok := rw.(io.Closer); ok {
		_ = c.Close()
	}
}
