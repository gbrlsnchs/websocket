package wsocket

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/gbrlsnchs/wsocket/internal"
)

type Writer struct {
	w      io.Writer
	size   int
	opcode Opcode
	cc     CloseCode
	err    error
}

func newWriter(w io.Writer, size int, opcode Opcode) *Writer {
	return &Writer{
		w:      w,
		size:   size,
		opcode: opcode,
	}
}

func (w *Writer) Write(b []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	wr := bufio.NewWriterSize(w.w, w.size)
	fin := byte(0x80 | w.opcode)
	if bsize, diff := internal.ByteSize(b), wr.Available(); bsize > diff { // check if message needs to be fragmented
		diff -= bsize - len(b)                             // resolve payload length
		fin &= byte(w.opcode)                              // set FIN bit to zero
		next := newWriter(w.w, w.size, opcodeContinuation) // prepare next frame to be sent
		defer next.Write(b[diff:])                         // schedule next write
		b = b[:diff]                                       // resolve current payload
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
	case size <= int(^uint16(0)):
		if w.err = wr.WriteByte(byte(126)); w.err != nil {
			return 0, w.err
		}
		if w.err = binary.Write(wr, binary.BigEndian, uint16(size)); w.err != nil {
			return 0, w.err
		}
	default:
		if w.err = wr.WriteByte(byte(127)); w.err != nil {
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

func (w *Writer) WriteByte(b byte) error {
	_, err := w.Write([]byte{b})
	return err
}

func (w *Writer) WriteRune(r rune) (int, error) {
	return w.Write([]byte(string(r)))
}

func (w *Writer) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}
