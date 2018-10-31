package websocket

type Request struct {
	payload []byte
	opcode  uint8
	cc      CloseCode
	err     error
}

func (r *Request) Bytes() []byte {
	return r.payload
}

func (r *Request) CloseCode() CloseCode {
	return r.cc
}

func (r *Request) Err() error {
	return r.err
}

func (r *Request) Opcode() uint8 {
	return r.opcode
}
