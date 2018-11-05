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
		payload, opcode := ws.Message()
		ws.SetOpcode(opcode)
		ws.Write(payload)
	}
	if err := ws.Err(); err != nil {
		fmt.Println(err)
	}
	fmt.Println(ws.CloseCode())
}
```

### Openning connection to a WebSocket server (client mode)
```go
ws, err := websocket.Open("ws://echo.websocket.org", 15*time.Second)
if err != nil {
	// handle error
}

ws.Write([]byte("Hello, WebSocket!"))

for ws.Next() {
	payload, _ := ws.Message()
	fmt.Printf("Message sent by server: %s\n", payload)
}
if err := ws.Err(); err != nil {
	fmt.Println(err)
}
fmt.Println(ws.CloseCode())
```

## Contributing
### How to help
- For bugs and opinions, please [open an issue](https://github.com/gbrlsnchs/websocket/issues/new)
- For pushing changes, please [open a pull request](https://github.com/gbrlsnchs/websocket/compare)
