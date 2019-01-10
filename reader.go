package websocket

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"unicode/utf8"
)

type step int

const (
	stepInfo step = iota // FIN, RSV, opcode, is masked and length
	stepMask
	stepPayload
	stepParse
)

var (
	errFragmentedControlFrame    = errors.New("websocket: fragmented control Frame")
	errInvalidOpcode             = errors.New("websocket: invalid opcode")
	errInvalidContinuationOpcode = errors.New("websocket: invalid opcode for continuation")
	errHeadlessContinuation      = errors.New("websocket: headless continuation")
	errUnmasked                  = errors.New("websocket: unmasked message sent from client")
	errLargeControlFrame         = errors.New("websocket: control Frame with length greater than 125")
	errUnnegotiatedRSV           = errors.New("websocket: unnegotiated RSV bits")
	errInvalidClosePayload       = errors.New("websocket: invalid application data for opcode close")
	errIllegalLength             = errors.New("websocket: illegal length indicator")
)

// Reader is a buffered reader.
type Reader struct {
	buf *bufio.Reader

	step         step   // current step
	opcode       Opcode // assembled frame's opcode
	first        bool   // is first frame?
	frame        frame  // individual frame
	message      []byte // full message
	closePayload []byte

	client bool // is the ws instance in client mode?
}

func (r *Reader) Len() int {
	return len(r.message)
}

func (r *Reader) Opcode() Opcode {
	return r.opcode
}

// Read parses frames to form a message and subsequently reads it.
//
// When a read is successful or the other node sends a close frame, io.EOF is returned.
func (r *Reader) Read(b []byte) (n int, err error) {
	return copy(b, r.message), io.EOF
}
func (r *Reader) nextStep(stepFunc func() error) error {
	r.step++
	return stepFunc()
}

func (r *Reader) read(b []byte) error {
	_, err := io.ReadFull(r.buf, b)
	return err
}

func (r *Reader) readInfo() error {
	b := make([]byte, 2)
	if err := r.read(b); err != nil {
		return err
	}
	r.frame = frame{
		fin:    b[0]&leftBit != 0,
		opcode: Opcode(b[0] & opcodeBits),
		rsv1:   b[0]&rsv1Bit != 0,
		rsv2:   b[0]&rsv2Bit != 0,
		rsv3:   b[0]&rsv3Bit != 0,
		masked: b[1]&leftBit != 0,
		length: b[1] & lengthBits,
	}
	if r.frame.isControl() && !r.frame.fin {
		return errFragmentedControlFrame
	}
	if r.frame.rsv1 || r.frame.rsv2 || r.frame.rsv3 {
		return errUnnegotiatedRSV
	}
	if !r.frame.opcode.isValid() {
		return errInvalidOpcode
	}
	if !r.frame.isControl() && !r.first && r.frame.opcode != opcodeContinuation {
		return errInvalidContinuationOpcode
	}
	if r.first {
		if !r.frame.isControl() {
			r.opcode = r.frame.opcode
		}
		if r.frame.opcode == opcodeContinuation {
			return errHeadlessContinuation
		}
		r.first = false
	}

	// Client messages must always be masked.
	// On the other hand, server messages must never be masked.
	if r.frame.masked == r.client {
		return errUnmasked // TODO: fix error message
	}

	switch {
	case r.frame.length == 0: // no-op
	case r.frame.length <= 125:
		r.frame.payload = make([]byte, r.frame.length)
	case r.frame.length == 126:
		return r.readLength16()
	case r.frame.length == 127:
		return r.readLength64()
	default:
		return errIllegalLength
	}
	return nil
}

func (r *Reader) readLength16() error {
	b := make([]byte, 2)
	if err := r.read(b); err != nil {
		return err
	}
	r.frame.payload = make([]byte, binary.BigEndian.Uint16(b))
	return nil
}

func (r *Reader) readLength64() error {
	b := make([]byte, 8)
	if err := r.read(b); err != nil {
		return err
	}
	r.frame.payload = make([]byte, binary.BigEndian.Uint64(b))
	return nil
}

func (r *Reader) readMask() error {
	if r.client {
		return nil
	}
	r.frame.mask = make(mask, 4)
	if err := r.read(r.frame.mask); err != nil {
		return err
	}
	return nil
}

func (r *Reader) readPayload() error {
	length := len(r.frame.payload)
	if length > 0 {
		if err := r.read(r.frame.payload); err != nil {
			return err
		}
		if !r.client {
			r.frame.transform()
		}
		if r.frame.opcode == opcodeClose {
			// If there's a payload when opcode is close,
			// the first two bytes must be a close code,
			// in other words, a 16-bit unsigned integer.
			if length < 2 {
				return errInvalidClosePayload
			}
			ccbin := r.frame.payload[:2]
			r.frame.cc = CloseCode(binary.BigEndian.Uint16(ccbin))
			if !r.frame.cc.isValid() {
				return errInvalidCloseCode
			}
			r.frame.payload = r.frame.payload[2:]
		}
		if r.frame.opcode == opcodeClose && !utf8.Valid(r.frame.payload) {
			return errInvalidUTF8
		}
	}
	return nil
}

func (r *Reader) reset() {
	r.step = stepInfo
	r.opcode = 0
	r.first = true
	r.frame = frame{}
}
