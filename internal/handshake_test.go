package internal_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/gbrlsnchs/websocket/internal"
)

func TestHandshake(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := Handshake(w, r)
		if err != nil {
			return
		}
		conn.Close()
	}))
	defer srv.Close()

	testCases := []struct {
		upgrade      string
		connection   string
		secWSVersion string
		secWSKey     string
		status       int
	}{
		{status: http.StatusBadRequest},
		{
			upgrade:      "websocket",
			connection:   "upgrade",
			secWSVersion: "13",
			secWSKey:     "dGhlIHNhbXBsZSBub25jZQ==",
			status:       http.StatusSwitchingProtocols,
		},
		{
			upgrade:      "foo",
			connection:   "upgrade",
			secWSVersion: "13",
			secWSKey:     "dGhlIHNhbXBsZSBub25jZQ==",
			status:       http.StatusBadRequest,
		},
		{
			upgrade:      "websocket",
			connection:   "bar",
			secWSVersion: "13",
			secWSKey:     "dGhlIHNhbXBsZSBub25jZQ==",
			status:       http.StatusBadRequest,
		},
		{
			upgrade:      "websocket",
			connection:   "upgrade",
			secWSVersion: "baz",
			secWSKey:     "dGhlIHNhbXBsZSBub25jZQ==",
			status:       http.StatusBadRequest,
		},
		{
			upgrade:      "websocket",
			connection:   "upgrade",
			secWSVersion: "13",
			secWSKey:     "qux",
			status:       http.StatusBadRequest,
		},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			r, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			if want, got := (error)(nil), err; want != got {
				t.Fatalf("want %v, got %v", want, got)
			}
			r.Header.Set("Upgrade", tc.upgrade)
			r.Header.Set("Connection", tc.connection)
			r.Header.Set("Sec-WebSocket-Version", tc.secWSVersion)
			r.Header.Set("Sec-WebSocket-Key", tc.secWSKey)

			var c http.Client
			rr, err := c.Do(r)
			if want, got := (error)(nil), err; want != got {
				t.Fatalf("want %v, got %v", want, got)
			}
			if want, got := tc.status, rr.StatusCode; want != got {
				t.Errorf("want %d, got %d", want, got)
			}
		})
	}
}
