package binproto

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ReadWriterMock struct {
	mock.Mock
}

func (rw *ReadWriterMock) Write(src []byte) (int, error) {
	args := rw.Called(src)
	return args.Int(0), args.Error(1)
}

func (rw *ReadWriterMock) Read(src []byte) (int, error) {
	args := rw.Called(src)
	return args.Int(0), args.Error(1)
}

func TestRetryNoRetriesRequired(t *testing.T) {
	attempts := 5
	delay := 2 * time.Second
	result := int(0)
	expectedResult := int(10)
	err := Retry(attempts, delay, func() (err error) {
		result = 10
		return
	})
	if err != nil {
		t.Errorf("Function ended with error: %v", err)
	}
	if result != expectedResult {
		t.Errorf("Final result value %v is not equal to the expected one %v", result, expectedResult)
	}
}

func TestRetryReturnsError(t *testing.T) {
	attempts := 5
	delay := 0 * time.Second
	err := Retry(attempts, delay, func() (err error) {
		return errors.New("internal error")
	})
	if err == nil {
		t.Error("Function didn't finished with error")
	}
}

func TestWriteReadShouldSucceed(t *testing.T) {
	// GIVEN
	message := []byte("hello")
	encoder := NewProtocolParser()
	encoder.Encode(message)
	encodedMsg := encoder.Copy()
	encodedMsg = append(encodedMsg, byte(0))

	expectedResponse := []byte("world")
	encoder.Encode(expectedResponse)
	encodedResp := encoder.Copy()
	encodedResp = append(encodedResp, byte(0))

	readWriterMock := &ReadWriterMock{}
	readWriterMock.On("Write", encodedMsg).Return(len(encodedMsg), nil)
	readWriterMock.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		bytes := args[0].([]byte)
		copy(bytes, encodedResp)
	}).Return(len(encodedResp), io.EOF)

	readWriter := NewProtocolReadWriter(
		3,
		1*time.Millisecond,
		1*time.Millisecond,
		1*time.Second)

	// WHEN
	response, err := readWriter.RetryWriteRead(readWriterMock, encodedMsg)

	// THEN
	assert.Nil(t, err, "write or read operation failed, but it shouldn't. %v", err)
	assert.Equal(t, response, expectedResponse, "received message is different than expected: %v", response)
}

func TestWriteReadShouldFailIfSourceIsNotZeroEnded(t *testing.T) {
	// GIVEN
	message := []byte("hello")
	encoder := NewProtocolParser()
	encoder.Encode(message)
	encodedMsg := encoder.Copy()

	readWriterMock := &ReadWriterMock{}

	readWriter := NewProtocolReadWriter(
		3,
		1*time.Millisecond,
		1*time.Millisecond,
		1*time.Second)

	// WHEN
	_, err := readWriter.RetryWriteRead(readWriterMock, encodedMsg)

	// THEN
	readWriterMock.AssertNotCalled(t, "Write", mock.Anything)
	readWriterMock.AssertNotCalled(t, "Read", mock.Anything)
	assert.NotNil(t, err, "write or read operation succeeded, but it shouldn't. %v", err)
	assert.Equal(t, ErrSourceNotEndsWithZero, err)
}

func TestWriteReadShouldFailIfWriteFailsWithError(t *testing.T) {
	// GIVEN
	message := []byte("hello")
	encoder := NewProtocolParser()
	encoder.Encode(message)
	encodedMsg := encoder.Copy()
	encodedMsg = append(encodedMsg, 0)

	readWriterMock := &ReadWriterMock{}
	errorMsg := "write failed"
	writeError := errors.New(errorMsg)
	readWriterMock.On("Write", encodedMsg).Return(0, writeError)

	expectedRetryCount := 3

	readWriter := NewProtocolReadWriter(
		expectedRetryCount,
		1*time.Millisecond,
		1*time.Millisecond,
		10*time.Second)

	// WHEN
	_, err := readWriter.RetryWriteRead(readWriterMock, encodedMsg)

	// THEN
	readWriterMock.AssertNumberOfCalls(t, "Write", expectedRetryCount)
	assert.NotNil(t, err)
	assert.Equal(t, writeError, err)
}

