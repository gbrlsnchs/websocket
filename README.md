# websocket (WebSocket client and server for Go)

## About
This is an easy-to-use WebSocket client and server implementation in [Go](https://golang.org).
It passes the [Autobahn Test Suite](https://crossbar.io/autobahn/testsuite/).

## Usage
Full documentation [here](https://godoc.org/github.com/gbrlsnchs/websocket).

### Installing
#### Go 1.10
`vgo get -u github.com/gbrlsnchs/websocket`
#### Go 1.11
`go get -u github.com/gbrlsnchs/websocket`

### Importing
```go
import (
	// ...

	"github.com/gbrlsnchs/websocket"
)
```

## Examples
### Upgrading an HTTP request
```go
func upgradingHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.UpgradeHTTP(w, r)
	if err != nil {
		// handle error
	}

	// Handle incoming messages.
	// Control frames are handled internally and don't reach this handler.
	ws.Handle(websocket.EventMessage, func(w websocket.ResponseWriter, r *websocket.Request) {
		// Echo message back.
		w.SetOpcode(r.Opcode)
		w.Write(r.Payload)
	})

	// Handle close event.
	ws.Handle(websocket.EventClose, func(_ websocket.ResponseWriter, r *websocket.Request) {
		fmt.Printf("Server closed with close code %d.\n", r.CloseCode)
	})

	// Handle errors caught while processing a request.
	ws.Handle(websocket.EventError, func(_ websocket.ResponseWriter, r *websocket.Request) {
		fmt.Println(r.Err())
	})
}
```

### Using a client
```go
ws, err := websocket.Open("ws://localhost:9001")
if err != nil {
	// handle error
}

// Handle incoming messages.
// Control frames are handled internally and don't reach this handler.
ws.Handle(websocket.EventMessage, func(w websocket.ResponseWriter, r *websocket.Request) {
	// Echo message back.
	w.SetOpcode(r.Opcode)
	w.Write(r.Payload)
})

// Handle close event.
ws.Handle(websocket.EventClose, func(_ websocket.ResponseWriter, r *websocket.Request) {
	fmt.Printf("Server closed with close code %d.\n", r.CloseCode)
})

// Handle errors caught while processing a request.
ws.Handle(websocket.EventError, func(_ websocket.ResponseWriter, r *websocket.Request) {
	fmt.Println(r.Err())
})

w := ws.NewWriter()
w.SetOpcode(websocket.OpcodeText)
w.Write([]byte("Hello, WebSocket!")
```

## Contributing
### How to help
- For bugs and opinions, please [open an issue](https://github.com/gbrlsnchs/websocket/issues/new)
- For pushing changes, please [open a pull request](https://github.com/gbrlsnchs/websocket/compare)