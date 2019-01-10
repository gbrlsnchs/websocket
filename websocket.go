package websocket

import (
	"bufio"
	"errors"
	"io"
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
	leftBit    = 0x80
	rsv1Bit    = 0x40
	rsv2Bit    = 0x20
	rsv3Bit    = 0x10
	opcodeBits = 0xF
	lengthBits = 0x7F
)

// WebSocket is a websocket instance that may be
// either a client or a server depending on how it is created.
type WebSocket struct {
	rbuf *bufio.Reader
	wbuf *bufio.Writer
	conn io.ReadWriteCloser

	payload []byte // close payload
	cc      CloseCode
	client  bool
	err     error
}

func newWS(rbuf *bufio.Reader, conn io.ReadWriteCloser, client bool) *WebSocket {
	return &WebSocket{
		rbuf:   rbuf,
		conn:   conn,
		client: client,
	}
}

// UpgradeHTTP switches the protocol from HTTP to the WebSocket Protocol.
func UpgradeHTTP(w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	conn, err := internal.Handshake(w, r)
	if err != nil {
		return nil, err
	}
	rbuf := bufio.NewReaderSize(conn, defaultRWSize)
	return newWS(rbuf, conn, false), nil
}

func (ws *WebSocket) CloseCode() CloseCode {
	return ws.cc
}

func (ws *WebSocket) Err() error {
	return ws.err
}

func (ws *WebSocket) NewReader() *Reader {
	return &Reader{
		buf:    ws.rbuf,
		first:  true,
		client: ws.client,
	}
}

func (ws *WebSocket) NewWriter() *Writer {
	if ws.wbuf == nil {
		ws.wbuf = bufio.NewWriterSize(ws.conn, defaultRWSize)
	}
	return &Writer{
		conn:   ws.conn,
		opcode: OpcodeText,
		buf:    ws.wbuf,
		client: ws.client,
	}
}

func (ws *WebSocket) Next(rd *Reader, wr *Writer) bool {
	for {
		switch rd.step {
		case stepInfo:
			ws.err = rd.nextStep(rd.readInfo)
		case stepMask:
			ws.err = rd.nextStep(rd.readMask)
		case stepPayload:
			ws.err = rd.nextStep(rd.readPayload)
		default:
			rd.step = stepInfo

			if rd.frame.isControl() {
				switch rd.frame.opcode {
				case opcodeClose:
					wr.resolveState()
					wr.opcode = opcodeClose
					wr.Close()
					ws.cc = rd.frame.cc
					if wr.state != stateClosed {
						if ws.cc > 0 && !utf8.Valid(rd.frame.payload) {
							ws.err = errInvalidUTF8
							wr.Close()
							return false
						}
						ws.payload = rd.frame.payload
						return true
					}
					return false
				case opcodePing:
					wr.opcode = opcodePong
					wr.Write(rd.frame.payload)
				case opcodePong: // no-op
				}
				continue
			}
			b := make([]byte, len(rd.message)+len(rd.frame.payload))
			n := copy(b, rd.message)
			copy(b[n:], rd.frame.payload)
			rd.message = b
			if !rd.frame.fin {
				continue
			}
			if wr.opcode == OpcodeText && !utf8.Valid(rd.frame.payload) {
				ws.err = errInvalidUTF8
				wr.Close()
				return false
			}
			return true
		}
		if ws.err != nil {
			return false
		}
	}
}
