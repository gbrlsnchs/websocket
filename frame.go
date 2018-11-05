package websocket

const (
	OpcodeText   uint8 = 0x1
	OpcodeBinary uint8 = 0x2

	opcodeContinuation uint8 = 0x0
	opcodeClose        uint8 = 0x8
	opcodePing         uint8 = 0x9
	opcodePong         uint8 = 0xA
)

type frame struct {
	final        bool
	opcode       uint8
	payload      []byte
	cc           uint16
	hasCloseCode bool
}
