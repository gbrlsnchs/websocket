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
		ww := ws.NewWriter() // one buffered writer is enough

		for ws.IsOpen() {
			b, opcode, err := ws.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}
			ww.SetOpcode(opcode)
			ww.Write(b)
		}
		fmt.Println(ws.CloseCode())
	}))
	log.Fatal(http.ListenAndServe(":9001", nil))
}
