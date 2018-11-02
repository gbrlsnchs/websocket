package websocket

type Request struct {
	Payload   []byte
	Opcode    uint8
	CloseCode CloseCode
	err       error
}

func (r *Request) Err() error {
	return r.err
}
