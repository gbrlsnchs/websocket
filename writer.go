package websocket

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"io"
	"math"
)

// Writer is a buffered writer that is able to fragment itself in several
// frames in order to send a message greater than its own internal buffer.
type Writer struct {
	wr     *bufio.Writer
	opcode uint8
	err    error
	client bool
}

func (w *Writer) SetOpcode(opcode uint8) {
	w.opcode = opcode
}

func (w *Writer) Write(b []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	wr := w.wr

	fin := leftBit | w.opcode
	// Check if message needs to be fragmented.
	if bsize, diff := w.byteSize(b), wr.Available(); bsize > diff {
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
	var maskedBit uint8
	if w.client {
		maskedBit = leftBit
	}
	switch {
	case size <= 125:
		if w.err = wr.WriteByte(uint8(size) | maskedBit); w.err != nil {
			return 0, w.err
		}
	case size <= math.MaxUint16:
		if w.err = wr.WriteByte(126 | maskedBit); w.err != nil {
			return 0, w.err
		}
		if w.err = binary.Write(wr, binary.BigEndian, uint16(size)); w.err != nil {
			return 0, w.err
		}
	default:
		if w.err = wr.WriteByte(127 | maskedBit); w.err != nil {
			return 0, w.err
		}
		if w.err = binary.Write(wr, binary.BigEndian, uint64(size)); w.err != nil {
			return 0, w.err
		}
	}

	// Mask the payload.
	if w.client {
		m := make(mask, 4)
		if _, w.err = io.ReadFull(rand.Reader, m); w.err != nil {
			return 0, w.err
		}
		if _, w.err = wr.Write(m); w.err != nil {
			return 0, w.err
		}
		m.transform(b)
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

func (w *Writer) byteSize(b []byte) (size int) {
	size++ // FIN

	length := len(b)
	switch {
	case length <= 125:
		size++ // indicator is the current length
	case length <= math.MaxUint16:
		size += 3 // indicator + 2 bytes for length value
	default:
		size += 9 // indicator + 8 bytes for length value
	}
	if w.client {
		size += 4 // 4 bytes for masking-key
	}
	return size + length
}
