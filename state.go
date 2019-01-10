package websocket

type state int

const (
	stateOpen = iota
	stateClosing
	stateClosed
)
