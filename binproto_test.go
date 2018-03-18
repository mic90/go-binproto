package binproto

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

func TestEncodeDecodePositive(t *testing.T) {
	//GIVEN
	src := []byte{1, 1, 1, 0, 0, 1, 5, 12, 44}
	proto := NewBinProto()
	//WHEN
	encoded, _ := proto.Encode(src)
	encodedSave := make([]byte, len(encoded))
	copy(encodedSave, encoded)

	decoded, _ := proto.Decode(encodedSave)
	decodedSave := make([]byte, len(decoded))
	copy(decodedSave, decoded)
	//THEN
	if !bytes.Equal(decodedSave, src) {
		t.Errorf("Decoded array %v does not equal to the source %v", decodedSave, src)
	}
}

func TestEncodeEmptyWithInput(t *testing.T) {
	//GIVEN
	var emptySrc []byte
	expectedEncoded := []byte{1, 1, 1}
	proto := NewBinProto()
	//WHEN
	encoded, err := proto.Encode(emptySrc)
	//THEN
	if err != nil {
		t.Error("decoding empty data failed: ", err)
	}
	if !bytes.Equal(encoded, expectedEncoded) {
		t.Errorf("encoded data is different than expected. Expected: %v, get: %v", expectedEncoded, encoded)
	}
}

func TestDecodeWithEmptyInput(t *testing.T) {
	//GIVEN
	var emptyEncoded []byte
	proto := NewBinProto()
	//WHEN
	_, err := proto.Decode(emptyEncoded)
	//THEN
	if err == nil {
		t.Error("empty message decoding succeed. This should fail")
	}
	if err.Error() != fmt.Errorf("decoded message is too short. Decoded length: %v", len(emptyEncoded)).Error() {
		t.Error("wrong error message received:", err)
	}
}

func TestDecodeWithTooShortInput(t *testing.T) {
	//GIVEN
	tooShortEncoded := []byte{1, 1}
	proto := NewBinProto()

	//WHEN/THEN
	_, err := proto.Decode(tooShortEncoded)
	if err == nil {
		t.Error("too short message decoding succeed. This should faild")
	}
	if err.Error() != fmt.Errorf("decoded message is too short. Decoded length: %v", 1).Error() {
		t.Error("wrong error message received:", err)
	}
}

func TestDecodeWithMalformedLengthByte(t *testing.T) {
	//GIVEN
	encoded := []byte{3, 1}
	proto := NewBinProto()
	//WHEN
	_, err := proto.Decode(encoded)
	//THEN
	if err == nil {
		t.Error("malformed message decoding succeed. This should fail")
	}
	if err.Error() != fmt.Errorf("encoded message is too short. Required: %v, get: %v", 3, 2).Error() {
		t.Error("wrong error message received:", err)
	}
}

func TestProtoPositiveWithBigData(t *testing.T) {
	//GIVEN
	dataSize := 10000
	src := make([]byte, dataSize)
	rand.Read(src)
	proto := NewBinProto()
	//WHEN
	encoded, _ := proto.Encode(src)
	encodedSave := make([]byte, len(encoded))
	copy(encodedSave, encoded)

	decoded, _ := proto.Decode(encodedSave)
	decodedSave := make([]byte, len(decoded))
	copy(decodedSave, decoded)
	//THEN
	if !bytes.Equal(decodedSave, src) {
		t.Errorf("Decoded array %v does not equal to the source %v", decodedSave, src)
	}
}

func TestProtoCrcMismatch(t *testing.T) {
	//GIVEN
	src := []byte{1, 1, 1, 0, 0, 1, 5, 12, 44}
	proto := NewBinProto()
	//WHEN
	encoded, _ := proto.Encode(src)
	encodedSave := make([]byte, len(encoded))
	copy(encodedSave, encoded)
	encodedSave[2] = -encodedSave[2]

	_, err := proto.Decode(encodedSave)
	if err == nil {
		t.Errorf("Failed to decode source array %v. Error: %v", encodedSave, err)
	}
}

func BenchmarkBinProto_Encode(b *testing.B) {
	src := NewBinProtoMessage(1, 1, 1, 0, 0, 1, 5, 12, 44)
	proto := NewBinProto()

	b.ResetTimer()

	for i := 0; i<b.N; i++ {
		_, err := proto.Encode(src)
		if err != nil {
			b.Errorf("Failed to encode source array %v. Error: %v", src, err)
		}
	}
}

func BenchmarkBinProto_Decode(b *testing.B) {
	src := NewBinProtoMessage(1, 1, 1, 0, 0, 1, 5, 12, 44)
	proto := NewBinProto()
	proto.Encode(src)
	encoded := proto.Copy()

	b.ResetTimer()

	for i := 0; i<b.N; i++ {
		_, err := proto.Decode(encoded)
		if err != nil {
			b.Errorf("Failed to decode source array %v. Error: %v", src, err)
		}
	}
}
