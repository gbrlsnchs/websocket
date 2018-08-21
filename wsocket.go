package wsocket

import (
	"bufio"
	"encoding/binary"
	"errors"
	"net"
	"net/http"
	"unicode/utf8"

	"github.com/gbrlsnchs/wsocket/internal"
)

var errInvalidCloseCode = errors.New("wsocket: invalid close code")

const (
	stateOpen byte = iota
	stateClosing
	stateClosed
	defaultRWSize = 4096
)

type WebSocket struct {
	conn         net.Conn
	rsize        int
	wsize        int
	pongSize     int
	closeSize    int
	state        byte
	fs           *frameStack
	errHandler   func(error)
	closeHandler func(CloseCode)
}

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
		fs:        newFrameStack(),
	}, nil
}

func (ws *WebSocket) Close() error {
	w := ws.NewWriterSize(opcodeClose, ws.closeSize)
	w.cc = 1000
	return ws.close(w)
}

func (ws *WebSocket) OnClose(handler func(c CloseCode)) {
	ws.closeHandler = handler
}

func (ws *WebSocket) OnError(handler func(error)) {
	ws.errHandler = handler
}

func (ws *WebSocket) OnMessage(handler func(msg Message, opcode Opcode)) {
	var err error
	go func() {
		var (
			cc        CloseCode
			f         *frame
			lastFinal = true
			rd        = bufio.NewReaderSize(ws.conn, ws.rsize)
		)
		for ws.state != stateClosed {
			f, cc, err = nextFrame(rd, lastFinal, ws.fs.opcode)
			if err != nil {
				ws.handleErr(nil, err)
				return
			}
			switch {
			case f.opcode == opcodePing:
				w := ws.NewWriterSize(opcodePong, ws.pongSize)
				w.Write(f.payload)
			case f.opcode == opcodePong: // no-op
			case f.opcode == opcodeClose:
				var w *Writer
				if ws.state == stateOpen {
					w = ws.NewWriterSize(opcodeClose, ws.closeSize)
					w.cc = 1000
					ws.state = stateClosing
				} else {
					w = ws.NewWriterSize(opcodeClose, ws.closeSize)
					ws.state = stateClosed
				}
				var err error
				if f.hasCloseCode && !cc.isValid() {
					err = errInvalidCloseCode
				} else if !utf8.Valid(f.payload) {
					err = errInvalidClosePayload
				}
				if err != nil {
					if w == nil {
						w = ws.NewWriterSize(opcodeClose, ws.closeSize)
					}
					w.cc = 1002
					ws.handleErr(w, err)
					continue
				}
				ws.close(w)
			default:
				lastFinal = f.final
				ws.fs.push(f)
				if ws.fs.done() {
					var err error
					f, err = ws.fs.frame()
					if err != nil {
						ws.handleErr(nil, err)
						return
					}
					handler(f.payload, f.opcode)
				}
			}
		}
		ws.close(nil)
		if ws.closeHandler != nil {
			ws.closeHandler(cc)
		}
	}()
}

func (ws *WebSocket) NewWriter(opcode Opcode) *Writer {
	return ws.NewWriterSize(opcode, ws.wsize)
}

func (ws *WebSocket) NewWriterSize(opcode Opcode, size int) *Writer {
	return newWriter(ws.conn, size, opcode)
}

func (ws *WebSocket) SetBufferSize(rsize, wsize int) {
	ws.rsize, ws.wsize = rsize, wsize
}

func (ws *WebSocket) SetCloseBufferSize(size int) {
	ws.closeSize = size
}

func (ws *WebSocket) SetPongBufferSize(size int) {
	ws.pongSize = size
}

func (ws *WebSocket) close(w *Writer) error {
	if w != nil { // closing handshake initiated by server
		binary.Write(w, binary.BigEndian, w.cc)
		if ws.state == stateOpen {
			ws.state = stateClosing
			return nil
		}
		if ws.state == stateClosing {
			ws.state = stateClosed
		}
	}
	return ws.conn.Close()
}

func (ws *WebSocket) handleErr(w *Writer, err error) {
	if ws.errHandler != nil {
		ws.errHandler(err)
	}
	ws.close(w)
}
