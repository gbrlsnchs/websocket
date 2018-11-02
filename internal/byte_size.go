package internal

import "math"

func ByteSize(b []byte, masked bool) (size int) {
	size++ // FIN

	length := len(b)
	switch {
	case length <= 125:
		size++ // indicator is the current length
	case length <= math.MaxUint16:
		size += 3 // indicator + 2 bytes for length value
	default:
		size += 9 // indicator + 8 bytes for length value
	}
	if masked {
		size += 4 // 4 bytes for masking-key
	}
	return size + length
}
