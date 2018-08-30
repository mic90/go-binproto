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
	encSize := cobsGetEncodedBufferSize(srcSize)
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
	encodedLen, _ := cobsEncode(src, encodeBuffer)
	decodedLen, _ := cobsDecode(encodeBuffer[:encodedLen], decodeBuffer)
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
	_, err := cobsEncode(src, encodeBuffer)
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
	encodedLen, err := cobsEncode(src, encodeBuffer)
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
	_, err := cobsDecode(encoded, encodeBuffer)
	//THEN
	if err == nil {
		t.Error("decoding with too small dest buffer succeed. This should fail")
	}
}

func BenchmarkCobsEncode(b *testing.B) {
	src := []byte{1, 1, 1, 0, 0, 5, 0}
	encodeBuffer := make([]byte, 100)

	b.ResetTimer()

	for i := 0; i<b.N; i++ {
		_, err := cobsEncode(src, encodeBuffer)
		if err != nil {
			b.Errorf("Failed to decode source array %v. Error: %v", src, err)
		}
	}
}

func BenchmarkCobsDecode(b *testing.B) {
	src := []byte{4, 1, 1, 1, 1, 2, 5, 1}
	encodeBuffer := make([]byte, 100)

	b.ResetTimer()

	for i := 0; i<b.N; i++ {
		_, err := cobsDecode(src, encodeBuffer)
		if err != nil {
			b.Errorf("Failed to decode source array %v. Error: %v", src, err)
		}
	}
}
