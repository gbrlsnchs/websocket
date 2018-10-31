package websocket

import "io"

type ResponseWriter interface {
	io.Writer
	SetOpcode(uint8)
}
