package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gbrlsnchs/websocket"
)

func main() {
	address := flag.String("addr", "ws://echo.websocket.org", "address to connect to")
	flag.Parse()

	ws, err := websocket.Open(*address)
	if err != nil {
		log.Fatal(err)
	}

	ws.Write([]byte("Hello, WebSocket!"))

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			b := scanner.Bytes()
			if string(b) == "close" {
				ws.Close()
				continue
			}
			ws.Write(scanner.Bytes())
		}
	}()
	for ws.Next() {
		fmt.Printf("Message sent from server: %s\n", ws.Payload())
	}
	if err := ws.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(ws.CloseCode())
}
