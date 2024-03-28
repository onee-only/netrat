package util

func BoolToUint8(b bool) (i uint8) {
	if b {
		i = 1
	} else {
		i = 0
	}
	return
}

func Uint8ToBool(i uint8) bool {
	return i != 0
}
