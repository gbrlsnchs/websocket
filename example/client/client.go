package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gbrlsnchs/websocket"
)

func main() {
	address := flag.String("addr", "ws://echo.websocket.org", "address to connect to")
	skip := flag.Bool("skip", false, "skip sending hello world")
	flag.Parse()

	ws, err := websocket.Open(*address, 15*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if !*skip {
		ws.Write([]byte("Hello, WebSocket!"))
	}

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
		payload, _ := ws.Message()
		fmt.Printf("Message sent from server: %s\n", payload)
	}
	if err := ws.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(ws.CloseCode())
}
