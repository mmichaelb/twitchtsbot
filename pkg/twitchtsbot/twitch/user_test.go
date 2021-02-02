package twitch

import (
	"errors"
	"github.com/nicklaw5/helix"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var (
	defaultOkUsersResponse = helix.UsersResponse{ResponseCommon: helix.ResponseCommon{StatusCode: http.StatusOK}}
	defaultUserParams      = helix.UsersParams{Logins: []string{"testuser"}}
)

func TestRetrieveIDs(t *testing.T) {
	client := new(testApiClient)
	response := defaultOkUsersResponse
	response.Data = helix.ManyUsers{Users: []helix.User{{ID: "0"}}}
	client.On("GetUsers", &defaultUserParams).Return(&response, nil)
	ids, err := RetrieveIDs(client, []string{"testuser"})
	client.AssertNumberOfCalls(t, "GetUsers", 1)
	assert.Nil(t, err, "returned err for retrieve ids method is not nil")
	assert.Equal(t, []string{"0"}, ids, "returned ids from retrieve ids method is invalid")
}

func TestRetrieveIDs_Error(t *testing.T) {
	client := new(testApiClient)
	testErr := errors.New("test error")
	client.On("GetUsers", &defaultUserParams).Return(nil, testErr)
	ids, err := RetrieveIDs(client, []string{"testuser"})
	client.AssertNumberOfCalls(t, "GetUsers", 1)
	assert.Nil(t, ids, "returned ids for retrieve ids method should be nil")
	assert.Equal(t, testErr, err, "returned error for retrieve ids should be equal to test error")
}

func TestRetrieveIDs_InvalidStatusCode(t *testing.T) {
	client := new(testApiClient)
	response := defaultOkUsersResponse
	response.StatusCode = http.StatusInternalServerError
	client.On("GetUsers", &defaultUserParams).Return(&response, nil)
	ids, err := RetrieveIDs(client, []string{"testuser"})
	client.AssertNumberOfCalls(t, "GetUsers", 1)
	assert.Nil(t, ids, "returned ids for retrieve ids method should be nil")
	assert.NotNil(t, err, "returned error for retrieve ids should not be nil")
}
