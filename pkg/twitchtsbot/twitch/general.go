package twitch

import "github.com/nicklaw5/helix"

type ApiClient interface {
	GetStreams(params *helix.StreamsParams) (*helix.StreamsResponse, error)
	GetUsers(params *helix.UsersParams) (*helix.UsersResponse, error)
}
