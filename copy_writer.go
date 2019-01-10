package websocket

type copyWriter struct {
	w *Writer
}

func (cw *copyWriter) Write(b []byte) (int, error) {
	if cw.w != nil {
		return cw.w.Write(b)
	}
	return 0, nil
}
