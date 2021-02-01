package twitch

import (
	"github.com/nicklaw5/helix"
	"github.com/stretchr/testify/mock"
)

type userApiClient struct {
	mock.Mock
}

func (client *userApiClient) GetStreams(params *helix.StreamsParams) (*helix.StreamsResponse, error) {
	args := client.Called(params)
	resp := args.Get(0)
	err := args.Error(1)
	if resp != nil {
		return resp.(*helix.StreamsResponse), err
	}
	return nil, err
}

func (client *userApiClient) GetUsers(params *helix.UsersParams) (*helix.UsersResponse, error) {
	args := client.Called(params)
	resp := args.Get(0)
	err := args.Error(1)
	if resp != nil {
		return resp.(*helix.UsersResponse), err
	}
	return nil, err
}
