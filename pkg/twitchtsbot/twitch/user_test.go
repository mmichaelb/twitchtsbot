package twitch

import (
	"context"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync"
	"testing"
	"time"
)

type mockUserApiClient struct {
	*sync.Mutex
	returnErr    bool
	requestCount int
	statusCode   int
	users        map[string]helix.User
}

func (client *mockUserApiClient) GetUsers(params *helix.UsersParams) (*helix.UsersResponse, error) {
	if client.returnErr {
		return nil, internalClientError
	}
	client.requestCount++
	client.Lock()
	defer client.Unlock()
	resp := &helix.UsersResponse{
		ResponseCommon: helix.ResponseCommon{
			StatusCode: client.statusCode,
		},
		Data: helix.ManyUsers{},
	}
	if client.users == nil {
		return resp, nil
	}
	for userId, user := range client.users {
		for _, login := range params.Logins {
			if login == userId {
				resp.Data.Users = append(resp.Data.Users, user)
			}
		}
	}
	return resp, nil
}

func TestUser_RetrieveIDs(t *testing.T) {
	mockClient := &mockUserApiClient{
		Mutex:      &sync.Mutex{},
		statusCode: http.StatusOK,
		users: map[string]helix.User{
			"testuser": {ID: "1"},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyChan := make(chan *UserState)
	defer close(notifyChan)
	ids, err := RetrieveIDs(mockClient, []string{"testuser"})
	assert.Nil(t, err)
	assert.Equal(t, []string{}, ids)
	go monitor.Start()
	assertStreamerStates(t, notifyChan, map[string]StreamerStatus{"1": StreamerStatusLive})
	mockClient.updateStreamStatus("1", false)
	assertStreamerStates(t, notifyChan, map[string]StreamerStatus{"1": StreamerStatusOffline})
}

func TestMonitor_StartMultipleStreams(t *testing.T) {
	mockClient := &mockStreamApiClient{
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
	mockClient := &mockStreamApiClient{
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
	mockClient := &mockStreamApiClient{
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
	mockClient := &mockStreamApiClient{
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
