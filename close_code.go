package websocket

import "math"

func validCloseCode(cc uint16) bool {
	return cc >= 1000 && cc <= 1003 ||
		cc >= 1007 && cc <= 1011 ||
		cc >= 3000 && cc <= 5000 ||
		int(cc) == math.MaxUint16
}
