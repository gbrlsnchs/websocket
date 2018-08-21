package main

import (
	"log"
	"net/http"

	"github.com/gbrlsnchs/wsocket"
)

func main() {
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "github.com/gbrlsnchs/wsocket")
		ws, err := wsocket.UpgradeHTTP(w, r)
		if err != nil {
			log.Print(err)
			return
		}

		ws.OnClose(func(cc wsocket.CloseCode) { log.Printf("Server closed with code %d", cc) })
		ws.OnError(func(err error) { log.Print(err) })
		ws.OnMessage(func(msg wsocket.Message, opcode wsocket.Opcode) {
			w := ws.NewWriter(opcode)
			w.Write(msg)
		})
	}))
	log.Fatal(http.ListenAndServe(":9001", nil))
}
