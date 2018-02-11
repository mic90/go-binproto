package binproto

import (
	"bytes"
	"fmt"
	"math/rand"
	"runtime"
	"testing"
)

func TestEncodeDecodePositive(t *testing.T) {
	//GIVEN
	src := []byte{1, 1, 1, 0, 0, 1, 5, 12, 44}
	proto := NewBinProto(len(src))
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
	emptySrc := []byte{}
	expectedEncoded := []byte{1, 1, 1}
	proto := NewBinProto(20)
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
	emptyEncoded := []byte{}
	proto := NewBinProto(20)
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
	proto := NewBinProto(20)

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
	proto := NewBinProto(20)
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
	proto := NewBinProto(len(src))
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
	proto := NewBinProto(len(src))
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

func TestProtoMemoryUsage(t *testing.T) {
	//GIVEN
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// 2000 is roughly the memory usage increase beacuse of runtime.ReadMemStats
	maxMemoryUsage := uint64(2000)
	iterationsCount := 10000
	src := NewBinProtoMessage(1, 1, 1, 0, 0, 1, 5, 12, 44)
	proto := NewBinProto(len(src))
	//WHEN
	for i := 0; i < iterationsCount; i++ {
		_, err := proto.Encode(src)
		if err != nil {
			t.Errorf("Failed to encode source array %v. Error: %v", src, err)
		}
	}
	//THEN
	allocBefore := m.Alloc
	runtime.ReadMemStats(&m)
	allocDiff := m.Alloc - allocBefore
	if allocDiff > maxMemoryUsage {
		t.Errorf("Memory usage after %d iterations is higher than expected. Got: %d, expected below: %d", iterationsCount, allocDiff, maxMemoryUsage)
	}
}
