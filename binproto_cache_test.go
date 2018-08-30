package binproto

import (
	"testing"
	"bytes"
)

func TestCacheEncodeDecodePositive(t *testing.T) {
	//GIVEN
	src := []byte{1, 1, 1, 0, 0, 1, 5, 12, 44}
	cache := NewCache()
	//WHEN
	encoded, _ := cache.Encode(src)
	encodedSave := make([]byte, len(encoded))
	copy(encodedSave, encoded)

	decoded, _ := cache.Decode(encodedSave)
	decodedSave := make([]byte, len(decoded))
	copy(decodedSave, decoded)
	//THEN
	if !bytes.Equal(decodedSave, src) {
		t.Errorf("Decoded array %v does not equal to the source %v", decodedSave, src)
	}
}

func BenchmarkCache_Encode(b *testing.B) {
	src := NewBinProtoMessage(1, 1, 1, 0, 0, 1, 5, 12, 44)
	cache := NewCache()
	cache.Encode(src)

	for i := 0; i<b.N; i++ {
		_, err := cache.Encode(src)
		if err != nil {
			b.Errorf("Failed to encode source array %v. Error: %v", src, err)
		}
	}
}

func BenchmarkCache_Decode(b *testing.B) {
	src := NewBinProtoMessage(1, 1, 1, 0, 0, 1, 5, 12, 44)
	cache := NewCache()
	cache.Encode(src)
	encoded := cache.Copy()

	for i := 0; i<b.N; i++ {
		_, err := cache.Decode(encoded)
		if err != nil {
			b.Errorf("Failed to decode source array %v. Error: %v", src, err)
		}
	}
}