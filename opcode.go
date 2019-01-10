package websocket

type Opcode uint8

const (
	OpcodeText   Opcode = 0x1
	OpcodeBinary Opcode = 0x2

	opcodeContinuation Opcode = 0x0
	opcodeClose        Opcode = 0x8
	opcodePing         Opcode = 0x9
	opcodePong         Opcode = 0xA
)

func (o Opcode) isValid() bool {
	return o >= opcodeContinuation && o <= OpcodeBinary ||
		o >= opcodeClose && o <= opcodePong
}
