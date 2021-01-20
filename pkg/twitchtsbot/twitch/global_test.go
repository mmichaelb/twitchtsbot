package twitch

import "sync"

type mockApiClient struct {
	*sync.Mutex
	returnErr    bool
	requestCount int
	statusCode   int
	streams      map[string]bool
}
