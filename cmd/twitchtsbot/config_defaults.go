package main

import (
	"github.com/spf13/viper"
	"time"
)

func setConfigDefaults() {
	viper.SetDefault("teamspeak.url", "<yourbaseurl>")
	viper.SetDefault("teamspeak.apikey", "<yourapikey>")
	viper.SetDefault("teamspeak.serverid", 1)
	viper.SetDefault("twitch.clientid", "<yourclientid>")
	viper.SetDefault("twitch.appaccesstoken", "<yourtoken>")
	viper.SetDefault("accounts", map[string]string{})
	viper.SetDefault("interval", time.Second)
	viper.SetDefault("servergroupid", -1)
}
