package websocket_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/gbrlsnchs/websocket"
)

func TestWebSocket(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := UpgradeHTTP(w, r)
		if err != nil {
			return
		}
		ws.Write([]byte("test"))
		ws.Close()
	}))
	defer srv.Close()

	uri := "ws://" + strings.TrimPrefix(srv.URL, "http://")
	ws, err := Open(uri, 15*time.Second)
	if want, got := (error)(nil), err; want != got {
		t.Fatalf("want %v, got %v", want, got)
	}

	// First frame.
	if want, got := true, ws.Next(); want != got {
		t.Fatalf("want %t, got %t", want, got)
	}
	payload, opcode := ws.Message()
	if want, got := OpcodeText, opcode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}
	if want, got := "test", payload; want != string(got) {
		t.Errorf("want %s, got %s", want, got)
	}

	// Closing frame.
	if want, got := false, ws.Next(); want != got {
		t.Fatalf("want %t, got %t", want, got)
	}
	if want, got := uint16(1000), ws.CloseCode(); want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}
