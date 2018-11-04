package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gbrlsnchs/websocket"
)

func main() {
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "github.com/gbrlsnchs/websocket")
		ws, err := websocket.UpgradeHTTP(w, r)
		if err != nil {
			fmt.Println(err)
			return
		}

		for ws.Next() {
			ws.SetWriteOpcode(ws.Opcode())
			ws.Write(ws.Payload())
		}
		if err = ws.Err(); err != nil {
			fmt.Println(err)
		}
		fmt.Println(ws.CloseCode())
	}))
	log.Fatal(http.ListenAndServe(":9001", nil))
}
