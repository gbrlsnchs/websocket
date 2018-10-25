package websocket

import "math"

type CloseCode uint16

func (cc CloseCode) isValid() bool {
	return cc >= 1000 && cc <= 1003 ||
		cc >= 1007 && cc <= 1011 ||
		cc >= 3000 && cc <= 5000 ||
		int(cc) == math.MaxUint16
}
