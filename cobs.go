package go_binproto

import (
	"errors"
)

func CobsEncode(src []byte, dest *[]byte) (int, error) {
	srcLen := len(src)
	if srcLen == 0 {
		return 0, nil
	}

	requiredLen := CobsGetEncodedBufferSize(srcLen)
	if len(*dest) < requiredLen {
		return 0, errors.New("destination array capacity is too small")
	}

	codePtr := 0
	code := byte(0x01)
	pos := 1

	for _, srcValue := range src {
		if srcValue == 0 {
			(*dest)[codePtr] = code
			codePtr = pos
			(*dest)[pos] = 0
			pos++
			code = byte(0x01)
			continue
		}

		(*dest)[pos] = srcValue
		pos++
		code++
		if code == 0xFF {
			(*dest)[codePtr] = code
			codePtr = pos
			(*dest)[pos] = 0
			pos++
			code = byte(0x01)
		}
	}
	(*dest)[codePtr] = code
	return pos, nil
}

func CobsDecode(enc []byte, dest *[]byte) (int, error) {
	encLen := len(enc)
	ptr := 0
	pos := 0

	for ptr < encLen {
		code := enc[ptr]

		if ptr+int(code) > encLen {
			return 0, errors.New("unable to decode, message is too short")
		}

		ptr++

		for i := 1; i < int(code); i++ {
			(*dest)[pos] = enc[ptr]
			pos++
			ptr++
		}
		if code < 0xFF {
			(*dest)[pos] = 0
			pos++
		}
	}

	if len(*dest) == 0 {
		return 0, nil
	}

	return pos - 1, nil // trim phantom zero
}

func CobsGetEncodedBufferSize(rawSize int) int {
	return rawSize + rawSize/254 + 1
}
