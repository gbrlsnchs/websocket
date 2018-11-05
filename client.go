package websocket

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gbrlsnchs/uuid"
	"github.com/gbrlsnchs/websocket/internal"
)

// Open creates a WebSocket instance in client mode.
//
// The address must use either "ws" or "wss" protocols.
// If the port is omitted, it assumes port 80 for "ws" and port 443 for "wss".
func Open(address string, timeout time.Duration) (*WebSocket, error) {
	return open(address, timeout, nil)
}

// OpenTLS creates a secure WebSocket instance in client mode.
//
// If the URI scheme is "ws", the TLS configuration is ignored.
func OpenTLS(address string, timeout time.Duration, config *tls.Config) (*WebSocket, error) {
	return open(address, timeout, config)
}

func open(address string, timeout time.Duration, config *tls.Config) (*WebSocket, error) {
	uri, err := url.Parse(address)
	if err != nil {
		return nil, err
	}
	noPort := uri.Port() == ""
	isWSS := false
	switch uri.Scheme {
	case "ws":
		uri.Scheme = "http"
		if noPort {
			uri.Host += ":80"
		}
	case "wss":
		uri.Scheme = "https"
		if noPort {
			uri.Host += ":443"
		}
		isWSS = true
	default:
		return nil, fmt.Errorf("websocket: unsupported protocol %s", uri.Scheme)
	}

	r, err := http.NewRequest(http.MethodGet, uri.String(), nil)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Upgrade", internal.UpgradeHeader)
	r.Header.Set("Connection", internal.ConnectionHeader)
	r.Header.Set("Sec-WebSocket-Version", internal.SecWebSocketVersionHeader)
	guid, err := uuid.GenerateV4(nil)
	if err != nil {
		return nil, err
	}
	encKey := base64.StdEncoding.EncodeToString(guid[:])
	r.Header.Set("Sec-WebSocket-Key", encKey)

	d := &net.Dialer{Timeout: timeout}
	var conn net.Conn

	if isWSS {
		conn, err = tls.DialWithDialer(d, "tcp", uri.Host, config)
	} else {
		conn, err = d.Dial("tcp", uri.Host)
	}
	if err != nil {
		return nil, err
	}
	if err = sendReq(r, conn, encKey); err != nil {
		conn.Close()
		return nil, err
	}
	return newWS(conn, true), nil
}

func sendReq(r *http.Request, conn net.Conn, encKey string) error {
	b, err := httputil.DumpRequestOut(r, true)
	if err != nil {
		return err
	}
	if _, err = conn.Write(b); err != nil {
		return err
	}
	rd := bufio.NewReader(conn)
	rr, err := http.ReadResponse(rd, r)
	if err != nil {
		return err
	}
	if rr.StatusCode != http.StatusSwitchingProtocols {
		return errors.New("websocket: client did not receive 101 status response")
	}
	return validateServerHeaders(rr.Header, encKey)
}

func validateServerHeaders(hdr http.Header, encKey string) error {
	switch {
	case strings.ToLower(hdr.Get("Upgrade")) != internal.UpgradeHeader:
		return internal.ErrUpgradeMismatch
	case strings.ToLower(hdr.Get("Connection")) != internal.ConnectionHeader:
		return internal.ErrConnectionMismatch
	}
	key, err := internal.ConcatKey(encKey)
	if err != nil {
		return err
	}
	srvKey := hdr.Get("Sec-WebSocket-Accept")
	dec, err := base64.StdEncoding.DecodeString(srvKey)
	if err != nil {
		return err
	}
	if string(key) != string(dec) {
		return internal.ErrSecWebSocketKeyMismatch
	}
	return nil
}
