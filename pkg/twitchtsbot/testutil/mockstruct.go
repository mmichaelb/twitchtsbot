package testutil

import (
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func WaitForMethodCall(t *testing.T, mock *mock.Mock, methodName string, requiredCalls, tries int, interval time.Duration) {
	var count int
	for i := 0; i < tries; i++ {
		time.Sleep(interval)
		for _, call := range mock.Calls {
			if call.Method == methodName {
				count++
			}
		}
		if count >= requiredCalls {
			return
		}
	}
	t.Fatalf("exceeded wait time: wait for method %s to be executed %d times with an interval of %s",
		methodName, requiredCalls, interval.String())
}
