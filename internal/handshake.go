package internal

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strings"
	"unsafe"
)

var guid = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

func Handshake(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	swsk, err := secKey(r)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return nil, err
	}
	// Generate SHA-1 hash and encode it using base64.
	sha := sha1.New()
	sha.Write(append(swsk, guid...))
	hd := w.Header()
	hd.Set("Upgrade", "websocket")
	hd.Set("Connection", "Upgrade")
	hd.Set("Sec-WebSocket-Accept", base64.StdEncoding.EncodeToString(sha.Sum(nil)))
	w.WriteHeader(http.StatusSwitchingProtocols)

	// Hijack the underlying connection.
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("websocket/internal: connection not hijackable")
	}
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

func secKey(r *http.Request) ([]byte, error) {
	switch {
	case len(r.Host) == 0:
		return nil, errors.New("websocket: missing Host header")
	case strings.ToLower(r.Header.Get("Upgrade")) != "websocket":
		return nil, errors.New("websocket: Upgrade header mismatch")
	case strings.ToLower(r.Header.Get("Connection")) != "upgrade":
		return nil, errors.New("websocket: Connection header mismatch")
	case len(r.Header.Get("Sec-WebSocket-Version")) == 0:
		return nil, errors.New("websocket: missing Sec-WebSocket-Version header")
	}
	swk := r.Header.Get("Sec-WebSocket-Key")
	if len(swk) == 0 {
		return nil, errors.New("websocket: missing Sec-WebSocket-Key header")
	}
	return *(*[]byte)(unsafe.Pointer(&swk)), nil // same as strings.(*Builder).String method
}
