package twitch

import (
	"errors"
	"sync"
)

var internalClientError = errors.New("internal client error")

type mockApiClient struct {
	*sync.Mutex
	returnErr    bool
	requestCount int
	statusCode   int
	streams      map[string]bool
}
