package binproto

import (
	"bytes"
	"fmt"
)

type BinProto struct {
	buffer  bytes.Buffer
	lastPos int
}

func NewBinProto() (binProto *BinProto) {
	var buffer bytes.Buffer
	return &BinProto{buffer, 0}
}

func (proto *BinProto) Encode(src []byte) ([]byte, error) {
	requiredBufferLen := CobsGetEncodedBufferSize(len(src) + CrcLen)
	proto.clear()
	proto.checkSizeAndGrow(requiredBufferLen)

	src = append(src, Fletcher16(src)...)

	encodedLen, err := CobsEncode(src, proto.buffer.Bytes())
	if err != nil {
		return nil, err
	}
	proto.lastPos = encodedLen
	return proto.buffer.Bytes()[:encodedLen], nil
}

func (proto *BinProto) Decode(src []byte) ([]byte, error) {
	proto.clear()
	proto.checkSizeAndGrow(len(src))

	decodedLen, err := CobsDecode(src, proto.buffer.Bytes())
	if err != nil {
		return nil, err
	}
	if decodedLen < 2 {
		return nil, fmt.Errorf("decoded message is too short. Decoded length: %v", decodedLen)
	}
	msgWithoutCrc := proto.buffer.Next(decodedLen - CrcLen)
	msgCrc := proto.buffer.Next(CrcLen)
	calculatedCrc := Fletcher16(msgWithoutCrc)
	if !bytes.Equal(msgCrc, calculatedCrc) {
		return nil, fmt.Errorf("calculated crc %v doesn't match received one %v", calculatedCrc, msgCrc)
	}
	proto.lastPos = decodedLen
	return msgWithoutCrc, nil
}

// Copy will make a copy of the last encode/decode operation
// ! This function will allocate a new buffer for each call, so use it wisely
func (proto BinProto) Copy() []byte {
	newArray := make([]byte, proto.lastPos)
	copy(newArray, proto.buffer.Bytes()[:proto.lastPos])
	return newArray
}

// NewBinProtoMessage will allocate new slice for given message, to contain crc checksum
// Allocated slice cap will be of len(source) + CrcLen, and len of len(source)
func NewBinProtoMessage(data ...byte) []byte {
	//reserve space for crc checksum
	requiredLen := len(data) + CrcLen
	msgSlice := make([]byte, len(data), requiredLen)
	copy(msgSlice, data)
	return msgSlice
}

func (proto *BinProto) clear() {
	proto.buffer.Reset()
}

func (proto *BinProto) checkSizeAndGrow(requiredLen int) {
	if proto.buffer.Len() < requiredLen {
		for i := 0; i<requiredLen; i++ {
			proto.buffer.WriteByte(0)
		}
	}
}
