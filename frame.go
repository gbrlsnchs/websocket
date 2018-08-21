package wsocket

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
)

var (
	errFragmentedControlFrame    = errors.New("wsocket: fragmented control frame")
	errInvalidOpcode             = errors.New("wsocket: invalid opcode")
	errInvalidContinuationOpcode = errors.New("wsocket: invalid opcode for continuation")
	errHeadlessContinuation      = errors.New("wsocket: headless continuation")
	errUnmasked                  = errors.New("wsocket: unmasked message sent from client")
	errLargeControlFrame         = errors.New("wsocket: control frame with length greater than 125")
	errUnnegotiatedRSV           = errors.New("wsocket: unnegotiated RSV bits")
	errInvalidClosePayload       = errors.New("wsocket: invalid application data for opcode close")
	errIllegalLength             = errors.New("wsocket: illegal length indicator")
)

type Opcode byte

const (
	opcodeContinuation Opcode = iota // opcodeNone is for internal usage
	OpcodeText
	OpcodeBinary
	_
	_
	_
	_
	_
	opcodeClose
	opcodePing
	opcodePong
)

type frame struct {
	final        bool
	opcode       Opcode
	payload      Message
	hasCloseCode bool
}

func nextFrame(rd *bufio.Reader, lastFinal bool, fsOpcode Opcode) (f *frame, closeCode CloseCode, err error) {
	var b byte

	// First byte (FIN + RSV + opcode).
	if b, err = rd.ReadByte(); err != nil {
		return nil, 0, err
	}
	fin, rsv, opcode := b&byte(0x80), b&byte(0x70), Opcode(b&byte(0xF))
	switch {
	case opcode >= opcodeClose && fin == 0:
		return nil, 0, errFragmentedControlFrame
	case rsv > 0:
		return nil, 0, errUnnegotiatedRSV
	case opcode < opcodeContinuation || opcode > OpcodeBinary && opcode < opcodeClose || opcode > opcodePong:
		return nil, 0, errInvalidOpcode
	// Last frame is not final, current is neither continuation nor is a control frame.
	case !lastFinal && opcode > opcodeContinuation && opcode < opcodeClose:
		return nil, 0, errInvalidContinuationOpcode
	// Frame stack opcode is neither text nor binary.
	case fsOpcode == opcodeContinuation && opcode == opcodeContinuation:
		return nil, 0, errHeadlessContinuation
	}

	// Second byte (mask flag, length indicator).
	if b, err = rd.ReadByte(); err != nil {
		return nil, 0, err
	}
	masked, length := b&byte(0x80), b&byte(0x7F)
	if masked == 0 {
		return nil, 0, errUnmasked
	}
	if opcode >= opcodeClose && length > 125 {
		return nil, 0, errLargeControlFrame
	}

	f = &frame{final: fin > 0, opcode: opcode}
	switch {
	case length == 0: // no-op
	case length > 0 && length <= 125:
		f.payload = make([]byte, length)
	case length == 126:
		b := make([]byte, byte(2))
		if _, err = io.ReadFull(rd, b); err != nil {
			return nil, 0, err
		}
		f.payload = make([]byte, binary.BigEndian.Uint16(b))
	case length == 127:
		b := make([]byte, byte(8))
		if _, err := io.ReadFull(rd, b); err != nil {
			return nil, 0, err
		}
		f.payload = make([]byte, binary.BigEndian.Uint64(b))
	default:
		return nil, 0, errIllegalLength
	}

	mask := make([]byte, byte(4))
	if _, err = io.ReadFull(rd, mask); err != nil {
		return nil, 0, err
	}
	if len(f.payload) > 0 {
		if _, err = io.ReadFull(rd, f.payload); err != nil {
			return nil, 0, err
		}
		f.decode(mask)
		if f.opcode == opcodeClose {
			if len(f.payload) < 2 {
				return nil, 0, errInvalidClosePayload
			}
			closeCode = CloseCode(binary.BigEndian.Uint16(f.payload[:2]))
			f.hasCloseCode = true
			f.payload = f.payload[2:]
		}
	}
	return f, closeCode, nil
}

func (f *frame) decode(mask []byte) {
	for i := range f.payload {
		f.payload[i] ^= mask[i%4]
	}
}
