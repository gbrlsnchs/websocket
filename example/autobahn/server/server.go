package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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
		rr := ws.NewReader()
		ww := ws.NewWriter()

		for ws.Next(rr, ww) {
			if _, err = io.Copy(ww, rr); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		if err := ws.Err(); err != nil {
			fmt.Println(err)
		}
		fmt.Println(ws.CloseCode())
	}))
	log.Fatal(http.ListenAndServe(":9001", nil))
}
