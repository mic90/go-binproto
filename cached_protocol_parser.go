package binproto

import "unsafe"

// CachedProtocolParser is a wrapper around binproto which adds simple memory cache
// Each message will be stored in cache on first encode or decode
// If given message was parsed previously, its cached results will be returned immediately
type CachedProtocolParser struct {
	protocol *ProtocolParser
	cache    map[string][]byte
}

// NewCachedProtocolParser returns new cached protocol object
func NewCachedProtocolParser() *CachedProtocolParser {
	cache := make(map[string][]byte)
	return &CachedProtocolParser{protocol: NewProtocolParser(), cache: cache}
}

// Encode encodes given source slice with COBS encoding and adds checksum
// Encoded data will be stored in memory.
// If given source have been encoded previously its encoded version will be obtained from memory
func (c *CachedProtocolParser) Encode(src []byte) ([]byte, error) {
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
func (c *CachedProtocolParser) Decode(src []byte) ([]byte, error) {
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
func (c *CachedProtocolParser) Copy() []byte {
	return c.protocol.Copy()
}
