package utils

import (
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

func WSShanke(u websocket.Upgrader, w http.ResponseWriter, r *http.Request) (io.ReadWriteCloser, error) {
	// websocket 协议
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &wsStream{
		conn: conn,
	}, nil
}

var (
	_ io.ReadWriteCloser = (*wsStream)(nil)
)

type wsStream struct {
	conn   *websocket.Conn
	reader io.Reader
}

func (ws *wsStream) Write(data []byte) (n int, err error) {
	return len(data), ws.conn.WriteMessage(websocket.BinaryMessage, data)

}

func (h *wsStream) Read(p []byte) (n int, err error) {
	for {
		if h.reader != nil {
			n, err = h.reader.Read(p)
			if errors.Is(err, io.EOF) {
				err = nil
				h.reader = nil
			}
			return
		}
		mt, reader, err := h.conn.NextReader()
		if err != nil {
			return 0, err
		}
		if mt == websocket.BinaryMessage {
			h.reader = reader
		}
	}
}

func (h *wsStream) Close() (err error) {
	return h.conn.Close()
}
