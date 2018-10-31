package websocket

const (
	OpcodeText   = 0x1
	OpcodeBinary = 0x2

	opcodeContinuation = 0x0
	opcodeClose        = 0x8
	opcodePing         = 0x9
	opcodePong         = 0xA
)

type frame struct {
	final        bool
	opcode       uint8
	payload      []byte
	cc           CloseCode
	hasCloseCode bool
}
