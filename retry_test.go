package binproto

import (
	"errors"
	"testing"
	"time"
)

func TestRetryNoRetrieRequired(t *testing.T) {
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
