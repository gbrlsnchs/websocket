package websocket

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"unicode/utf8"

	"github.com/gbrlsnchs/websocket/internal"
)

var (
	errInvalidCloseCode = errors.New("websocket: invalid close code")
	errInvalidUTF8      = errors.New("websocket: payload contains invalid UTF-8 content")
)

const defaultRWSize = 4096
const (
	stateOpen = iota
	stateClosing
	stateClosed
)

// WebSocket is a websocket instance that is created
// when an HTTP connection is upgraded.
type WebSocket struct {
	conn      net.Conn
	fb        *frameBuffer
	wsize     int
	pongSize  int
	closeSize int
	state     int
	cc        uint16
	client    bool
}

func newWS(conn net.Conn, client bool) *WebSocket {
	return &WebSocket{
		conn:      conn,
		fb:        fb,
		wsize:     defaultRWSize,
		pongSize:  defaultRWSize,
		closeSize: defaultRWSize,
		client:    client,
	}
}

// UpgradeHTTP switches the protocol from HTTP to the WebSocket Protocol.
func UpgradeHTTP(w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	conn, err := internal.Handshake(w, r)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return nil, err
	}
	return newWebSocket(conn, false), nil
}

func (ws *WebSocket) Accept() ([]byte, uint8, error) {
	for {
		f, err := ws.fb.next()
		if err != nil {
			ws.state = stateClosed
			if err == io.EOF {
				return nil, 0, nil
			}
			ws.conn.Close()
			return nil, 0, err
		}

		switch {
		case f.opcode == opcodePing:
			ws.handlePing(f.payload)
		case f.opcode == opcodePong: // no-op
		case f.opcode == opcodeClose:
			defer ws.Close()
			ws.cc = 1000
			ws.resolveState()
			if f.hasCloseCode && !validCloseCode(f.cc) {
				ws.cc = 1002
				return nil, 0, errInvalidCloseCode
			}
			if !utf8.Valid(f.payload) {
				ws.cc = 1002
				return nil, 0, errInvalidClosePayload
			}
			b := make([]byte, len(ws.fb.payload))
			copy(b, ws.fb.payload)
			return b, f.opcode, nil
		default:
			ws.fb.add(f)
			if f.final {
				defer ws.fb.reset()
				if ws.fb.opcode == OpcodeText && !utf8.Valid(ws.fb.payload) {
					ws.conn.Close()
					return nil, 0, errInvalidUTF8
				}
				b := make([]byte, len(ws.fb.payload))
				copy(b, ws.fb.payload)
				return b, ws.fb.opcode, nil
			}
		}
	}
}

// Close closes the connection manually by sending the close code 1000.
func (ws *WebSocket) Close() error {
	if ws.cc == 0 {
		ws.cc = 1000
	}
	w := ws.NewWriterSize(ws.closeSize)
	w.SetOpcode(opcodeClose)
	binary.Write(w, binary.BigEndian, ws.cc)

	var err error
	if ws.state >= stateClosing {
		err = ws.conn.Close()
	}
	ws.resolveState()
	return err
}

func (ws *WebSocket) CloseCode() uint16 {
	return ws.cc
}

func (ws *WebSocket) IsOpen() bool {
	return ws.state != stateClosed
}

func (ws *WebSocket) NewWriter() *Writer { return ws.NewWriterSize(ws.wsize) }
func (ws *WebSocket) NewWriterSize(size int) *Writer {
	return &Writer{wr: bufio.NewWriterSize(ws.conn, size), opcode: OpcodeText, client: ws.client}
}

func (ws *WebSocket) SetBufferSize(rsize, wsize int) {
	ws.fb.rd = bufio.NewReaderSize(ws.conn, rsize)
	ws.wsize = wsize
}

func (ws *WebSocket) SetCloseBufferSize(size int) { ws.closeSize = size }

func (ws *WebSocket) SetCloseCode(cc uint16) error {
	if !validCloseCode(cc) {
		return errInvalidCloseCode
	}
	ws.cc = cc
	return nil
}

func (ws *WebSocket) SetPongBufferSize(size int) { ws.pongSize = size }

func (ws *WebSocket) handlePing(b []byte) {
	w := ws.NewWriterSize(ws.pongSize)
	w.SetOpcode(opcodePong)
	w.Write(b)
}

func (ws *WebSocket) resolveState() {
	switch ws.state {
	case stateOpen:
		ws.state = stateClosing
	case stateClosing:
		ws.state = stateClosed
	}
}
