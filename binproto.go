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

// BinProto implements COBS encoder/decoder with crc checksum
type BinProto struct {
	// buffer will contain result of encode/decode
	buffer bytes.Buffer
	// crcBuffer is temporary buffer which holds concatenated src slice and checksum
	crcBuffer bytes.Buffer
	arr       []byte
	crcArr    []byte
	lastPos   int
}

// NewBinProto returns new BinProto object
func NewBinProto() (binProto *BinProto) {
	return &BinProto{bytes.Buffer{}, bytes.Buffer{}, []byte{}, []byte{}, 0}
}

// Encode encodes given source slice with COBS encoding
// Before data is encoded, crc checksum is calculated over it and joined with source data
// Encoded data is stored in the internal buffer and pointer for it is returned
// If one wants to store the data for later use, Copy function must be used
func (proto *BinProto) Encode(src []byte) ([]byte, error) {
	srcWithChecksumLen := len(src) + crcLen
	requiredBufferLen := cobsGetEncodedBufferSize(srcWithChecksumLen)
	if len(proto.arr) < requiredBufferLen {
		proto.arr = make([]byte, requiredBufferLen)
	}
	proto.crcArr = proto.crcArr[:0]
	proto.crcArr = append(proto.crcArr, src...)
	proto.crcArr = append(proto.crcArr, fletcher16(src)...)

	encodedLen, err := cobsEncode(proto.crcArr[:srcWithChecksumLen], proto.arr)
	if err != nil {
		return nil, err
	}
	proto.lastPos = encodedLen
	return proto.arr[:encodedLen], nil
}

// Decode decodes given source slice to the raw data
// It is assumed that the source slice was encoded with COBS encoding
// It is also assumed that after encoding removal, raw data consist of data + crc check sum
// If checksum read after decoding is not correct, error will be returned
func (proto *BinProto) Decode(src []byte) ([]byte, error) {
	sourceLength := len(src)
	if len(proto.arr) < sourceLength {
		proto.arr = make([]byte, sourceLength)
	}
	decodedLength, err := cobsDecode(src, proto.arr)
	if err != nil {
		return nil, err
	}
	if decodedLength < 2 {
		return nil, fmt.Errorf("decoded message is too short. Decoded length: %v", decodedLength)
	}
	msgWithoutCrcLen := decodedLength - crcLen
	msgWithoutCrc := proto.arr[:msgWithoutCrcLen]
	msgCrc := proto.arr[msgWithoutCrcLen:decodedLength]
	calculatedCrc := fletcher16(msgWithoutCrc)
	if !bytes.Equal(msgCrc, calculatedCrc) {
		return nil, fmt.Errorf("calculated crc %v doesn't match received one %v", calculatedCrc, msgCrc)
	}
	proto.lastPos = decodedLength
	return msgWithoutCrc, nil
}

// Copy will make a copy of the last encode/decode operation
// ! This function will allocate a new buffer for each call, so use it wisely
func (proto BinProto) Copy() []byte {
	newArray := make([]byte, proto.lastPos)
	copy(newArray, proto.arr[:proto.lastPos])
	return newArray
}

