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

		ws.Handle(websocket.EventClose, func(_ websocket.ResponseWriter, r *websocket.Request) {
			fmt.Printf("Server closed with code %d\n", r.CloseCode())
		})
		ws.Handle(websocket.EventError, func(_ websocket.ResponseWriter, r *websocket.Request) {
			fmt.Println(r.Err())
		})
		ws.Handle(websocket.EventMessage, func(w websocket.ResponseWriter, r *websocket.Request) {
			w.SetOpcode(r.Opcode())
			w.Write(r.Bytes())
		})
	}))
	log.Fatal(http.ListenAndServe(":9001", nil))
}
