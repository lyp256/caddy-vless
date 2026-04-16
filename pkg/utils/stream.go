package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

type closeWriter interface {
	CloseWrite() error
}

type closeReader interface {
	CloseRead() error
}

func Transport(upstream, downstream io.ReadWriter) (up, down int64, err error) {
	var upErr, downErr error
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		up, upErr = io.Copy(upstream, downstream)
		if cw, ok := upstream.(closeWriter); ok {
			if closeErr := cw.CloseWrite(); upErr == nil && ignoreErrs(closeErr, net.ErrClosed) != nil {
				upErr = closeErr
			}
		}
		if cr, ok := downstream.(closeReader); ok {
			if closeErr := cr.CloseRead(); upErr == nil && ignoreErrs(closeErr, net.ErrClosed) != nil {
				upErr = closeErr
			}
		}

	}()
	go func() {
		defer wg.Done()
		down, downErr = io.Copy(downstream, upstream)
		if cw, ok := downstream.(closeWriter); ok {
			if closeErr := cw.CloseWrite(); downErr == nil && ignoreErrs(closeErr, net.ErrClosed) != nil {
				downErr = closeErr
			}
		}
		if cr, ok := upstream.(closeReader); ok {
			if closeErr := cr.CloseRead(); downErr == nil && ignoreErrs(closeErr, net.ErrClosed) != nil {
				downErr = closeErr
			}
		}
	}()

	wg.Wait()

	errs := make([]error, 0, 2)
	if ignoreErrs(downErr, io.EOF) != nil {
		switch {
		case errors.Is(downErr, net.ErrClosed):
			fallthrough
		case errors.Is(downErr, context.Canceled):
			logrus.Debugf("[upstream=>downstream]:%s", err)
		default:
			errs = append(errs, fmt.Errorf("[upstream=>downstream]:%w", downErr))
		}
	}
	if ignoreErrs(upErr, io.EOF) != nil {
		switch {
		case errors.Is(upErr, net.ErrClosed):
			fallthrough
		case errors.Is(upErr, context.Canceled):
			logrus.Debugf("[downstream=>upstream]:%s", upErr)
		default:
			errs = append(errs, fmt.Errorf("[downstream=>upstream]:%w", upErr))
		}
	}
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}
	return
}

func ignoreErrs(err error, ignores ...error) error {
	for _, e := range ignores {
		if errors.Is(err, e) {
			return nil
		}
	}
	return err
}
