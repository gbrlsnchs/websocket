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
	ww := ws.NewWriter() // one buffered writer is enough

	for ws.IsOpen() {
		b, opcode, err := ws.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}
		ww.SetOpcode(opcode)
		ww.Write(b)
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

w := ws.NewWriter()
w.Write([]byte("Hello, WebSocket!"))

for ws.IsOpen() {
	b, _, err := ws.Accept()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Message sent from server: %s\n", b)
}
fmt.Println(ws.CloseCode())
```

## Contributing
### How to help
- For bugs and opinions, please [open an issue](https://github.com/gbrlsnchs/websocket/issues/new)
- For pushing changes, please [open a pull request](https://github.com/gbrlsnchs/websocket/compare)
