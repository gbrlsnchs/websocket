package websocket

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"unicode/utf8"
)

const (
	leftBit    = 0x80
	rsvBits    = 0x70
	opcodeBits = 0xF
	lengthBits = 0x7F
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

// frameBuffer is a sequence of frames buffered in a stack.
type frameBuffer struct {
	done    bool
	first   bool
	opcode  uint8
	payload []byte
	rd      *bufio.Reader
	client  bool
}

func newFrameBuffer(r io.Reader, size int) *frameBuffer {
	return &frameBuffer{rd: bufio.NewReaderSize(r, size)}
}

// Bytes returns the internal payload that was buffered
// from the first frame until the final frame sequence.
func (fb *frameBuffer) Bytes() []byte {
	return fb.payload
}

// Len returns the payload byte size.
func (fb *frameBuffer) Len() int {
	return len(fb.payload)
}

// Opcode returns whether the frame is binary or text.
func (fb *frameBuffer) Opcode() uint8 {
	return fb.opcode
}

func (fb *frameBuffer) Read(b []byte) (int, error) {
	return copy(b, fb.payload), nil
}

func (fb *frameBuffer) add(f *frame) {
	if fb.first {
		fb.opcode = f.opcode
		fb.first = false
	}
	fb.payload = append(fb.payload, f.payload...)
	fb.done = f.final
}

func (fb *frameBuffer) next() (*frame, error) {
	var (
		b   byte
		err error
		rd  = fb.rd
	)
	// Check FIN, RSV 1 to 3 and the opcode.
	if b, err = rd.ReadByte(); err != nil {
		return nil, err
	}
	fin, rsv, opcode := b&leftBit, b&rsvBits, b&opcodeBits
	// Validate first byte.
	switch {
	case fin == 0 && opcode >= opcodeClose:
		return nil, errFragmentedControlFrame
	case rsv > 0:
		return nil, errUnnegotiatedRSV
	case opcode < opcodeContinuation ||
		opcode > OpcodeBinary && opcode < opcodeClose ||
		opcode > opcodePong:
		return nil, errInvalidOpcode
		// Previous frame is not final, current is neither continuation nor is a control frame.
	case !fb.first &&
		opcode > opcodeContinuation &&
		opcode < opcodeClose:
		return nil, errInvalidContinuationOpcode
	case fb.opcode == opcodeContinuation && opcode == opcodeContinuation:
		return nil, errHeadlessContinuation
	}

	if b, err = rd.ReadByte(); err != nil {
		return nil, err
	}
	masked, length := b&leftBit, int(b&lengthBits)
	if masked == 0 && !fb.client {
		return nil, errUnmasked
	}
	if opcode >= opcodeClose && length > 125 {
		return nil, errLargeControlFrame
	}

	// Read the payload according to the length indicator:
	// 0 until 125 is the literal length.
	// 126 means the length is indicated by an unsigned 16-bit integer.
	// 127 means the length is indicated by an unsigned 64-bit integer.
	var payload []byte
	switch {
	case length == 0: // no-op
	case length > 0 && length <= 125:
		payload = make([]byte, length)
	case length == 126:
		b := make([]byte, 2)
		if _, err = io.ReadFull(rd, b); err != nil {
			return nil, err
		}
		payload = make([]byte, binary.BigEndian.Uint16(b))
	case length == 127:
		b := make([]byte, 8)
		if _, err = io.ReadFull(rd, b); err != nil {
			return nil, err
		}
		payload = make([]byte, binary.BigEndian.Uint64(b))
	default:
		return nil, errIllegalLength
	}

	var m mask
	if !fb.client {
		// Unmask the payload.
		m = make(mask, 4)
		if _, err = io.ReadFull(rd, m); err != nil {
			return nil, err
		}
	}
	length = len(payload)
	f := &frame{
		final:   fin != 0,
		opcode:  opcode,
		payload: payload,
	}
	if length > 0 {
		if _, err = io.ReadFull(rd, payload); err != nil {
			return nil, err
		}
		if !fb.client {
			// Decode the payload according to the RFC 6455.
			m.transform(payload)
		}
		// Read close data if there's any.
		if opcode == opcodeClose {
			if length < 2 {
				return nil, errInvalidClosePayload
			}
			f.hasCloseCode = true
			f.cc = CloseCode(binary.BigEndian.Uint16(payload[:2]))
			f.payload = payload[2:]
		}
	}
	return f, nil
}

func (fb *frameBuffer) reset() {
	fb.first = true
	fb.done = false
	fb.opcode = 0
	fb.payload = nil
}

func (fb *frameBuffer) validate() error {
	if fb.opcode == OpcodeText && !utf8.Valid(fb.payload) {
		return errors.New("websocket: payload contains invalid UTF-8 text")
	}
	return nil
}
