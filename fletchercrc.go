package go_binproto

var buff = make([]byte, 2)

func Fletcher16(src []byte) []byte {
	sumA, sumB := uint16(0), uint16(0)

	for _, val := range src {
		sumA = (sumA + uint16(val)) % 255
		sumB = (sumB + sumA) % 255
	}

	buff[0] = byte(sumA)
	buff[1] = byte(sumB)
	return buff
}
