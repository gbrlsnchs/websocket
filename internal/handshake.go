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
	ErrSecWebSocketKeyMismatch    = errors.New("websocket: key mismatch")
)

func Handshake(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	if r.Host == "" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, ErrMissingHost
	}

	var err error
	if err = validateClientHeaders(r.Header); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	hdr := w.Header()
	hdr.Set("Upgrade", "websocket")
	hdr.Set("Connection", "Upgrade")
	key, err := ConcatKey(r.Header.Get("Sec-WebSocket-Key"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	hdr.Set("Sec-WebSocket-Accept", base64.StdEncoding.EncodeToString(key))

	// Hijack the underlying connection.
	if hj, ok := w.(http.Hijacker); ok {
		w.WriteHeader(http.StatusSwitchingProtocols)
		conn, bufrw, err := hj.Hijack()
		if err != nil {
			return nil, err
		}
		if err = bufrw.Flush(); err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil
	}
	w.WriteHeader(http.StatusBadRequest)
	return nil, errors.New("websocket: connection not hijackable")
}

func ConcatKey(key string) ([]byte, error) {
	// Generate SHA-1 hash and encode it using Base64.
	sha := sha1.New()
	b := make([]byte, len(key)+len(guid))
	copy(b[copy(b, key):], guid)
	sha.Write(b)
	return sha.Sum(nil), nil
}

func validateClientHeaders(hdr http.Header) error {
	switch {
	case strings.ToLower(hdr.Get("Upgrade")) != UpgradeHeader:
		return ErrUpgradeMismatch
	case strings.ToLower(hdr.Get("Connection")) != ConnectionHeader:
		return ErrConnectionMismatch
	case hdr.Get("Sec-WebSocket-Version") != SecWebSocketVersionHeader:
		return ErrSecWebSocketVersionMissing
	}
	key := hdr.Get("Sec-WebSocket-Key")
	dec, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return err
	}
	if len(dec) != 16 {
		return ErrInvalidSecWebSocketKey
	}
	return nil
}