func TestWriteReadShouldFailIfNumberOfWrittenBytesIsDifferentThanSrcLength(t *testing.T) {
	// GIVEN
	message := []byte("hello")
	encoder := NewProtocolParser()
	encoder.Encode(message)
	encodedMsg := encoder.Copy()
	encodedMsg = append(encodedMsg, byte(0))

	readWriterMock := &ReadWriterMock{}
	readWriterMock.On("Write", encodedMsg).Return(len(encodedMsg)-1, nil)

	expectedRetryCount := 3

	readWriter := NewProtocolReadWriter(
		expectedRetryCount,
		1*time.Millisecond,
		1*time.Millisecond,
		10*time.Second)

	// WHEN
	_, err := readWriter.RetryWriteRead(readWriterMock, encodedMsg)

	// THEN
	readWriterMock.AssertNumberOfCalls(t, "Write", expectedRetryCount)
	assert.NotNil(t, err)
	assert.Equal(t, ErrWrittenLengthDoesNotMatch, err)
}

func TestWriteReadShouldTimeoutIfNoEndingZeroWasReadFromStream(t *testing.T) {
	// GIVEN
	message := []byte("hello")
	encoder := NewProtocolParser()
	encoder.Encode(message)
	encodedMsg := encoder.Copy()
	encodedMsg = append(encodedMsg, byte(0))

	response := []byte{2}

	readWriterMock := &ReadWriterMock{}
	readWriterMock.On("Write", encodedMsg).Return(len(encodedMsg), nil)
	readWriterMock.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		bytes := args[0].([]byte)
		copy(bytes, response)
	}).Return(len(response), io.EOF)

	expectedRetryCount := 3

	readWriter := NewProtocolReadWriter(
		expectedRetryCount,
		0,
		0,
		50*time.Millisecond)

	// WHEN
	_, err := readWriter.RetryWriteRead(readWriterMock, encodedMsg)
	// THEN
	readWriterMock.AssertNumberOfCalls(t, "Write", expectedRetryCount)
	assert.NotNil(t, err)
	assert.Equal(t, ErrTimeout, err)
}

func TestWriteReadShouldSucceedAndDropMessageAfterFirstZeroSign(t *testing.T) {
	// GIVEN
	message := []byte("hello")
	encoder := NewProtocolParser()
	encoder.Encode(message)
	encodedMsg := encoder.Copy()
	encodedMsg = append(encodedMsg, byte(0))

	expectedResponse := []byte("world")
	encoder.Encode(expectedResponse)
	encodedResp := encoder.Copy()
	encodedResp = append(encodedResp, []byte{0, 2, 2, 2, 2, 0}...)

	readWriterMock := &ReadWriterMock{}
	readWriterMock.On("Write", encodedMsg).Return(len(encodedMsg), nil)
	readWriterMock.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		bytes := args[0].([]byte)
		copy(bytes, encodedResp)
	}).Return(len(encodedResp), io.EOF)

	readWriter := NewProtocolReadWriter(
		3,
		0,
		0,
		10*time.Second)

	// WHEN
	response, err := readWriter.RetryWriteRead(readWriterMock, encodedMsg)
	// THEN
	assert.Nil(t, err)
	readWriterMock.AssertNumberOfCalls(t, "Write", 1)
	assert.Equal(t, expectedResponse, response)
}

type ReadWriterBenchmarkMock struct {
	data []byte
}

func (rw *ReadWriterBenchmarkMock) Write(src []byte) (int, error) {
	return len(src), nil
}

func (rw *ReadWriterBenchmarkMock) Read(src []byte) (int, error) {
	copy(src, rw.data)
	return len(rw.data), io.EOF
}

func BenchmarkWriteReadShouldSucceed(b *testing.B) {
	// GIVEN
	message := []byte("hello")
	encoder := NewProtocolParser()
	encoder.Encode(message)
	encodedMsg := encoder.Copy()
	encodedMsg = append(encodedMsg, byte(0))

	expectedResponse := []byte("world")
	encoder.Encode(expectedResponse)
	encodedResp := encoder.Copy()
	encodedResp = append(encodedResp, byte(0))

	readWriterMock := &ReadWriterBenchmarkMock{encodedResp}

	readWriter := NewProtocolReadWriter(
		3,
		0*time.Millisecond,
		0*time.Millisecond,
		1*time.Second)

	// WHEN
	for i := 0; i < b.N; i++ {
		_, err := readWriter.RetryWriteRead(readWriterMock, encodedMsg)
		if err != nil {
			b.Fail()
		}
	}
}
