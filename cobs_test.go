package go_binproto

import (
	"bytes"
	"testing"
)

func TestCobsGetEncodedBufferSize(t *testing.T) {
	//GIVEN
	srcSize := 900
	expectedEncSize := 904
	//WHEN
	encSize := CobsGetEncodedBufferSize(srcSize)
	//THEN
	if encSize != expectedEncSize {
		t.Errorf("Encoded buffer size does not equal to the expected one. Got: %v, expected: %v", encSize, expectedEncSize)
	}
}

func TestCobsPositive(t *testing.T) {
	//GIVEN
	src := []byte{1, 1, 1, 0, 0, 5, 0}
	encodeBuffer := make([]byte, 100)
	decodeBuffer := make([]byte, 100)
	expectedEncoded := []byte{4, 1, 1, 1, 1, 2, 5, 1}
	//WHEN
	encodedLen, _ := CobsEncode(src, &encodeBuffer)
	decodedLen, _ := CobsDecode(encodeBuffer[:encodedLen], &decodeBuffer)
	////THEN
	if !bytes.Equal(expectedEncoded, encodeBuffer[:encodedLen]) {
		t.Errorf("Encoded byte array %v does not equal to the expected %v", encodeBuffer[:encodedLen], expectedEncoded)
	}
	if !bytes.Equal(decodeBuffer[:decodedLen], src) {
		t.Errorf("Decoded byte array %v does nto equal to the source %v", decodeBuffer[:decodedLen], src)
	}
}
