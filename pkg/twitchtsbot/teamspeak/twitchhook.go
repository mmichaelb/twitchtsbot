package teamspeak

import (
	"context"
	ts3 "github.com/jkoenig134/go-ts3"
	"github.com/mmichaelb/twitchtsbot/pkg/twitchtsbot/twitch"
	"github.com/sirupsen/logrus"
)

type TwitchUpdateHook struct {
	TsClient   *ts3.TeamspeakHttpClient
	Monitor    *twitch.Monitor
	NotifyChan chan *twitch.UserState
	Ctx        context.Context
	// teamspeak database identifier: twitch login id
	UserMapping   map[int]string
	ServerGroupId int
}

func NewHook(teamspeakHttpClient *ts3.TeamspeakHttpClient, monitor *twitch.Monitor, notifyChan chan *twitch.UserState,
	ctx context.Context, userMapping map[int]string, serverGroupId int) *TwitchUpdateHook {
	return &TwitchUpdateHook{
		TsClient:      teamspeakHttpClient,
		Monitor:       monitor,
		NotifyChan:    notifyChan,
		Ctx:           ctx,
		UserMapping:   userMapping,
		ServerGroupId: serverGroupId,
	}
}

func (hook *TwitchUpdateHook) Start() error {
	err := hook.TsClient.SubscribeEvent(ts3.NotifyClientEnterView, hook.enterClientHook)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-hook.Ctx.Done():
				return
			case state := <-hook.NotifyChan:
				go func() {
					teamspeakDatabaseId, ok := hook.retrieveTeamspeakDatabaseId(state.ID)
					if !ok {
						return
					}
					hook.updateTeamspeakRank(teamspeakDatabaseId, state)
				}()
			}
		}
	}()
	return nil
}

func (hook *TwitchUpdateHook) retrieveTeamspeakDatabaseId(searchLoginId string) (int, bool) {
	for databaseId, loginId := range hook.UserMapping {
		if loginId == searchLoginId {
			return databaseId, true
		}
	}
	return 0, false
}

func (hook *TwitchUpdateHook) enterClientHook(event *ts3.ClientEnterViewEvent) {
	// check if login is voice client
	if event.ClientType != 0 {
		return
	}
	twitchLoginId, ok := hook.UserMapping[event.ClientDatabaseId]
	if !ok {
		return
	}
	state, ok := hook.Monitor.GetState(twitchLoginId)
	if !ok {
		return
	}
	hook.updateTeamspeakRank(event.ClientDatabaseId, state)
}

func (hook *TwitchUpdateHook) updateTeamspeakRank(clientDbId int, state *twitch.UserState) {
	add := state.StreamerStatus == twitch.StreamerStatusLive
	var hasServerGroup bool
	members, err := hook.TsClient.ServerGroupClientList(hook.ServerGroupId)
	if err != nil {
		logrus.WithError(err).WithField("serverGroupId", hook.ServerGroupId).Errorln("could not retrieve server group members")
		return
	} else {
		for _, member := range *members {
			if member.ClientDbId != clientDbId {
				continue
			}
			if add {
				return
			} else {
				hasServerGroup = true
				break
			}
		}
	}
	if !add && !hasServerGroup {
		Log.WithFields(logrus.Fields{"clientDbId": clientDbId, "serverGroupId": hook.ServerGroupId}).
			Infoln("client does not have the server group which should be removed")
		return
	}
	if add {
		err := hook.TsClient.ServerGroupAddClient(hook.ServerGroupId, clientDbId)
		if err != nil {
			Log.WithFields(logrus.Fields{"clientDbId": clientDbId, "serverGroupId": hook.ServerGroupId}).
				WithError(err).Warnln("could not add client to server group")
		}
	} else {
		err := hook.TsClient.ServerGroupDeleteClient(hook.ServerGroupId, clientDbId)
		if err != nil {
			Log.WithFields(logrus.Fields{"clientDbId": clientDbId, "serverGroupId": hook.ServerGroupId}).
				WithError(err).Warnln("could not remove client from server group")
		}
	}
}
