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

type DeactivationMode int

const (
	Deactivate DeactivationMode = iota
	Remove
	Delete
)

func (s *UserService) BatchAdd(users []api.User) error {
	s.command = "batchadd"
	s.data = users
	_, err := s.DoRequest()
	return err
}

// DeleteUser - delete individual users.
/*
0 - Deactivate user. This blocks logins but retains data and enterprise membership
1 - Remove user. This removed the user from the enterprise but otherwise keeps the account itself active.
2 - Delete user. This will delete the account entirely.
*/
func (s *UserService) DeleteUser(name string, mode DeactivationMode) error {
	s.command = "deleteaction"
	s.data = struct {
		UserName     string `json:"username"`
		DeleteAction int    `json:"deleteaction"`
	}{UserName: name, DeleteAction: int(mode)}
	_, err := s.DoRequest()
	return err
}

func (s *UserService) GetInactiveUser() ([]api.User, error) {
	s.command = "getuserdata"
	s.data = api.User{IsAdmin: false}
	res, err := s.DoRequest()
	if err != nil {
		return nil, err
	}

	var AdminUsers api.Users
	err = DecodeBody(res, &AdminUsers)
	if err != nil {
		return nil, err
	}
	return AdminUsers.GetInactiveUsers(), nil
}

func (s *UserService) GetDisabledUser() ([]api.User, error) {
	s.command = "getuserdata"
	s.data = api.User{Disabled: true}
	res, err := s.DoRequest()
	if err != nil {
		return nil, err
	}

	var Users api.Users
	err = DecodeBody(res, &Users)
	if err != nil {
		return nil, err
	}
	return Users.GetUsers(), nil
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
