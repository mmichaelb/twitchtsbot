package twitch

import (
	"fmt"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
	"net/http"
)

func RetrieveIDs(client *helix.Client, names []string) ([]string, error) {
	Log.WithField("nameCount", len(names)).Debugln("Fetching Twitch User IDs...")
	resp, err := client.GetUsers(&helix.UsersParams{
		Logins: names,
	})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received unexpected status code from twitch api: %d", resp.StatusCode)
	}
	ids := make([]string, 0)
	for _, user := range resp.Data.Users {
		ids = append(ids, user.ID)
	}
	logrus.WithField("idCount", len(ids)).Debugln("Fetched Twitch User IDs!")
	return ids, nil
}
