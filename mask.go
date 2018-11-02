package websocket

type mask []byte

func (m mask) transform(b []byte) {
	for i := range b {
		b[i] ^= m[i%4]
	}
}
