package main

import (
	"github.com/spf13/viper"
	"time"
)

type accountEntry struct {
	TsIdentifier   string `mapstructure:"ts" yaml:"ts"`
	TwitchUsername string `mapstructure:"twitch" yaml:"twitch"`
}

func setConfigDefaults() {
	viper.SetDefault("teamspeak.url", "<yourbaseurl>")
	viper.SetDefault("teamspeak.apikey", "<yourapikey>")
	viper.SetDefault("teamspeak.serverid", 1)
	viper.SetDefault("twitch.clientid", "<yourclientid>")
	viper.SetDefault("twitch.appaccesstoken", "<yourtoken>")
	viper.SetDefault("accounts", []accountEntry{})
	viper.SetDefault("interval", time.Second)
	viper.SetDefault("servergroupid", -1)
}
