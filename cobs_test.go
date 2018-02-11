package binproto

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

func TestCobsEncodeDecodePositive(t *testing.T) {
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
		t.Errorf("Decoded byte array %v does not equal to the source %v", decodeBuffer[:decodedLen], src)
	}
}

func TestCobsEncodeWithTooSmallDest(t *testing.T) {
	//GIVEN
	src := []byte{1, 1, 1, 0, 0, 5, 0}
	encodeBuffer := make([]byte, len(src)-2)
	//WHEN
	_, err := CobsEncode(src, &encodeBuffer)
	//THEN
	if err == nil {
		t.Error("encoding with too small dest buffer succeed. This should fail")
	}
}

func TestCobsEncodeEmptySourceDest(t *testing.T) {
	//GIVEN
	src := []byte{}
	encodeBuffer := make([]byte, len(src))
	expectedEncodedLen := 0
	//WHEN
	encodedLen, err := CobsEncode(src, &encodeBuffer)
	//THEN
	if err != nil {
		t.Error("encoding with empty input failed. Error: ", err)
	}
	if encodedLen != expectedEncodedLen {
		t.Errorf("encoded message length is different than expected. Expected: %v, get: %v", expectedEncodedLen, encodedLen)
	}
}

func TestCobsDecodeWithTooSmallDest(t *testing.T) {
	//GIVEN
	encoded := []byte{4, 1, 1, 1, 1, 2, 5, 1}
	encodeBuffer := make([]byte, len(encoded)-2)
	//WHEN
	_, err := CobsDecode(encoded, &encodeBuffer)
	//THEN
	if err == nil {
		t.Error("decoding with too small dest buffer succeed. This should fail")
	}
}
