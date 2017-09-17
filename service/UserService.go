package service

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	lp "lastpass_provisioning"
)

// UserService is a service class that sends a request to LastPass provisioning API.
type UserService struct {
	client  *lp.LastPassClient
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
func (s *UserService) GetUserData(userName string) (user User, err error) {
	s.command = "getuserdata"
	s.data = User{UserName: userName}
	res, err := s.doRequest()

	if err != nil {
		return
	}

	users := &Users{}
	err = lp.JSONBodyDecoder(res, users)
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
func (s *UserService) BatchAdd(users []User) error {
	s.command = "batchadd"
	s.data = users
	res, err := s.doRequest()
	status := &APIResultStatus{}
	err = lp.JSONBodyDecoder(res, status)
	if err != nil {
		return err
	}
	 return status.Error()
}

// UpdateUser updates user's info.
func (s *UserService) UpdateUser(user User) error {
	s.command = "batchadd"
	s.data = user
	res, err := s.doRequest()
	status := &APIResultStatus{}
	err = lp.JSONBodyDecoder(res, status)
	if err != nil {
		return err
	}
	return status.Error()
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
	res, err := s.doRequest()
	status := &APIResultStatus{}
	err = lp.JSONBodyDecoder(res, status)
	if err != nil {
		return err
	}
	return status.Error()
}

// GetNon2faUsers retrieves users without 2 factor authentication setting.
func (s *UserService) GetNon2faUsers() ([]User, error) {
	s.command = "getuserdata"
	s.data = User{}
	res, err := s.doRequest()
	var users Users
	err = lp.JSONBodyDecoder(res, &users)
	if err != nil {
		return nil, err
	}
	return users.GetNon2faUsers(), nil
}

// GetAllUsers simply retrieves all users
func (s *UserService) GetAllUsers() ([]User, error) {
	s.command = "getuserdata"
	s.data = User{}
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}
	var nonAdminUsers Users
	err = lp.JSONBodyDecoder(res, &nonAdminUsers)
	if err != nil {
		return nil, err
	}
	return nonAdminUsers.GetUsers(), nil
}

// GetInactiveUsers is Deactivated user(Deleted user in mode 0)
func (s *UserService) GetInactiveUsers() ([]User, error) {
	s.command = "getuserdata"
	s.data = User{}
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var nonAdminUsers Users
	err = lp.JSONBodyDecoder(res, &nonAdminUsers)
	if err != nil {
		return nil, err
	}
	return nonAdminUsers.GetNeverLoggedInUsers(), nil
}

// GetDisabledUsers gets Deactivated user(Deleted user in mode 0)
func (s *UserService) GetDisabledUsers() ([]User, error) {
	s.command = "getuserdata"
	s.data = User{Disabled: true}
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var Users Users
	err = lp.JSONBodyDecoder(res, &Users)
	if err != nil {
		return nil, err
	}
	return Users.GetUsers(), nil
}

// GetAdminUserData gets admin users
func (s *UserService) GetAdminUserData() ([]User, error) {
	s.command = "getuserdata"
	s.data = User{IsAdmin: true}
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var adminUsers Users
	err = lp.JSONBodyDecoder(res, &adminUsers)
	if err != nil {
		return nil, err
	}
	return adminUsers.GetUsers(), nil
}

// DisableMultifactor disables multifactor setting of user
func (s *UserService) DisableMultifactor(username string) (*APIResultStatus, error) {
	s.command = "disablemultifactor"
	s.data = User{UserName:username}
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var status APIResultStatus
	err = lp.JSONBodyDecoder(res, &status)
	if err != nil {
		return nil, err
	}
	return &status, status.Error()
}

// ResetPassword reset password for the user
func (s *UserService) ResetPassword(username string) (*APIResultStatus, error) {
	s.command = "resetpassword"
	s.data = User{UserName:username}
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var status APIResultStatus
	err = lp.JSONBodyDecoder(res, &status)
	if err != nil {
		return nil, err
	}
	return &status, status.Error()
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
//func (s *UserService) ChangeGroupsMembership(groups []TransferringUser) (*http.Response, error) {
//	return .DoRequest("batchchangegrp", groups)
//}

// NewUserService creates a new UserService
func NewUserService(client *lp.LastPassClient) (s *UserService) {
	return &UserService{client: client}
}

func (s *UserService) doRequest() (*http.Response, error) {
	return s.client.DoRequest(s.command, s.data)
}

// User is a structure
type User struct {
	UserName               string   `json:"username"`
	FullName               string   `json:"fullname,omitempty"`
	MasterPasswordStrength string   `json:"mpstrength,omitempty"`
	Created                string   `json:"created,omitempty"`
	LastPasswordChange     string   `json:"last_pw_change,omitempty"`
	LastLogin              string   `json:"lastlogin,omitempty"`
	Disabled               bool     `json:"disabled,omitempty"`
	NeverLoggedIn          bool     `json:"neverloggedin,omitempty"`
	LinkedAccount          string   `json:"linked,omitempty"`
	NumberOfSites          float64  `json:"sites,omitempty"`
	NumberOfNotes          float64  `json:"notes,omitempty"`
	NumberOfFormFills      float64  `json:"formfills,omitempty"`
	NumberOfApplications   float64  `json:"applications,omitempty"`
	NumberOfAttachments    float64  `json:"attachment,omitempty"`
	Groups                 []string `json:"groups,omitempty"`
	Readonly               string   `json:"readonly,omitempty"`       // ShareFolderの設定に利用. BoolでもなくIntでもない...
	Give                   string   `json:"give,omitempty"`           // ShareFolderの設定に利用
	Can_Administer         string   `json:"can_administer,omitempty"` // ShareFolderの設定に利用
	IsAdmin                bool     `json:"admin,omitempty"`
	Duousername            string   `json:"duousername,omitempty"`
	LastPwChange           string   `json:"last_pw_change,omitempty"`
	Mpstrength             string   `json:"mpstrength,omitempty"`
	Multifactor            string   `json:"multifactor,omitempty"`
}

type Users struct {
	Users   map[string]User     `json:"Users,omitempty"`
	Groups  map[string][]string `json:"Groups,omitempty"`
	Invited []string            `json:"invited,omitempty"`
}

type TransferringUser struct {
	Username   string   `json:"username"`
	JoiningDepName []string `json:"add,omitempty"`
	LeavingDepName []string `json:"del,omitempty"`
}

func (u *User) Contains(users []string) bool {
	for _, user := range users {
		if user == u.UserName {
			return true
		}
	}
	return false
}

func (u *User) Is2FA() bool {
	return u.Multifactor!=""
}

func (us *Users) GetUsers() []User {
	users := []User{}
	for _, user := range us.Users {
		users = append(users, user)
	}
	return users
}

func (us *Users) GetNon2faUsers() (users []User) {
	for _, user := range us.Users {
		if user.Multifactor == "" {
			users = append(users, user)
		}
	}
	return users
}

func (us *Users) GetNeverLoggedInUsers() (users []User) {
	for _, user := range us.Users{
		if user.NeverLoggedIn {
			users = append(users, user)
		}
	}
	return
}

func PrintUserNames(users []User) {
	for _, user := range users {
		fmt.Println(user.UserName)
	}
}