package main

import (
	"lastpass_provisioning/api"
	"net/http"
)

type UserService struct {
	client  *LastpassClient
	command string
	data    interface{}
}

func (s *UserService) BatchAdd(users []api.User) (error) {
	s.command = "batchadd"
	s.data = users
	_, err := s.DoRequest()
	return err
}

func (s *UserService) GetAdminUserData() ([]api.User, error) {
	s.command = "getuserdata"
	s.data = api.User{IsAdmin: true}
	res, err := s.DoRequest()
	if err != nil {
		return nil, err
	}

	var AdminUsers api.Users
	err = DecodeBody(res, &AdminUsers)
	if err != nil {
		return nil, err
	}
	return AdminUsers.GetUsers(), nil
}

func NewService(client *LastpassClient) (s *UserService) {
	return &UserService{client: client}
}

func (s *UserService) DoRequest() (*http.Response, error) {
	return s.client.DoRequest(s.command, s.data)
}
