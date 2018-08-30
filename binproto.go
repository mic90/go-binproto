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
	buffer  bytes.Buffer
	// crcBuffer is temporary buffer which holds concatenated src slice and checksum
	crcBuffer bytes.Buffer
	lastPos int
}

// NewBinProto returns new BinProto object
func NewBinProto() (binProto *BinProto) {
	return &BinProto{bytes.Buffer{}, bytes.Buffer{}, 0}
}

// Encode encodes given source slice with COBS encoding
// Before data is encoded, crc checksum is calculated over it and joined with source data
// Encoded data is stored in the internal buffer and pointer for it is returned
// If one wants to store the data for later use, Copy function must be used
func (proto *BinProto) Encode(src []byte) ([]byte, error) {
	srcWithChecksumLen := len(src) + crcLen
	requiredBufferLen := cobsGetEncodedBufferSize(srcWithChecksumLen)
	proto.clear()
	proto.checkSizeAndGrow(requiredBufferLen)

	proto.crcBuffer.Write(src)
	proto.crcBuffer.Write(fletcher16(src))

	encodedLen, err := cobsEncode(proto.crcBuffer.Bytes(), proto.buffer.Bytes())
	if err != nil {
		return nil, err
	}
	proto.lastPos = encodedLen
	return proto.buffer.Bytes()[:encodedLen], nil
}

// Decode decodes given source slice to the raw data
// It is assumed that the source slice was encoded with COBS encoding
// It is also assumed that after encoding removal, raw data consist of data + crc check sum
// If checksum read after decoding is not correct, error will be returned
func (proto *BinProto) Decode(src []byte) ([]byte, error) {
	proto.clear()
	proto.checkSizeAndGrow(len(src))

	decodedLen, err := cobsDecode(src, proto.buffer.Bytes())
	if err != nil {
		return nil, err
	}
	if decodedLen < 2 {
		return nil, fmt.Errorf("decoded message is too short. Decoded length: %v", decodedLen)
	}
	msgWithoutCrc := proto.buffer.Next(decodedLen - crcLen)
	msgCrc := proto.buffer.Next(crcLen)
	calculatedCrc := fletcher16(msgWithoutCrc)
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
// Allocated slice cap will be of len(source) + crcLen, and len of len(source)
func NewBinProtoMessage(data ...byte) []byte {
	//reserve space for crc checksum
	requiredLen := len(data) + crcLen
	msgSlice := make([]byte, len(data), requiredLen)
	copy(msgSlice, data)
	return msgSlice
}

func (proto *BinProto) clear() {
	proto.crcBuffer.Reset()
	proto.buffer.Reset()
	proto.lastPos = 0
}

func (proto *BinProto) checkSizeAndGrow(requiredLen int) {
	if proto.buffer.Len() < requiredLen {
		for i := 0; i<requiredLen; i++ {
			proto.buffer.WriteByte(0)
		}
	}
}
