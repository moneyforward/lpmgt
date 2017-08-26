package service

import (
	"lastpass_provisioning/lastpassclient"
	"net/http"
)

type EventService struct {
	client  *lastpassclient.LastpassClient
	command string
	data    interface{}
}

// NewEventService creates a new EventService
func NewEventService(client *lastpassclient.LastpassClient) (s *UserService) {
	return &UserService{client: client}
}

func (s *EventService) doRequest() (*http.Response, error) {

	return s.client.DoRequest(s.command, s.data)
}
