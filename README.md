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
### Upgrading an HTTP request and listening to messages
```go
func upgradingHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.UpgradeHTTP(w, r)
	if err != nil {
		// handle error
	}

	for ws.Next() {
		ws.SetWriteOpcode(ws.Opcode())
		ws.Write(ws.Payload())
	}
	if err := ws.Err(); err != nil {
		fmt.Println(err)
	}
	fmt.Println(ws.CloseCode())
}
```

### Openning connection to a WebSocket server (client mode)
```go
ws, err := websocket.Open("ws://echo.websocket.org")
if err != nil {
	// handle error
}

ws.Write([]byte("Hello, WebSocket!"))

for ws.Next() {
	fmt.Printf("Message sent from server: %s\n", ws.Payload())
}
if err := ws.Err(); err != nil {
	fmt.Println(err)
	os.Exit(1)
}
fmt.Println(ws.CloseCode())
```

## Contributing
### How to help
- For bugs and opinions, please [open an issue](https://github.com/gbrlsnchs/websocket/issues/new)
- For pushing changes, please [open a pull request](https://github.com/gbrlsnchs/websocket/compare)
