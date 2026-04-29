package vless

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/lyp256/caddy-vless/pkg/utils"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type Handler interface {
	Handle(ctx context.Context, connect io.ReadWriteCloser) error
}

type handler struct {
	preHandle  func(ctx context.Context, request Requester) error
	postHandle func(ctx context.Context, request Requester, upBytes, downBytes int64, err error)
	dialer     Dialer
	logger     *zap.Logger
}

func (h *handler) Handle(ctx context.Context, connect io.ReadWriteCloser) error {
	// vless handshake
	request := requestPool.Get().(*requestInfo)
	defer requestPool.Put(request)
	err := request.FromReader(connect)
	if err != nil {
		return fmt.Errorf("handshake:%w", err)
	}
	if h.logger != nil {
		clientID := request.UUID()
		h.logger.Debug("parsed vless request",
			zap.String("client_id", (&clientID).String()),
			zap.String("destination", request.DestAddr()),
			zap.Uint8("command", uint8(request.Command())),
		)
	}
	// verify
	if h.preHandle != nil {
		err = h.preHandle(ctx, request)
		if err != nil {
			return fmt.Errorf("preVerify:%w", err)
		}
	}
	var network string
	switch request.Command() {
	case TCP:
		network = "tcp"
	case UDP:
		network = "udp"
	default:
		return fmt.Errorf("unknown command:%d", request.Command())
	}

	// replay
	resp := Response{
		Version: request.Version(),
		Addons:  nil,
	}
	_, err = resp.WriteTo(connect)
	if err != nil {
		return fmt.Errorf("reply :%w", err)
	}

	// dial
	upStream, err := h.dialer.DialContext(ctx, network, request.DestAddr())
	if err != nil {
		return fmt.Errorf("dial %s %s fail:%s", network, request.DestAddr(), err)
	}
	defer func() { _ = upStream.Close() }()

	up, down, err := utils.Transport(upStream, connect, h.logger)
	if h.postHandle != nil {
		h.postHandle(ctx, request, up, down, err)
	}
	if err != nil {
		return fmt.Errorf("traffic forward:%w", err)
	}
	return nil

}

type Option func(*handler)

func WithPreHandle(preHand func(ctx context.Context, request Requester) error) Option {
	return func(h *handler) {
		h.preHandle = preHand
	}
}

func WithPostHandle(postHandle func(ctx context.Context, request Requester, upBytes, downBytes int64, err error)) Option {
	return func(h *handler) {
		h.postHandle = postHandle
	}
}

func WithDial(dialer Dialer) Option {
	return func(h *handler) {
		h.dialer = dialer
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(h *handler) {
		h.logger = logger
	}
}

func NewHandler(opts ...Option) Handler {
	h := &handler{
		dialer: &net.Dialer{
			Timeout: time.Second * 10,
		},
	}
	for i := range opts {
		if opts[i] != nil {
			opts[i](h)
		}
	}
	return h
}

type httpHandler struct {
	Handler
	logger *zap.Logger
}

func (h httpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	connect, err := utils.H2Hijack(writer, request)
	if err != nil {
		if h.logger != nil {
			h.logger.Warn("vless hijack failed", zap.Error(err))
		}
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer func() { _ = connect.Close() }()

	err = h.Handle(request.Context(), connect)
	if err == nil {
		return
	}

	if h.logger != nil {
		h.logger.Debug("vless request failed", zap.Error(err))
	}
	// 为了 vless 隐蔽性，直接返回 404
	http.NotFound(writer, request)
}

// NewHTTPHandler create handle over http
func NewHTTPHandler(opts ...Option) http.Handler {
	h := &handler{
		dialer: &net.Dialer{
			Timeout: time.Second * 10,
		},
	}
	for i := range opts {
		if opts[i] != nil {
			opts[i](h)
		}
	}
	return httpHandler{
		Handler: h,
		logger:  h.logger,
	}
}
