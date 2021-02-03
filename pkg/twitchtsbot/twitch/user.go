package twitch

import (
	"fmt"
	"github.com/nicklaw5/helix"
	"net/http"
)

func RetrieveIDs(client ApiClient, names []string) ([]string, error) {
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
	Log.WithField("idCount", len(ids)).Debugln("Fetched Twitch User IDs!")
	return ids, nil
}
