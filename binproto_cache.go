package binproto

import "unsafe"

// Cache is a wrapper around binproto which adds simple memory cache
// Each message will be stored in cache on first encode or decode
// If given message was parsed previously, its cached results will be returned immediately
type Cache struct {
	protocol *BinProto
	cache    map[string][]byte
}

func NewCache() *Cache {
	cache := make(map[string][]byte)
	return &Cache{protocol: NewBinProto(), cache: cache}
}

// Encode encodes given source slice with COBS encoding and adds checksum
// Encoded data will be stored in memory.
// If given source have been encoded previously its encoded version will be obtained from memory
func (c *Cache) Encode(src []byte) ([]byte, error) {
	srcHash := *(*string)(unsafe.Pointer(&src))
	if encoded, ok := c.cache[srcHash]; ok {
		return encoded, nil
	}
	data, err := c.protocol.Encode(src)
	if err != nil {
		return nil, err
	}
	// copy src to make the cache immune to future src changes
	c.cache[string(src)] = data
	return data, nil
}

// Decode decodes given source slice, which was previously encoded with COBS encoding
// Decoded data will be stored in memory.
// If given source have been decoded previously its decoded version will be obtained from memory
func (c *Cache) Decode(src []byte) ([]byte, error) {
	srcHash := *(*string)(unsafe.Pointer(&src))
	if decoded, ok := c.cache[srcHash]; ok {
		return decoded, nil
	}
	data, err := c.protocol.Decode(src)
	if err != nil {
		return nil, err
	}
	// copy src to make the cache immune to future src changes
	c.cache[string(src)] = data
	return data, nil
}

// Copy will make a copy of the last encode/decode operation
// ! This function will allocate a new buffer for each call, so use it wisely
func (c *Cache) Copy() []byte {
	return c.protocol.Copy()
}
