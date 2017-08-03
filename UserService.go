package main

import (
	"net/http"
	"lastpass_provisioning/api"
)

type UserService struct {
	client *LastpassClient
	command string
	data api.User
}

func (s *UserService) GetAdminUserData() ([]api.User, error) {
	s.command = "getuserdata"
	s.data = api.User{IsAdmin: true}
	res, err := s.DoRequest()
	if err != nil {
		return nil, err
	}

	var AdminUsers api.Users
	err =  DecodeBody(res, &AdminUsers)
	if err != nil {
		return nil, err
	}
	return AdminUsers.GetUsers(), nil
}

func NewService(client *LastpassClient) (s *UserService) {
	return &UserService{client:client}
}

func (s *UserService) DoRequest() (*http.Response, error) {
	return s.client.DoRequest(s.command, s.data)
}