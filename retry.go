package binproto

import (
	"bytes"
	"errors"
	"io"
	"time"
)

// ProtocolReadWriter is a helper class to ease i/o operations with encoded data
// It contains internal protocol decoder which will decode incoming messages
type ProtocolReadWriter struct {
	decoder *CachedProtocolParser

	retryCount  int
	retryDelay  time.Duration
	readDelay   time.Duration
	readTimeout time.Duration
	timeout     *time.Timer

	readBuffer    bytes.Buffer
	messageBuffer bytes.Buffer
}

var (
	// ErrSourceNotEndsWithZero is returned when source bytes does not ends with 0 sign
	ErrSourceNotEndsWithZero = errors.New("source data does not ends with 0")
	// ErrWrittenLengthDoesNotMatch is returned when number of written bytes is different than source length
	ErrWrittenLengthDoesNotMatch = errors.New("number of written bytes is different than expected")
	// ErrNoDataRead is returned when no data was read from input stream
	ErrNoDataRead = errors.New("no data was read from input stream")
	// ErrTimeout is returned when write/read cycle was unable to finish in given time
	ErrTimeout = errors.New("write/read operation timed out")
)

func NewProtocolReadWriter(retryCount int, retryDelay, readDelay, readTimeout time.Duration) *ProtocolReadWriter {
	return &ProtocolReadWriter{NewCachedProtocolParser(), retryCount, retryDelay, readDelay,
		readTimeout, time.NewTimer(0), bytes.Buffer{}, bytes.Buffer{}}
}

func (p *ProtocolReadWriter) RetryWriteRead(readWriter io.ReadWriter, src []byte) ([]byte, error) {
	sourceLength := len(src)
	if src[sourceLength-1] != 0 {
		return nil, ErrSourceNotEndsWithZero
	}

	p.messageBuffer.Reset()

	err := Retry(p.retryCount, p.retryDelay, func() error {
		p.resetTimer()
		p.readBuffer.Reset()

		written, err := readWriter.Write(src)
		if err != nil {
			return err
		}
		if written != sourceLength {
			return ErrWrittenLengthDoesNotMatch
		}

		stopRead := false
		timeout := false
		lastReadBytes := int64(0)
		zeroIndex := 0
		for stopRead == false {
			select {
			default:
				readLen, inErr := p.readBuffer.ReadFrom(readWriter)
				if inErr != nil {
					return inErr
				}
				// no data in input stream -> stop reader loop
				if readLen == 0 {
					stopRead = true
				}
				// data contains 0 sign, which means we get whole message -> stop reader loop
				lastReadBytes += readLen
				for i, value := range p.readBuffer.Bytes() {
					if value == 0 {
						zeroIndex = i
						stopRead = true
						break
					}
				}
			case <-p.timeout.C:
				stopRead = true
				timeout = true
				break
			}
		}

		// timer was fired, we might not receive data at all or do not receive the ending 0 sign
		if timeout {
			return ErrTimeout
		}
		// no data was read from input -> repeat write/read cycle
		if lastReadBytes == 0 {
			return ErrNoDataRead
		}
		// get message without ending 0 sign
		message := p.readBuffer.Bytes()[:zeroIndex]
		decodedMessage, err := p.decoder.Decode(message)
		if err != nil {
			return err
		}
		_, err = p.messageBuffer.Write(decodedMessage)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return p.messageBuffer.Bytes(), nil
}

func (p *ProtocolReadWriter) resetTimer() bool {
	return p.timeout.Reset(p.readTimeout)
}

// Retry will try to run callback function
// If function fails with any error, execution will be retried after given sleep time
// If all tries will fail, the last error returned from callback will be returned
func Retry(attempts int, sleep time.Duration, callback func() error) error {
	var lastErr error
	for i := 0; ; i++ {
		lastErr = callback()
		if lastErr == nil {
			return nil
		}
		if i >= (attempts - 1) {
			break
		}
		time.Sleep(sleep)
	}
	return lastErr
}
