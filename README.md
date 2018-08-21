# wsocket (WebSocket implementation for Go)

## About
This is an easy-to-use WebSocket server implementation in [Go].
It passes [Autobahn Test Suite].

## Examples
### Simple usage inside HTTP handler
```go
ws, err := wsocket.UpgradeHTTP(w, r)
if err != nil {
	// ...
}
// Control frames are handled internally and don't reach this handler.
ws.OnMessage(func(msg wsocket.Message, opcode wsocket.Opcode) {
	w := ws.NewWriter(opcode)
	w.Write(msg) // echo message back

	// Use read/write buffer.
	rw := bufio.NewReadWriter(msg, w)
	// ...
})
```

### Checking errors
```go
ws.OnError(func(err error) {
	log.Print(err)
})
```

### Run function on close
```go
ws.OnClose(func(cc wsocket.CloseCode) {
	log.Printf("server closed with close code %d", cc)
})
```

[Go]: https://golang.org
[Autobahn Test Suite]: https://crossbar.io/autobahn/testsuite/