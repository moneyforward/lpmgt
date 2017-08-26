package service

import (
	"fmt"
	"github.com/pkg/errors"
	"lastpass_provisioning/api"
	"net/http"
	"lastpass_provisioning/util"
	"lastpass_provisioning/lastpass_client"
)

// ToDO
/* reinviteuser -> status.go
disablemultifactor ->
 */
// UserService is a service class that sends a request to LastPass provisioning API.
type UserService struct {
	client  *lastpass_client.LastPassClient
	command string
	data    interface{}
}

// DeactivationMode is enum which deactivate/delete users
type DeactivationMode int

const (
	// Deactivate deactivates user
	Deactivate DeactivationMode = iota
	// Remove removes user from Org
	Remove
	// Delete deletes user account (completely)
	Delete
)

// GetUserData gets information on users enterprise.
/*Request
  {
    "cid": "8771312",
    "provhash": "<Your API secret>",
    "cmd": "getuserdata",
    "data": {
        "username": "user1@lastpass.com" // This can be either UserName, disabled, or admin
    }
  }

# Response
  {
    "Users": {
        "101": {
            "username": "user1@lastpass.com",
            "fullname": "Ned Flanders",
            "mpstrength": "100",
            "created": "2014-03-12 10:02:56",
            "last_pw_change": "2015-05-19 10:58:33",
            "last_login": "2015-05-29 11:45:05",
            "disabled": false,
            "neverloggedin": false,
            "linked": "personal.account@mydomain.com",
            "sites": 72,
            "notes": 19,
            "formfills": 2,
            "applications": 0,
            "attachments": 1,
            "groups": [
                "Domain Admins",
                "Dev Team",
                "Support Team"
            ]
        }
    },
    "Groups": {
        "Domain Admins": [
            "user1@lastpass.com"
        ],
        "Dev Team": [
            "user1@lastpass.com"
        ],
        "Support Team": [
            "user1@lastpass.com"
        ]
    }
}
*/
// GetUserData
func (s *UserService) GetUserData(userName string) (user api.User, err error) {
	s.command = "getuserdata"
	s.data = api.User{UserName: userName}
	res, err := s.doRequest()

	if err != nil {
		return
	}

	users := &api.Users{}
	err = util.JSONBodyDecoder(res, users)
	if err != nil {
		return
	}
	if len(users.GetUsers()) != 0 {
		user = users.GetUsers()[0]
	} else {
		eMessage := fmt.Sprintf("User %v does not exist", userName)
		return user, errors.New(eMessage)
	}
	return
}

// BatchAdd - add users.
func (s *UserService) BatchAdd(users []api.User) error {
	s.command = "batchadd"
	s.data = users
	_, err := s.doRequest()
	return err
}

// UpdateUser updates user's info.
func (s *UserService) UpdateUser(user api.User) error {
	s.command = "batchadd"
	s.data = user
	res, err := s.doRequest()
	status := &api.ApiResultStatus{}
	err = util.JSONBodyDecoder(res, status)
	if err != nil {
		return err
	}
	if !status.IsOK() {
		return errors.New(status.Status)
	}
	return nil
}

// DeleteUser - delete individual users.
/*
0 - Deactivate user. This blocks logins but retains data and enterprise membership
1 - Remove user. This removed the user from the enterprise but otherwise keeps the account itself active.
2 - Delete user. This will delete the account entirely.
*/
func (s *UserService) DeleteUser(name string, mode DeactivationMode) error {
	s.command = "deluser"
	s.data = struct {
		UserName     string `json:"username"`
		DeleteAction int    `json:"deleteaction"`
	}{UserName: name, DeleteAction: int(mode)}
	_, err := s.doRequest()
	return err
}

// GetNon2faUsers retrieves users without 2 factor authentication setting.
func (s *UserService) GetNon2faUsers() ([]api.User, error) {
	s.command = "getuserdata"
	s.data = api.User{}
	res, err := s.doRequest()
	var users api.Users
	err = util.JSONBodyDecoder(res, &users)
	if err != nil {
		return nil, err
	}
	return users.GetNon2faUsers(), nil
}

// GetAllUsers simply retrieves all users
func (s *UserService) GetAllUsers() ([]api.User, error) {
	s.command = "getuserdata"
	s.data = api.User{}
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}
	var nonAdminUsers api.Users
	err = util.JSONBodyDecoder(res, &nonAdminUsers)
	if err != nil {
		return nil, err
	}
	return nonAdminUsers.GetUsers(), nil
}

// GetInactiveUsers is Deactivated user(Deleted user in mode 0)
func (s *UserService) GetInactiveUsers() ([]api.User, error) {
	s.command = "getuserdata"
	s.data = api.User{}
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var nonAdminUsers api.Users
	err = util.JSONBodyDecoder(res, &nonAdminUsers)
	if err != nil {
		return nil, err
	}
	return nonAdminUsers.GetNeverLoggedInUsers(), nil
}

// GetDisabledUsers gets Deactivated user(Deleted user in mode 0)
func (s *UserService) GetDisabledUsers() ([]api.User, error) {
	s.command = "getuserdata"
	s.data = api.User{Disabled: true}
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var Users api.Users
	err = util.JSONBodyDecoder(res, &Users)
	if err != nil {
		return nil, err
	}
	return Users.GetUsers(), nil
}

// GetAdminUserData gets admin users
func (s *UserService) GetAdminUserData() ([]api.User, error) {
	s.command = "getuserdata"
	s.data = api.User{IsAdmin: true}
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var AdminUsers api.Users
	err = util.JSONBodyDecoder(res, &AdminUsers)
	if err != nil {
		return nil, err
	}
	return AdminUsers.GetUsers(), nil
}

// ChangeGroupsMembership changes Group in batch(cmd = batchchangegrp)
/*
# Request

  {
    "cid": "8771312",
    "provhash": "<Your API secret>",
    "cmd": "batchchangegrp",
    "data": [
        {
            "username": "user1@lastpass.com",
            "add": [
                "Group1",
                "Group2"
            ]
        },
        {
            "username": "user2@lastpass.com",
            "add": [
                "Group1"
            ],
            "del": [
                "Group2",
                "Group3"
            ]
        }
    ]
}

# Response
{
    "status": "WARN", // OK, WARN or FAIL
    "errors": [
        "user2@lastpass.com does not exist"
    ]
}
*/
//func (s *UserService) ChangeGroupsMembership(groups []api.TransferringUser) (*http.Response, error) {
//	return .DoRequest("batchchangegrp", groups)
//}

// NewUserService creates a new UserService
func NewUserService(client *lastpass_client.LastPassClient) (s *UserService) {
	return &UserService{client: client}
}

func (s *UserService) doRequest() (*http.Response, error) {
	return s.client.DoRequest(s.command, s.data)
}
