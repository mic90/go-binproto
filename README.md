# go-binproto 

[![Build Status](https://travis-ci.org/mic90/go-binproto.svg?branch=master)](https://travis-ci.org/mic90/go-binproto)
[![go report card](https://goreportcard.com/badge/github.com/mic90/go-binproto)](https://goreportcard.com/report/github.com/mic90/go-binproto)
[![coverage](https://gocover.io/_badge/github.com/mic90/go-binproto)](https://gocover.io/github.com/mic90/go-binproto)
[![godocs](https://godoc.org/github.com/mic90/go-binproto?status.svg)](https://godoc.org/github.com/mic90/go-binproto) 

This package provides simple binary protocol implementation in Golang. 

The protocol is intended to be used on low-memory devices, like MIPS processors, so it's written in such a manner to maintan low memory-usage and minimize memory reallocation between operations.

## How it works ##
Protocol uses COBS encoding with Fletcher CRC16 checksum. Source message is first concatenated with its checksum and then encoded using COBS.

Thanks to the used encoding method, the resulting data contains only one '0' sign - at the frame end, so it's easy to check where each frame is ending.

## Memory usage ##
Library contains internal memory buffer to which the encoded/decoded messages are written, so each call to encode or decode will overwrite last data. 

Internal memory buffer will grow only if its required to store a message bigger than its current length.

To obtain a copy of the last result use the **Copy** method. !This method will allocate new memory for the result data on each call!.

```bash
BenchmarkCache_Encode-4             	100000000	        14.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkCache_Decode-4             	100000000	        15.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkBinProto_Encode-4          	20000000	        73.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkBinProto_Decode-4          	20000000	        69.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkCobsEncode-4               	100000000	        21.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkCobsDecode-4               	100000000	        17.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkFletcher16-4               	100000000	        21.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkWriteReadShouldSucceed-4   	5000000	                313  ns/op	       0 B/op	       0 allocs/op
```

## Thread safety ##
Currently this library is not thread-safe, but it should be fairly easy to implement using for example mutexes.

## Usage ##
```golang
proto := NewBinProto()
src := []byte{1, 1, 1, 0, 0, 1, 5, 12, 44}
encoded, _ := proto.Encode(src)
// to save data for later use
encodedCopy := proto.Copy()
...
decoded, _ := proto.Decode(encodedCopy)
// to save data for later use
decodedCopy := proto.Copy()
```
