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

func (s *UserService) GetAdminUserData() (api.Users, error) {
	var AdminUsers api.Users

	s.command = "getuserdata"
	s.data = api.User{IsAdmin: 1}
	res, err := s.DoRequest()
	if err != nil {
		return AdminUsers, err
	}

	if err :=  DecodeBody(res, &AdminUsers); err != nil {
		return AdminUsers, err
	}

	return AdminUsers, nil
}

func NewService(client *LastpassClient) (s *UserService) {
	return &UserService{client:client}
}

func (s *UserService) DoRequest() (*http.Response, error) {
	return s.client.DoRequest(s.command, s.data)
}