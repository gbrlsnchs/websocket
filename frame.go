package websocket

type frame struct {
	fin     bool
	opcode  Opcode
	rsv1    bool
	rsv2    bool
	rsv3    bool
	masked  bool
	length  uint8 // length indicator
	mask    mask
	payload []byte
	cc      CloseCode
}

func (f frame) isControl() bool {
	return f.opcode >= opcodeClose
}

func (f *frame) transform() {
	f.mask.transform(f.payload)
}
