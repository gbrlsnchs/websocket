package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gbrlsnchs/websocket"
)

type test struct {
	Msg string `json:"message,omitempty"`
}

func main() {
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "github.com/gbrlsnchs/websocket")
		ws, err := websocket.UpgradeHTTP(w, r)
		if err != nil {
			fmt.Println(err)
			return
		}
		dec := json.NewDecoder(ws)
		enc := json.NewEncoder(ws)
		var t test

		for ws.Next() {
			if err = dec.Decode(&t); err != nil {
				fmt.Println(err)
				ws.Write([]byte(err.Error()))
				ws.Close()
				return
			}
			switch t.Msg {
			case "hello":
				t.Msg = "world"
			case "ping":
				t.Msg = "pong"
			default:
				t.Msg = "dunno"
			}
			if err = enc.Encode(t); err != nil {
				fmt.Println(err)
				ws.Write([]byte(err.Error()))
				ws.Close()
				return
			}
		}
		if err = ws.Err(); err != nil {
			fmt.Println(err)
		}
		fmt.Println(ws.CloseCode())
	}))
	log.Fatal(http.ListenAndServe(":9001", nil))
}
