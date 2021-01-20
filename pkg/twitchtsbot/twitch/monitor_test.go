package twitch

import (
	"context"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync"
	"testing"
	"time"
)

func (client *mockApiClient) GetStreams(params *helix.StreamsParams) (*helix.StreamsResponse, error) {
	if client.returnErr {
		return nil, internalClientError
	}
	client.requestCount++
	client.Lock()
	defer client.Unlock()
	resp := &helix.StreamsResponse{
		ResponseCommon: helix.ResponseCommon{
			StatusCode: client.statusCode,
		},
		Data: helix.ManyStreams{
			Streams: make([]helix.Stream, 0),
		},
	}
	if client.streams == nil {
		return resp, nil
	}
	for userId, live := range client.streams {
		if !live {
			continue
		}
		for _, paramUserId := range params.UserIDs {
			if paramUserId == userId {
				resp.Data.Streams = append(resp.Data.Streams, helix.Stream{UserID: userId})
			}
		}
	}
	return resp, nil
}

func (client *mockApiClient) updateStreamStatus(user string, live bool) {
	client.Lock()
	defer client.Unlock()
	client.streams[user] = live
}

func TestMonitor_StartSingleStream(t *testing.T) {
	mockClient := &mockApiClient{
		Mutex:      &sync.Mutex{},
		statusCode: http.StatusOK,
		streams: map[string]bool{
			"1": true,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyChan := make(chan *UserState)
	defer close(notifyChan)
	monitor := NewMonitor(mockClient, []string{"1"}, time.Second, ctx, notifyChan)
	go monitor.Start()
	assertStreamerStates(t, notifyChan, map[string]StreamerStatus{"1": StreamerStatusLive})
	mockClient.updateStreamStatus("1", false)
	assertStreamerStates(t, notifyChan, map[string]StreamerStatus{"1": StreamerStatusOffline})
}

func TestMonitor_StartMultipleStreams(t *testing.T) {
	mockClient := &mockApiClient{
		Mutex:      &sync.Mutex{},
		statusCode: http.StatusOK,
		streams: map[string]bool{
			"1": true,
			"2": false,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyChan := make(chan *UserState)
	defer close(notifyChan)
	monitor := NewMonitor(mockClient, []string{"1", "2"}, time.Second, ctx, notifyChan)
	go monitor.Start()
	assertStreamerStates(t, notifyChan, map[string]StreamerStatus{
		"1": StreamerStatusLive,
		"2": StreamerStatusOffline,
	})
	mockClient.updateStreamStatus("1", false)
	mockClient.updateStreamStatus("2", true)
	assertStreamerStates(t, notifyChan, map[string]StreamerStatus{
		"1": StreamerStatusOffline,
		"2": StreamerStatusLive,
	})
}

func TestMonitor_GetState(t *testing.T) {
	mockClient := &mockApiClient{
		Mutex:      &sync.Mutex{},
		statusCode: http.StatusOK,
		streams: map[string]bool{
			"1": true,
			"2": false,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyChan := make(chan *UserState)
	defer close(notifyChan)
	monitor := NewMonitor(mockClient, []string{"1", "2"}, time.Second, ctx, notifyChan)
	go monitor.Start()
	go func() {
		for {
			<-notifyChan
		}
	}()
	waitForMonitorRequest(mockClient)
	state, ok := monitor.GetState("1")
	assert.True(t, ok, "expected GetState to return state")
	assert.Equal(t, state.ID, "1", "expected GetState to return correct state id")
	assert.Equal(t, state.StreamerStatus, StreamerStatusLive, "expected GetState to return streamer live status")
}

func TestMonitor_StartInvalidCode(t *testing.T) {
	logger, hook := test.NewNullLogger()
	Log = logger
	mockClient := &mockApiClient{
		Mutex:      &sync.Mutex{},
		statusCode: http.StatusUnauthorized,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyChan := make(chan *UserState)
	monitor := NewMonitor(mockClient, []string{"1", "2"}, time.Second, ctx, notifyChan)
	go monitor.Start()
	waitForMonitorRequest(mockClient)
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
}

func TestMonitor_StartClientError(t *testing.T) {
	logger, hook := test.NewNullLogger()
	Log = logger
	mockClient := &mockApiClient{
		Mutex:     &sync.Mutex{},
		returnErr: true,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyChan := make(chan *UserState)
	monitor := NewMonitor(mockClient, []string{"1", "2"}, time.Second, ctx, notifyChan)
	go monitor.Start()
	waitForMonitorRequest(mockClient)
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
}

func waitForMonitorRequest(mockClient *mockApiClient) {
	// wait for monitor to request the first time
	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 400)
		if mockClient.requestCount >= 1 {
			break
		}
	}
}

func assertStreamerStates(t *testing.T, notifyChan chan *UserState, statusMap map[string]StreamerStatus) {
	for i := 0; i < len(statusMap); i++ {
		state := <-notifyChan
		expectedStatus, ok := statusMap[state.ID]
		assert.Truef(t, ok, "unexpected user state from notifyChan: %s", state.ID)
		assert.Equalf(t, expectedStatus, state.StreamerStatus, "unexpected streamer status from id %s", state.ID)
	}
}
