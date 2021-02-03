package twitch

import (
	"context"
	"errors"
	"fmt"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"sync"
	"time"
)

type StreamerStatus int

const (
	StreamerStatusOffline StreamerStatus = iota
	StreamerStatusLive
	changesRequired = 3
)

type UserState struct {
	UserLogin      string
	StreamerStatus StreamerStatus
}

type ChangeState struct {
	Status StreamerStatus
	Count  int
}

type Monitor struct {
	*sync.Mutex
	States       map[string]*UserState
	ChangeActive map[string]*ChangeState
	Client       ApiClient
	UserLogins   []string
	Interval     time.Duration
	Context      context.Context
	NotifyChan   chan *UserState
}

func NewMonitor(client ApiClient, userLogins []string, interval time.Duration, context context.Context, notifyChan chan *UserState) *Monitor {
	monitor := &Monitor{
		Mutex:      &sync.Mutex{},
		Client:     client,
		UserLogins: userLogins,
		Interval:   interval,
		Context:    context,
		NotifyChan: notifyChan,
	}
	return monitor
}

func (monitor *Monitor) Start() {
	Log.WithFields(logrus.Fields{
		"interval":        monitor.Interval.String(),
		"userLoginNumber": len(monitor.UserLogins),
	}).Infoln("Starting Twitch stream monitor")
	go func() {
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
	}()
}

func (monitor *Monitor) updateUserStates() error {
	resp, err := monitor.Client.GetStreams(&helix.StreamsParams{
		Type:       "live",
		UserLogins: monitor.UserLogins,
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
	for userLogin, state := range monitor.States {
		fetchedStatus := StreamerStatusOffline
		for _, stream := range streams {
			if strings.EqualFold(stream.UserName, userLogin) {
				fetchedStatus = StreamerStatusLive
			}
		}
		if state.StreamerStatus != fetchedStatus {
			if changeStatus, ok := monitor.ChangeActive[state.UserLogin]; !ok {
				monitor.ChangeActive[state.UserLogin] = &ChangeState{Status: fetchedStatus, Count: 1}
				continue
			} else if changeStatus.Status == fetchedStatus {
				if changeStatus.Count >= changesRequired {
					state.StreamerStatus = fetchedStatus
					monitor.NotifyChan <- state
				} else {
					changeStatus.Count++
					continue
				}
			}
			delete(monitor.ChangeActive, state.UserLogin)
		}
	}
}

func (monitor *Monitor) initializeStreamerStates(streams []helix.Stream) {
	monitor.States = make(map[string]*UserState, len(monitor.UserLogins))
	monitor.ChangeActive = make(map[string]*ChangeState, len(monitor.UserLogins))
	for _, userLogin := range monitor.UserLogins {
		state := &UserState{
			UserLogin:      userLogin,
			StreamerStatus: StreamerStatusOffline,
		}
		for _, stream := range streams {
			if strings.EqualFold(stream.UserName, userLogin) {
				state.StreamerStatus = StreamerStatusLive
			}
		}
		monitor.States[userLogin] = state
		monitor.NotifyChan <- state
	}
}

func (monitor *Monitor) GetState(userLogin string) (*UserState, bool) {
	monitor.Lock()
	defer monitor.Unlock()
	state, ok := monitor.States[userLogin]
	return state, ok
}
