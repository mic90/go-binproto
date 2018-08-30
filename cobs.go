package binproto

import (
	"fmt"
)

func cobsEncode(src []byte, dest []byte) (int, error) {
	srcLen := len(src)
	if srcLen == 0 {
		return 0, nil
	}

	requiredLen := cobsGetEncodedBufferSize(srcLen)
	if len(dest) < requiredLen {
		return 0, fmt.Errorf("destination array length is too small. Required: %v, get: %v", requiredLen, len(dest))
	}

	codePtr := 0
	code := byte(0x01)
	pos := 1

	for _, srcValue := range src {
		if srcValue == 0 {
			dest[codePtr] = code
			codePtr = pos
			dest[pos] = 0
			pos++
			code = byte(0x01)
			continue
		}

		dest[pos] = srcValue
		pos++
		code++
		if code == 0xFF {
			dest[codePtr] = code
			codePtr = pos
			dest[pos] = 0
			pos++
			code = byte(0x01)
		}
	}
	dest[codePtr] = code
	return pos, nil
}

func cobsDecode(enc []byte, dest []byte) (int, error) {
	encLen := len(enc)
	destLen := len(dest)
	ptr := 0
	pos := 0

	if encLen == 0 {
		return 0, nil
	}

	for ptr < encLen {
		code := enc[ptr]

		if ptr+int(code) > encLen {
			return 0, fmt.Errorf("encoded message is too short. Required: %v, get: %v", ptr+int(code), encLen)
		}
		ptr++

		if pos+int(code) > destLen {
			return 0, fmt.Errorf("destination array length is too short. Required: %v, get: %v", pos+int(code), destLen)
		}

		for i := 1; i < int(code); i++ {
			dest[pos] = enc[ptr]
			pos++
			ptr++
		}
		if code < 0xFF {
			dest[pos] = 0
			pos++
		}
	}

	return pos - 1, nil // trim phantom zero
}

func cobsGetEncodedBufferSize(rawSize int) int {
	return rawSize + rawSize/254 + 1
}
