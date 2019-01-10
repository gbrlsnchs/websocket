package websocket

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"io"
	"math"
)

// Writer is a buffered frame writer.
type Writer struct {
	conn io.WriteCloser
	buf  *bufio.Writer

	opcode Opcode
	cc     CloseCode

	state state

	client bool
}

func (w *Writer) ReadFrom(r io.Reader) (int64, error) {
	cw := &copyWriter{w}
	if rr, ok := r.(*Reader); ok {
		w.SetOpcode(rr.opcode)
		if length := rr.Len(); length > 0 {
			return io.CopyBuffer(cw, rr, make([]byte, length))
		}
		if _, err := w.Write(nil); err != nil {
			return 0, err
		}
		return 0, nil
	}
	w.SetOpcode(OpcodeText)
	return io.Copy(cw, r)
}

// Close closes the connection.
func (w *Writer) Close() error {
	if w.cc == 0 {
		w.cc = 1000
	}
	w.opcode = opcodeClose
	binary.Write(w, binary.BigEndian, w.cc)

	var err error
	if w.state == stateClosing {
		err = w.conn.Close()
	}
	w.resolveState()
	return err
}

func (w *Writer) SetCloseCode(cc CloseCode) error {
	if !cc.isValid() {
		return errInvalidCloseCode
	}
	w.cc = cc
	return nil
}

func (w *Writer) SetOpcode(opcode Opcode) error {
	if !opcode.isValid() {
		return errInvalidOpcode
	}
	w.opcode = opcode
	return nil
}

// Write builds a WebSocket frame to a buffered writer
// and flushes the frame when it finishes building it.
//
// If the message size is greater than the buffer size,
// the writer automatically prepares and sends fragmented frames.
func (w *Writer) Write(b []byte) (int, error) {
	w.buf.Reset(w.conn)
	var (
		next        *Writer // next frame
		nextPayload []byte
		fin         = uint8(leftBit | w.opcode)
		err         error
	)
	// Check if message needs to be fragmented.
	if bsize, diff := w.byteSize(b), w.buf.Available(); bsize > diff {
		diff -= bsize - len(b) // resolve payload length
		fin &= uint8(w.opcode) // set FIN bit to zero
		next = &Writer{        // prepare next frame to be sent
			buf:    w.buf,
			conn:   w.conn,
			state:  w.state,
			opcode: opcodeContinuation,
			client: w.client,
		}
		nextPayload = b[diff:]
		// defer next.Write(b[diff:]) // schedule next write
		b = b[:diff] // resolve current payload
	}

	if err = w.buf.WriteByte(fin); err != nil {
		return 0, err
	}

	size := len(b)
	var maskedBit uint8
	if w.client {
		maskedBit = leftBit
	}
	switch {
	case size <= 125:
		if err = w.buf.WriteByte(uint8(size) | maskedBit); err != nil {
			return 0, err
		}
	case size <= math.MaxUint16:
		if err = w.buf.WriteByte(126 | maskedBit); err != nil {
			return 0, err
		}
		if err = binary.Write(w.buf, binary.BigEndian, uint16(size)); err != nil {
			return 0, err
		}
	default:
		if err = w.buf.WriteByte(127 | maskedBit); err != nil {
			return 0, err
		}
		if err = binary.Write(w.buf, binary.BigEndian, uint64(size)); err != nil {
			return 0, err
		}
	}

	// Mask the payload.
	if w.client {
		m := make(mask, 4)
		if _, err = io.ReadFull(rand.Reader, m); err != nil {
			return 0, err
		}
		if _, err = w.buf.Write(m); err != nil {
			return 0, err
		}
		m.transform(b)
	}

	// Write message and flush.
	var n int
	if n, err = w.buf.Write(b); err != nil {
		return 0, err
	}
	if err = w.buf.Flush(); err != nil {
		return 0, err
	}

	if next != nil {
		m, err := next.Write(nextPayload)
		if err != nil {
			return n, err
		}
		n += m
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

func (w *Writer) resolveState() {
	switch w.state {
	case stateOpen:
		w.state = stateClosing
	case stateClosing:
		w.state = stateClosed
	}
}
