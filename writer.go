package websocket

import (
	"bufio"
	"encoding/binary"
	"math"

	"github.com/gbrlsnchs/websocket/internal"
)

// Writer is a buffered writer that is able to fragment itself in several
// frames in order to send a message greater than its own internal buffer.
type Writer struct {
	wr     *bufio.Writer
	opcode uint8
	err    error
}

func (w *Writer) Write(b []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	wr := w.wr

	fin := leftBit | w.opcode
	// Check if message needs to be fragmented.
	if bsize, diff := internal.ByteSize(b), wr.Available(); bsize > diff {
		diff -= bsize - len(b)                                // resolve payload length
		fin &= w.opcode                                       // set FIN bit to zero
		next := &Writer{wr: w.wr, opcode: opcodeContinuation} // prepare next frame to be sent
		defer next.Write(b[diff:])                            // schedule next write
		b = b[:diff]                                          // resolve current payload
	}

	if w.err = wr.WriteByte(fin); w.err != nil {
		return 0, w.err
	}

	size := len(b)
	switch {
	case size <= 125:
		if w.err = wr.WriteByte(byte(size)); w.err != nil {
			return 0, w.err
		}
	case size <= math.MaxUint16:
		if w.err = wr.WriteByte(126); w.err != nil {
			return 0, w.err
		}
		if w.err = binary.Write(wr, binary.BigEndian, uint16(size)); w.err != nil {
			return 0, w.err
		}
	default:
		if w.err = wr.WriteByte(127); w.err != nil {
			return 0, w.err
		}
		if w.err = binary.Write(wr, binary.BigEndian, uint64(size)); w.err != nil {
			return 0, w.err
		}
	}
	// Write message and flush.
	if _, w.err = wr.Write(b); w.err != nil {
		return 0, w.err
	}
	n := wr.Buffered()
	if w.err = wr.Flush(); w.err != nil {
		return 0, w.err
	}
	return n, nil
}

func (w *Writer) SetOpcode(opcode uint8) {
	w.opcode = opcode
}
