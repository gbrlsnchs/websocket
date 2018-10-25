package internal

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strings"
)

const guid = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func Handshake(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	key, err := secKey(r)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return nil, err
	}
	// Generate SHA-1 hash and encode it using Base64.
	sha := sha1.New()
	b := make([]byte, len(key)+len(guid))
	copy(b, key)
	copy(b, guid)
	sha.Write(b)

	hd := w.Header()
	hd.Set("Upgrade", "websocket")
	hd.Set("Connection", "Upgrade")
	hd.Set("Sec-WebSocket-Accept", base64.StdEncoding.EncodeToString(sha.Sum(nil)))
	w.WriteHeader(http.StatusSwitchingProtocols)

	// Hijack the underlying connection.
	if hj, ok := w.(http.Hijacker); ok {
		conn, bufrw, err := hj.Hijack()
		if err != nil {
			return nil, err
		}
		if err = bufrw.Flush(); err != nil {
			defer conn.Close()
			return nil, err
		}
		return conn, nil
	}
	return nil, errors.New("websocket/internal: connection not hijackable")
}

func secKey(r *http.Request) (string, error) {
	switch {
	case r.Host == "":
		return "", errors.New("websocket: missing Host header")
	case strings.ToLower(r.Header.Get("Upgrade")) != "websocket":
		return "", errors.New("websocket: Upgrade header mismatch")
	case strings.ToLower(r.Header.Get("Connection")) != "upgrade":
		return "", errors.New("websocket: Connection header mismatch")
	case r.Header.Get("Sec-WebSocket-Version") == "":
		return "", errors.New("websocket: missing Sec-WebSocket-Version header")
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return "", errors.New("websocket: missing Sec-WebSocket-Key header")
	}
	return key, nil
}
