# websocket (WebSocket implementation for Go)

## About
This is an easy-to-use WebSocket server implementation in [Go](https://golang.org).
It passes [Autobahn Test Suite](https://crossbar.io/autobahn/testsuite/).

## Examples
### Simple usage inside HTTP handler
```go
ws, err := websocket.UpgradeHTTP(w, r)
if err != nil {
	// ...
}
// Control frames are handled internally and don't reach this handler.
ws.Handle(websocket.EventMessage, func(w websocket.ResponseWriter, r *websocket.Request) {
	// Echo message back.
	w.SetOpcode(r.Opcode)
	w.Write(r.Payload)
})
```

### Checking errors
```go
ws.Handle(websocket.EventError, func(_ websocket.ResponseWriter, r *websocket.Request) {
	fmt.Println(r.Err())
})
```

### Run function on close
```go
ws.Handle(websocket.EventClose, func(_ websocket.ResponseWriter, r *websocket.Request) {
	fmt.Printf("server closed with close code %d\n", r.CloseCode)
})
```