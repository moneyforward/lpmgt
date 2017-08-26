package service

import (
	lc "lastpass_provisioning/lastpass_client"
	"net/http"
)

type EventService struct {
	client  *lc.LastPassClient
	command string
	data    interface{}
}

// NewEventService creates a new EventService
func NewEventService(client *lc.LastPassClient) (s *UserService) {
	return &UserService{client: client}
}

func (s *EventService) doRequest() (*http.Response, error) {

	return s.client.DoRequest(s.command, s.data)
}
