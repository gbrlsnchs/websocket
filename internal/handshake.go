package internal

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strings"
)

const (
	UpgradeHeader             = "websocket"
	ConnectionHeader          = "upgrade"
	SecWebSocketVersionHeader = "13"
	guid                      = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

var (
	ErrMissingHost                = errors.New("websocket: missing Host header")
	ErrUpgradeMismatch            = errors.New("websocket: Upgrade header mismatch")
	ErrConnectionMismatch         = errors.New("websocket: Connection header mismatch")
	ErrSecWebSocketVersionMissing = errors.New("websocket: missing Sec-WebSocket-Version header")
	ErrInvalidSecWebSocketKey     = errors.New("websocket: invalid Sec-WebSocket-Key")
)

func Handshake(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	if r.Host == "" {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return nil, ErrMissingHost
	}

	var err error
	if err = Validate(r.Header); err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return nil, err
	}

	hdr := w.Header()
	hdr.Set("Upgrade", "websocket")
	hdr.Set("Connection", "Upgrade")
	key, err := ConcatKey(r.Header.Get("Sec-WebSocket-Key"))
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return nil, err
	}
	hdr.Set("Sec-WebSocket-Accept", base64.StdEncoding.EncodeToString(key))
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
	return nil, errors.New("websocket: connection not hijackable")
}

func ConcatKey(key string) ([]byte, error) {
	if len(key) != 16 {
		return nil, ErrInvalidSecWebSocketKey
	}
	// Generate SHA-1 hash and encode it using Base64.
	sha := sha1.New()
	b := make([]byte, len(key)+len(guid))
	copy(b[copy(b, key):], guid)
	sha.Write(b)
	return sha.Sum(nil), nil
}

func Validate(hdr http.Header) error {
	switch {
	case strings.ToLower(hdr.Get("Upgrade")) != UpgradeHeader:
		return ErrUpgradeMismatch
	case strings.ToLower(hdr.Get("Connection")) != ConnectionHeader:
		return ErrConnectionMismatch
	case hdr.Get("Sec-WebSocket-Version") != SecWebSocketVersionHeader:
		return ErrSecWebSocketVersionMissing
	}
	return nil
}
