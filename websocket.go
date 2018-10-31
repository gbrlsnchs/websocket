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

type HandlerFunc func(ResponseWriter, *Request)

var errInvalidCloseCode = errors.New("websocket: invalid close code")

const defaultRWSize = 4096
const (
	stateOpen = iota
	stateClosing
	stateClosed
)

// WebSocket is an websocket instance that is created
// when an HTTP connection is upgraded.
type WebSocket struct {
	conn         net.Conn
	rsize        int
	wsize        int
	pongSize     int
	closeSize    int
	state        int
	cc           CloseCode
	handler      HandlerFunc
	errHandler   HandlerFunc
	closeHandler HandlerFunc
}

// UpgradeHTTP switches the protocol from HTTP to the WebSocket Protocol.
func UpgradeHTTP(w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	conn, err := internal.Handshake(w, r)
	if err != nil {
		return nil, err
	}
	return &WebSocket{
		conn:      conn,
		rsize:     defaultRWSize,
		wsize:     defaultRWSize,
		pongSize:  defaultRWSize,
		closeSize: defaultRWSize,
		cc:        1000, // default close code
	}, nil
}

// Close closes the connection manually by sending the close code 1000.
func (ws *WebSocket) Close() error {
	w := ws.NewWriterSize(ws.closeSize)
	w.SetOpcode(opcodeClose)
	binary.Write(w, binary.BigEndian, ws.cc)

	var err error
	if ws.state >= stateClosing {
		if ws.closeHandler != nil {
			defer ws.closeHandler(nil, &Request{cc: ws.cc})
		}
		err = ws.conn.Close()
	}
	ws.resolveState()
	return err
}

func (ws *WebSocket) Handle(e Event, handler HandlerFunc) {
	switch e {
	case EventClose:
		ws.closeHandler = handler
	case EventError:
		ws.errHandler = handler
	case EventMessage:
		go ws.handleMessage(handler)
	case EventOpen:
	}
}

func (ws *WebSocket) NewWriter() ResponseWriter { return ws.NewWriterSize(ws.wsize) }
func (ws *WebSocket) NewWriterSize(size int) ResponseWriter {
	return &Writer{wr: bufio.NewWriterSize(ws.conn, size)}
}

func (ws *WebSocket) SetBufferSize(rsize, wsize int) { ws.rsize, ws.wsize = rsize, wsize }
func (ws *WebSocket) SetCloseBufferSize(size int)    { ws.closeSize = size }
func (ws *WebSocket) SetCloseCode(cc CloseCode)      { ws.cc = cc }
func (ws *WebSocket) SetPongBufferSize(size int)     { ws.pongSize = size }

func (ws *WebSocket) handleClose(b []byte) {
	switch ws.state {
	case stateOpen:
	case stateClosing:
		ws.conn.Close()
	case stateClosed: // no-op
	}
}

func (ws *WebSocket) handleErr(err error) {
	if ws.errHandler != nil {
		ws.errHandler(nil, &Request{err: err})
	}
}

func (ws *WebSocket) handleMessage(handler HandlerFunc) {
	fb := newFrameBuffer(ws.conn, ws.rsize)
	fb.reset()
	for ws.state != stateClosed {
		f, err := fb.next()
		if err != nil {
			if err != io.EOF {
				ws.handleErr(err)
				ws.conn.Close()
			}
			return
		}

		switch {
		case f.opcode == opcodePing:
			ws.handlePing(f.payload)
		case f.opcode == opcodePong: // no-op
		case f.opcode == opcodeClose:
			ws.resolveState()
			var err error
			if f.hasCloseCode && !f.cc.isValid() {
				err = errInvalidCloseCode
			} else if !utf8.Valid(f.payload) {
				err = errInvalidClosePayload
			}
			if err != nil {
				ws.state = stateClosing
				ws.cc = 1002
			}
			ws.Close()
		default:
			fb.push(f)
			if f.final {
				if fb.opcode == OpcodeText && !utf8.Valid(fb.payload) {
					ws.handleErr(errors.New("websocket: payload contains invalid UTF-8 text"))
					ws.conn.Close()
					return
				}
				r := &Request{
					payload: make([]byte, len(fb.payload)),
					opcode:  fb.opcode,
					cc:      f.cc,
				}
				copy(r.payload, fb.payload)
				handler(ws.NewWriter(), r)
				fb.reset()
			}
		}
	}
}

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
