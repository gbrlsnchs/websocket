package wsocket

import (
	"errors"
	"unicode/utf8"
)

type frameStack struct {
	opcode  Opcode
	payload Message
	ok      bool
}

func newFrameStack() *frameStack {
	return &frameStack{payload: make(Message, 0)}
}

func (fs *frameStack) done() bool {
	return fs.ok
}

func (fs *frameStack) reset() {
	*fs = *newFrameStack()
}

func (fs *frameStack) frame() (*frame, error) {
	if fs.opcode == OpcodeText && !utf8.Valid(fs.payload) {
		return nil, errors.New("wsocket: payload contains invalid UTF-8 text")
	}
	f := &frame{final: true, opcode: fs.opcode, payload: fs.payload}
	fs.reset()
	return f, nil
}

func (fs *frameStack) push(f *frame) {
	fs.payload = append(fs.payload, f.payload...)
	fs.ok = f.final
	if fs.opcode == 0 {
		fs.opcode = f.opcode
	}
}
