package wsocket

type Message []byte

func (m Message) Read(b []byte) (int, error) {
	return copy(b, m), nil
}
