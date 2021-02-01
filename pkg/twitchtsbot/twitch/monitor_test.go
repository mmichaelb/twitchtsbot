package twitch

import (
	"context"
	"errors"
	"github.com/mmichaelb/twitchtsbot/pkg/twitchtsbot/testutil"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

const (
	testStreamId1 = "1"
	testStreamId2 = "2"
)

var (
	defaultOkStreamsResponse = helix.StreamsResponse{ResponseCommon: helix.ResponseCommon{StatusCode: http.StatusOK}}
)

func TestMonitor_StartSingleStream(t *testing.T) {
	mockClient := new(testApiClient)
	response := defaultOkStreamsResponse
	response.Data.Streams = []helix.Stream{{UserID: testStreamId1}}
	mockClient.On("GetStreams", &helix.StreamsParams{
		UserIDs: []string{testStreamId1},
		Type:    "live",
	}).Return(&response, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyChan := make(chan *UserState)
	defer close(notifyChan)
	monitor := NewMonitor(mockClient, []string{testStreamId1}, time.Second, ctx, notifyChan)
	go monitor.Start()
	assertStreamerStates(t, notifyChan, map[string]StreamerStatus{testStreamId1: StreamerStatusLive})
	response.Data.Streams = nil
	assertStreamerStates(t, notifyChan, map[string]StreamerStatus{testStreamId1: StreamerStatusOffline})
}

func TestMonitor_StartMultipleStreams(t *testing.T) {
	mockClient := new(testApiClient)
	response := defaultOkStreamsResponse
	response.Data.Streams = []helix.Stream{{UserID: testStreamId1}}
	mockClient.On("GetStreams", &helix.StreamsParams{
		UserIDs: []string{testStreamId1, testStreamId2},
		Type:    "live",
	}).Return(&response, nil)
	response.Data.Streams = []helix.Stream{{UserID: testStreamId1}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyChan := make(chan *UserState)
	defer close(notifyChan)
	monitor := NewMonitor(mockClient, []string{testStreamId1, testStreamId2}, time.Second, ctx, notifyChan)
	go monitor.Start()
	assertStreamerStates(t, notifyChan, map[string]StreamerStatus{
		testStreamId1: StreamerStatusLive,
		testStreamId2: StreamerStatusOffline,
	})
	response.Data.Streams = []helix.Stream{{UserID: testStreamId2}}
	assertStreamerStates(t, notifyChan, map[string]StreamerStatus{
		testStreamId1: StreamerStatusOffline,
		testStreamId2: StreamerStatusLive,
	})
}

func TestMonitor_GetState(t *testing.T) {
	mockClient := new(testApiClient)
	response := defaultOkStreamsResponse
	response.Data.Streams = []helix.Stream{{UserID: testStreamId1}}
	mockClient.On("GetStreams", &helix.StreamsParams{
		UserIDs: []string{testStreamId1, testStreamId2},
		Type:    "live",
	}).Return(&response, nil)
	response.Data.Streams = []helix.Stream{{UserID: testStreamId1}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyChan := make(chan *UserState)
	defer close(notifyChan)
	monitor := NewMonitor(mockClient, []string{testStreamId1, testStreamId2}, time.Second, ctx, notifyChan)
	go monitor.Start()
	go func() {
		for {
			<-notifyChan
		}
	}()
	testutil.WaitForMethodCall(t, &mockClient.Mock, "GetStreams", 1, 10, time.Millisecond*400)
	state, ok := monitor.GetState(testStreamId1)
	assert.True(t, ok, "expected GetState to return state")
	assert.Equal(t, state.ID, testStreamId1, "expected GetState to return correct state id")
	assert.Equal(t, state.StreamerStatus, StreamerStatusLive, "expected GetState to return streamer live status")
}

func TestMonitor_StartInvalidCode(t *testing.T) {
	logger, hook := test.NewNullLogger()
	Log = logger
	mockClient := new(testApiClient)
	response := defaultOkStreamsResponse
	response.StatusCode = http.StatusInternalServerError
	mockClient.On("GetStreams", &helix.StreamsParams{
		UserIDs: []string{testStreamId1, testStreamId2},
		Type:    "live",
	}).Return(&response, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyChan := make(chan *UserState)
	monitor := NewMonitor(mockClient, []string{testStreamId1, testStreamId2}, time.Second, ctx, notifyChan)
	go monitor.Start()
	testutil.WaitForMethodCall(t, &mockClient.Mock, "GetStreams", 1, 10, time.Millisecond*400)
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level, "last entry level should be of type ErrorLevel")
}

func TestMonitor_StartClientError(t *testing.T) {
	logger, hook := test.NewNullLogger()
	Log = logger
	mockClient := new(testApiClient)
	mockClient.On("GetStreams", &helix.StreamsParams{
		UserIDs: []string{testStreamId1, testStreamId2},
		Type:    "live",
	}).Return(nil, errors.New("test error"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyChan := make(chan *UserState)
	monitor := NewMonitor(mockClient, []string{testStreamId1, testStreamId2}, time.Second, ctx, notifyChan)
	go monitor.Start()
	testutil.WaitForMethodCall(t, &mockClient.Mock, "GetStreams", 1, 10, time.Millisecond*400)
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
}

func assertStreamerStates(t *testing.T, notifyChan chan *UserState, statusMap map[string]StreamerStatus) {
	for i := 0; i < len(statusMap); i++ {
		state := <-notifyChan
		expectedStatus, ok := statusMap[state.ID]
		assert.Truef(t, ok, "unexpected user state from notifyChan: %s", state.ID)
		assert.Equalf(t, expectedStatus, state.StreamerStatus, "unexpected streamer status from id %s", state.ID)
	}
}
