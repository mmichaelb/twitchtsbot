package main

import (
	"context"
	"errors"
	"flag"
	ts3 "github.com/jkoenig134/go-ts3"
	"github.com/mmichaelb/twitchtsbot/pkg/twitchtsbot/teamspeak"
	"github.com/mmichaelb/twitchtsbot/pkg/twitchtsbot/twitch"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const defaultLogLevel = logrus.InfoLevel

var (
	// application parameters
	logLevel = flag.String("level", "info",
		"Set the logging level. See https://github.com/sirupsen/logrus#level-logging for more details.")
	configPath      = flag.String("config", "./config.yml", "Set the config file path.")
	teamspeakClient ts3.TeamspeakHttpClient
	helixClient     *helix.Client
)

func main() {
	setLogLevel()
	setConfigDefaults()
	loadConfigOrWriteDefault()
	logrus.RegisterExitHandler(func() {
		logrus.Infoln("Shut down twitchtsbot. Goodbye!")
	})
	viper.SetConfigFile(*configPath)
	if err := viper.ReadInConfig(); err != nil {
		logrus.WithError(err).Fatalln("Could not load config file.")
	}
	initializeTeamspeakQueryClient()
	pairs := loadTwitchAccountPairs()
	monitor, notifyChan := initializeTwitchHelixClient(pairs)
	monitor.Start()
	ctx, cancel := context.WithCancel(context.Background())
	logrus.DeferExitHandler(func() {
		logrus.Infoln("Stopping Teamspeak Hook...")
		cancel()
	})
	hook := teamspeak.NewHook(&teamspeakClient, monitor, notifyChan, ctx, pairs, viper.GetInt("servergroupid"))
	if err := hook.Start(); err != nil {
		logrus.WithError(err).Fatalln("Could not start Teamspeak hook.")
	}
	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	<-signalChannel
	logrus.Exit(0)
}

func loadConfigOrWriteDefault() {
	if _, err := os.Stat(*configPath); errors.Is(err, os.ErrNotExist) {
		if err = viper.SafeWriteConfigAs(*configPath); err != nil {
			if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
				logrus.WithField("configPath", *configPath).WithError(err).
					Fatalln("Could not safe write default configuration.")
			}
		} else {
			logrus.WithField("configPath", *configPath).Infoln("Written default config. Please update the values.")
			logrus.Exit(0)
		}
	}
}

func initializeTwitchHelixClient(pairs map[int]string) (*twitch.Monitor, chan *twitch.UserState) {
	twitchLogins := make([]string, 0)
	for _, twitchLogin := range pairs {
		twitchLogins = append(twitchLogins, twitchLogin)
	}
	var err error
	helixClient, err = helix.NewClient(&helix.Options{
		ClientID: viper.GetString("twitch.clientid"),
	})
	if err != nil {
		logrus.WithError(err).Fatalln("Could not authenticate with Twitch Helix API.")
	}
	appAccessToken := viper.GetString("twitch.appaccesstoken")
	valid, _, err := helixClient.ValidateToken(appAccessToken)
	if err != nil {
		logrus.WithError(err).Fatalln("Could not validate Twitch App Access Token.")
	}
	helixClient.SetAppAccessToken(appAccessToken)
	if !valid {
		logrus.WithError(err).WithField("appAccessToken", appAccessToken).Fatalln("Twitch App Access Token is invalid.")
	}
	notifyChan := make(chan *twitch.UserState)
	ctx, cancel := context.WithCancel(context.Background())
	logrus.DeferExitHandler(func() {
		logrus.Infoln("Stopping Helix Twitch Monitor...")
		cancel()
	})
	monitor := twitch.NewMonitor(helixClient, twitchLogins, viper.GetDuration("interval"), ctx, notifyChan)
	return monitor, notifyChan
}

func initializeTeamspeakQueryClient() {
	config := ts3.NewConfig(viper.GetString("teamspeak.url"), viper.GetString("teamspeak.apikey"))
	teamspeakClient = ts3.NewClient(config)
	teamspeakClient.SetServerID(viper.GetInt("teamspeak.serverid"))
	version, err := teamspeakClient.Version()
	if err != nil {
		logrus.WithError(err).Fatalln("Could not retrieve Teamspeak Server version.")
	}
	logrus.WithField("teamspeakVersion", version).Infoln("Retrieved Teamspeak Server version.")
}

func loadTwitchAccountPairs() map[int]string {
	pairs := viper.GetStringSlice("accounts")
	fetchedPairs := make(map[int]string)
	for _, entry := range pairs {
		entrySplit := strings.SplitN(entry, "/", 2)
		if len(entrySplit) != 2 {
			logrus.WithField("accountEntry", entry).Warnln("Could not split account entry on slash (\"/\").")
			continue
		}
		fetchPair(fetchedPairs, entrySplit[0], entrySplit[1])
	}
	logrus.WithField("pairAmount", len(fetchedPairs)).Infoln("Fetched account pairs.")
	return fetchedPairs
}

func fetchPair(fetchedPairs map[int]string, identifier string, twitchLoginName string) {
	var teamspeakDatabaseId int
	var err error
	if teamspeakDatabaseId, err = strconv.Atoi(identifier); err != nil {
		logrus.WithField("identifier", identifier).Traceln("Could not parse int from identifier. Falling back to Teamspeak fetch.")
		fetchedTeamspeakDatabaseId, err := teamspeakClient.ClientGetDbIdFromUid(identifier)
		if err != nil {
			logrus.WithError(err).WithField("identifier", identifier).Warnln("Could not retrieve Teamspeak database id.")
			return
		}
		teamspeakDatabaseId = *fetchedTeamspeakDatabaseId
	}
	fetchedPairs[teamspeakDatabaseId] = twitchLoginName
}

func setLogLevel() {
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		level = defaultLogLevel
		logrus.WithError(err).Errorln("Could not parse error level.")
	}
	logrus.SetLevel(level)
}
