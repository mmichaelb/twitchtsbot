package twitch

import (
	"context"
	"errors"
	"fmt"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

type StreamerStatus int

const (
	StreamerStatusOffline StreamerStatus = iota
	StreamerStatusLive
)

type UserState struct {
	ID             string
	StreamerStatus StreamerStatus
}

type Monitor struct {
	*sync.Mutex
	States     map[string]*UserState
	Client     ApiClient
	UserIds    []string
	Interval   time.Duration
	Context    context.Context
	NotifyChan chan *UserState
}

func NewMonitor(client ApiClient, userIds []string, interval time.Duration, context context.Context, notifyChan chan *UserState) *Monitor {
	monitor := &Monitor{
		Mutex:      &sync.Mutex{},
		Client:     client,
		UserIds:    userIds,
		Interval:   interval,
		Context:    context,
		NotifyChan: notifyChan,
	}
	return monitor
}

func (monitor *Monitor) Start() {
	Log.WithFields(logrus.Fields{
		"interval":     monitor.Interval.String(),
		"userIdNumber": len(monitor.UserIds),
	}).Infoln("Starting Twitch stream monitor")
	defer func() {
		Log.Infoln("Stopped Twitch stream monitor.")
	}()
	for {
		select {
		case <-time.After(monitor.Interval):
			if err := monitor.updateUserStates(); err != nil {
				Log.WithError(err).Errorln("Could not update streamer states!")
			}
		case <-monitor.Context.Done():
			return
		}
	}
}

func (monitor *Monitor) updateUserStates() error {
	resp, err := monitor.Client.GetStreams(&helix.StreamsParams{
		Type:       "live",
		UserLogins: monitor.UserIds,
	})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("twitch api returned enexpected status code: %d", resp.StatusCode))
	}
	Log.WithField("streams", resp.Data.Streams).Debugln("Fetched live streams from Twitch API.")
	monitor.updateStreamerStates(resp.Data.Streams)
	return nil
}

func (monitor *Monitor) updateStreamerStates(streams []helix.Stream) {
	monitor.Lock()
	defer monitor.Unlock()
	if monitor.States == nil {
		monitor.initializeStreamerStates(streams)
		return
	}
	// check for default states
	for userId, state := range monitor.States {
		fetchedStatus := StreamerStatusOffline
		for _, stream := range streams {
			if stream.UserID == userId {
				fetchedStatus = StreamerStatusLive
			}
		}
		if state.StreamerStatus != fetchedStatus {
			state.StreamerStatus = fetchedStatus
			monitor.NotifyChan <- state
		}
	}
}

func (monitor *Monitor) initializeStreamerStates(streams []helix.Stream) {
	monitor.States = make(map[string]*UserState, len(monitor.UserIds))
	for _, userId := range monitor.UserIds {
		state := &UserState{
			ID:             userId,
			StreamerStatus: StreamerStatusOffline,
		}
		for _, stream := range streams {
			if stream.UserID == userId {
				state.StreamerStatus = StreamerStatusLive
			}
		}
		monitor.States[userId] = state
		monitor.NotifyChan <- state
	}
}

func (monitor *Monitor) GetState(id string) (*UserState, bool) {
	monitor.Lock()
	defer monitor.Unlock()
	state, ok := monitor.States[id]
	return state, ok
}
