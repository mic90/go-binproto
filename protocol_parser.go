package binproto

import (
	"bytes"
	"fmt"
)

// Encoder encodes given bytes slice into new data format
// The source bytes are not modified by the Encode function
type Encoder interface {
	Encode([]byte) ([]byte, error)
}

// Decoder decodes given bytes slice back into its original format
// The source bytes are not modified by the Decode function
type Decoder interface {
	Decode([]byte) ([]byte, error)
}

// EncodeDecoder can perform both encoding and decoding operations
type EncodeDecoder interface {
	Encoder
	Decoder
}

// ProtocolParser implements COBS encoder/decoder with crc checksum
type ProtocolParser struct {
	buffer    []byte
	crcBuffer []byte
	lastPos   int
}

// NewProtocolParser returns new BinProto object
func NewProtocolParser() (binProto *ProtocolParser) {
	return &ProtocolParser{[]byte{}, []byte{}, 0}
}

// Encode encodes given source slice with COBS encoding
// Before data is encoded, crc checksum is calculated over it and joined with source data
// Encoded data is stored in the internal buffer and pointer for it is returned
// If one wants to store the data for later use, Copy function must be used
func (proto *ProtocolParser) Encode(src []byte) ([]byte, error) {
	srcWithChecksumLen := len(src) + crcLen
	requiredBufferLen := cobsGetEncodedBufferSize(srcWithChecksumLen)
	if len(proto.buffer) < requiredBufferLen {
		proto.buffer = make([]byte, requiredBufferLen)
	}
	proto.crcBuffer = proto.crcBuffer[:0]
	proto.crcBuffer = append(proto.crcBuffer, src...)
	proto.crcBuffer = append(proto.crcBuffer, fletcher16(src)...)

	encodedLen, err := cobsEncode(proto.crcBuffer[:srcWithChecksumLen], proto.buffer)
	if err != nil {
		return nil, err
	}
	proto.lastPos = encodedLen
	return proto.buffer[:encodedLen], nil
}

// Decode decodes given source slice to the raw data
// It is assumed that the source slice was encoded with COBS encoding
// It is also assumed that after encoding removal, raw data consist of data + crc check sum
// If checksum read after decoding is not correct, error will be returned
func (proto *ProtocolParser) Decode(src []byte) ([]byte, error) {
	sourceLength := len(src)
	if len(proto.buffer) < sourceLength {
		proto.buffer = make([]byte, sourceLength)
	}
	decodedLength, err := cobsDecode(src, proto.buffer)
	if err != nil {
		return nil, err
	}
	if decodedLength < 2 {
		return nil, fmt.Errorf("decoded message is too short. Decoded length: %v", decodedLength)
	}
	msgWithoutCrcLen := decodedLength - crcLen
	msgWithoutCrc := proto.buffer[:msgWithoutCrcLen]
	msgCrc := proto.buffer[msgWithoutCrcLen:decodedLength]
	calculatedCrc := fletcher16(msgWithoutCrc)
	if !bytes.Equal(msgCrc, calculatedCrc) {
		return nil, fmt.Errorf("calculated crc %v doesn't match received one %v", calculatedCrc, msgCrc)
	}
	proto.lastPos = decodedLength
	return msgWithoutCrc, nil
}

// Copy will make a copy of the last encode/decode operation
// ! This function will allocate a new buffer for each call, so use it wisely
func (proto *ProtocolParser) Copy() []byte {
	newArray := make([]byte, proto.lastPos)
	copy(newArray, proto.buffer[:proto.lastPos])
	return newArray
}
