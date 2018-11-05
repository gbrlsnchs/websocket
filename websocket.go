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

// WebSocket is a websocket instance that may be
// either a client or a server depending on how it is created.
type WebSocket struct {
	*writer

	fb   *frameBuffer
	conn io.Closer

	state int
	cc    uint16

	opcode  uint8
	payload []byte
	err     error
}

func newWS(conn net.Conn, client bool) *WebSocket {
	return &WebSocket{
		fb: &frameBuffer{
			rd:     bufio.NewReaderSize(conn, defaultRWSize),
			first:  true,
			client: client,
		},
		writer: &writer{
			wr:     bufio.NewWriterSize(conn, defaultRWSize),
			opcode: OpcodeText,
			client: client,
		},
		conn: conn,
	}
}

// UpgradeHTTP switches the protocol from HTTP to the WebSocket Protocol.
func UpgradeHTTP(w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	conn, err := internal.Handshake(w, r)
	if err != nil {
		return nil, err
	}
	return newWS(conn, false), nil
}

// Close closes the connection manually by sending the close code 1000.
func (ws *WebSocket) Close() error {
	if ws.cc == 0 {
		ws.cc = 1000
	}
	ws.SetOpcode(opcodeClose)
	binary.Write(ws, binary.BigEndian, ws.cc)

	var err error
	if ws.state >= stateClosing {
		err = ws.conn.Close()
	}
	ws.resolveState()
	return err
}

func (ws *WebSocket) CloseCode() uint16        { return ws.cc }
func (ws *WebSocket) Err() error               { return ws.err }
func (ws *WebSocket) Message() ([]byte, uint8) { return ws.payload, ws.opcode }

func (ws *WebSocket) Next() bool {
	for {
		f, err := ws.fb.next()
		if err != nil {
			ws.state = stateClosed
			if err == io.EOF {
				return false
			}
			ws.conn.Close()
			ws.err = err
			return false
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
				ws.err = errInvalidCloseCode
				return false
			}
			if !utf8.Valid(f.payload) {
				ws.cc = 1002
				ws.err = errInvalidClosePayload
				return false
			}
			ws.opcode = f.opcode
			ws.payload = f.payload
			return false
		default:
			ws.fb.add(f)
			if f.final {
				defer ws.fb.reset()
				if ws.fb.opcode == OpcodeText && !utf8.Valid(ws.fb.payload) {
					ws.conn.Close()
					ws.err = errInvalidUTF8
					return false
				}
				ws.opcode = ws.fb.opcode
				ws.payload = ws.fb.payload
				return true
			}
		}
	}
}

func (ws *WebSocket) Read(b []byte) (int, error) { return copy(b, ws.payload), nil }

func (ws *WebSocket) SetCloseCode(cc uint16) error {
	if !validCloseCode(cc) {
		return errInvalidCloseCode
	}
	ws.cc = cc
	return nil
}

func (ws *WebSocket) SetOpcode(opcode uint8) { ws.writer.opcode = opcode }

func (ws *WebSocket) handlePing(b []byte) {
	ws.SetOpcode(opcodePong)
	ws.Write(b)
}

func (ws *WebSocket) resolveState() {
	switch ws.state {
	case stateOpen:
		ws.state = stateClosing
	case stateClosing:
		ws.state = stateClosed
	}
}
