package main

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"lastpass_provisioning/api"
	"lastpass_provisioning/lastpass_time"
	"log"
	"net/http"
	"net/url"
)

//https://lastpass.com/enterprise_apidoc.php
type LastpassClient struct {
	URL              *url.URL
	HttpClient       *http.Client
	CompanyId        string
	ProvisioningHash string
	Logger           *log.Logger
}

type Request struct {
	CompanyID        string      `json:"cid"`
	ProvisioningHash string      `json:"provhash"`
	Command          string      `json:"cmd"`
	Data             interface{} `json:"data"`
}

func NewClient(logger *log.Logger) (*LastpassClient, error) {
	config := NewConfig()
	parsedURL, err := url.ParseRequestURI(config.EndPoint)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", config.EndPoint)
	}

	var discardLogger = log.New(ioutil.Discard, "", log.LstdFlags)
	if logger == nil {
		logger = discardLogger
	}

	return &LastpassClient{
		URL:              parsedURL,
		HttpClient:       http.DefaultClient,
		CompanyId:        config.CompanyId,
		ProvisioningHash: config.Secret,
		Logger:           logger,
	}, err
}

/*
Get Shared Folder Data returns a JSON object containing information on all Shared Folders in the enterprise and the permissions granted to them.
# Request
{
	"cid": "8771312",
	"provhash": "<Your API secret>",
    "cmd": "getsfdata"
}

# Response
{
    "101": {
        "sharedfoldername": "ThisSFName",
        "score": 99,
        "users": [
            {
                "username": "joe.user@lastpass.com",
                "readonly": 0,
                "give": 1,
                "can_administer": 1
            },
            {
                "username": "jane.user@lastpass.com",
                "readonly": 1,
                "give": 0,
                "can_administer": 0
            }
        ]
    }
}
*/
func (c *LastpassClient) GetSharedFolderData() (*http.Response, error) {
	return c.DoRequest("getsfdata", nil)
}

/* Batch Change Group (cmd = batchchangegrp)
group membership manipulation
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
func (c *LastpassClient) ChangeGroupsMembership(groups []api.BelongingGroup) (*http.Response, error) {
	return c.DoRequest("batchchangegrp", groups)
}

/*
Get information on users enterprise.
# Request
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
func (c *LastpassClient) GetUserData(user string) (*http.Response, error) {
	return c.DoRequest("getuserdata", api.User{UserName: user})
}

//// GetAdminUserData
//func (c *LastpassClient) GetAdminUserData() (*http.Response, error) {
//	return c.DoRequest("getuserdata", api.User{IsAdmin: 1})
//}

// DeleteUser - delete individual users.
/*
0 - Deactivate user. This blocks logins but retains data and enterprise membership
1 - Remove user. This removed the user from the enterprise but otherwise keeps the account itself active.
2 - Delete user. This will delete the account entirely.
*/
func (c *LastpassClient) DeleteUser(user string, mode DeactivationMode) (*http.Response, error) {
	data := struct {
		UserName     string `json:"username"`
		DeleteAction int    `json:"deleteaction"`
	}{UserName: user, DeleteAction: int(mode)}
	return c.DoRequest("deluser", data)
}

type DeactivationMode int

const (
	Deactivate DeactivationMode = iota
	Remove
	Delete
)

// DisableMultifactor
func (c *LastpassClient) DisableMultifactor(user string) (*http.Response, error) {
	return c.DoRequest("disablemultifactor", api.User{UserName: user})
}

// ResetPassword
func (c *LastpassClient) ResetPassword(user string) (*http.Response, error) {
	return c.DoRequest("resetpassword", api.User{UserName: user})
}

// GetEventReport
func (c *LastpassClient) GetEventReport(user, search string, from, to lastpass_time.JsonLastPassTime) (*http.Response, error) {
	data := struct {
		From   lastpass_time.JsonLastPassTime `json:"from"`
		To     lastpass_time.JsonLastPassTime `json:"to"`
		Search string                         `json:"search"`
		User   string                         `json:"user"`
		Format string                         `json:"format"`
	}{User: user, Search: search, From: from, To: to, Format: "siem"}
	return c.DoRequest("reporting", data)
}

// GetAllEventReports
func (c *LastpassClient) GetAllEventReports() (*http.Response, error) {
	data := struct {
		From   lastpass_time.JsonLastPassTime `json:"from"`
		To     lastpass_time.JsonLastPassTime `json:"to"`
		Search string                         `json:"search"`
		User   string                         `json:"user"`
		Format string                         `json:"format"`
	}{User: "allusers", Format: "siem"}
	return c.DoRequest("reporting", data)
}

// DoRequest
func (c *LastpassClient) DoRequest(command string, data interface{}) (*http.Response, error) {
	req := struct {
		CompanyID        string      `json:"cid"`
		ProvisioningHash string      `json:"provhash"`
		Command          string      `json:"cmd"`
		Data             interface{} `json:"data"`
	}{
		CompanyID:        c.CompanyId,
		ProvisioningHash: c.ProvisioningHash,
		Command:          command,
	}

	if data != nil {
		req.Data = data
	}

	r, err := JSONReader(req)
	if err != nil {
		return nil, err
	}

	return http.Post(c.URL.String(), "application/json; charset=utf-8", r)
}
