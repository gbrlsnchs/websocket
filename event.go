package websocket

// Event is a websocket event.
type Event int

const (
	EventClose Event = iota
	EventError
	EventMessage
	EventOpen
)
