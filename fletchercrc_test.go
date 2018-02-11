package binproto

import (
	"bytes"
	"testing"
)

func TestFletcher16(t *testing.T) {
	//GIVEN
	src := []byte{1, 2, 3, 4, 5, 6}
	expectedCrc := []byte{21, 56}
	//WHEN
	crc := Fletcher16(src)
	//THEN
	if !bytes.Equal(crc, expectedCrc) {
		t.Errorf("Crc value %v is not equal to the expected one %v", crc, expectedCrc)
	}
}
