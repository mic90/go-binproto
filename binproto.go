package go_binproto

import (
	"bytes"
	"errors"
	"sync"
)

type BinProto struct {
	buffer []byte
	mutex  sync.Mutex
}

func NewBinProto(maxMsgLen int) (binProto *BinProto) {
	requiredLen := maxMsgLen + 2
	requiredEncodedLen := CobsGetEncodedBufferSize(requiredLen)
	buffer := make([]byte, requiredEncodedLen)

	return &BinProto{buffer, sync.Mutex{}}
}

func (proto *BinProto) clear() {
	for index := range proto.buffer {
		proto.buffer[index] = 0
	}
}

func (proto *BinProto) Encode(src []byte) ([]byte, error) {
	proto.mutex.Lock()
	defer proto.mutex.Unlock()

	proto.clear()

	src = append(src, Fletcher16(src)...)

	encodedLen, err := CobsEncode(src, &proto.buffer)
	if err != nil {
		return nil, err
	}
	return proto.buffer[:encodedLen], nil
}

func (proto *BinProto) Decode(src []byte) ([]byte, error) {
	proto.mutex.Lock()
	defer proto.mutex.Unlock()

	proto.clear()

	decodedLen, err := CobsDecode(src, &proto.buffer)
	if err != nil {
		return nil, err
	}
	msgWithoutCrc := proto.buffer[:decodedLen-2]
	msgCrc := proto.buffer[decodedLen-2 : decodedLen]
	calculatedCrc := Fletcher16(msgWithoutCrc)
	if !bytes.Equal(msgCrc, calculatedCrc) {
		return nil, errors.New("calculated crc doesn't match received one")
	}
	newBuff := make([]byte, len(msgWithoutCrc))
	copy(newBuff, msgWithoutCrc)
	return newBuff, nil
}

func NewBinProtoMessage(data ...byte) []byte {
	//reserve space for crc checksum
	requiredLen := len(data) + 2
	msgSlice := make([]byte, len(data), requiredLen)
	copy(msgSlice, data)
	return msgSlice
}
