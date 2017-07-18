package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
)

//https://lastpass.com/enterprise_apidoc.php

type Client struct {
	URL              *url.URL
	HttpClient       *http.Client
	CompanyId        string
	ProvisioningHash string
	Command          string
	Logger           *log.Logger
}

type Request struct {
	CompanyID        string      `json:"cid"`
	ProvisioningHash string      `json:"provhash"`
	Command          string      `json:"cmd"`
	Data             interface{} `json:"data"`
}

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
	NumberOfSites          int      `json:"sites,omitempty"`
	NumberOfNotes          int      `json:"notes,omitempty"`
	NumberOfFormFills      int      `json:"formfills,omitempty"`
	NumberOfApplications   int      `json:"applications,omitempty"`
	NumberOfAttachments    int      `json:"attachment,omitempty"`
	Groups                 []string `json:"groups,omitempty"`
	Readonly			   string		`json:"readonly,omitempty"`			// ShareFolderの設定に利用. BooldでもなくIntでもない...
	Give				   string		`json:"give,omitempty"`				// ShareFolderの設定に利用
	Can_Administer		   string		`json:"can_administer,omitempty"` 	// ShareFolderの設定に利用
}

type SharedFolder struct {
	ShareFolderName string `json:"sharedfoldername"`
	Score float32 `json:"score"`
	Users []User `json:"users"`
}

type BelongingGroup struct {
	Username   string `json:"username"`
	GroupToAdd []string `json:"add,omitempty"`
	GroupToDel []string `json:"del,omitempty"`
}

type Status struct {
	Status string `json:"status"`
	Errors []string `json:"errors,omitempty"`
}

const (
	CompanyID   = "8771312"
	Secret      = "359fdfbc93bc6b8f1963c84e9db3539a5f3d688f394bd536e1ca6b77f8d5f101"
	EndPointURL = "https://lastpass.com/enterpriseapi.php"
)

func main() {
	//noinspection SpellCheckingInspection
	c, err := NewClient(EndPointURL, nil)

	if err != nil {
		fmt.Errorf(err.Error())
		return
	}

	// Get an User
	res, err := c.GetUserData("suzuki.kengo@moneyforward.co.jp")
	if err != nil {
		fmt.Println(err)
		return
	}

	var user struct {
		Users  map[string]User     `json:"Users,omitempty"`
		Groups map[string][]string `json:"Groups,omitempty"`
	}
	decodeBody(res, &user)
	fmt.Println(user)

	// Add Users
	var users []User
	users = append(users, User{UserName: "kengoscal@gmail.com"})
	res, err = c.BatchAddOrUpdateUsers(users)
	if err != nil {
		fmt.Println(err)
		return
	}
	var status Status
	err = decodeBody(res, &status)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(status.Status)

	// Get Shared Folder Data
	res, err = c.GetSharedFolderData()
	if err != nil {
		fmt.Println(err)
		return
	}
	var sharedFolders map[string]SharedFolder
	err = decodeBody(res, &sharedFolders)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, sf := range sharedFolders {
		fmt.Println(sf)
	}

	// BatchChange Group
	data := BelongingGroup{
		Username:"suzuki.kengo@moneyforward.co.jp",
		GroupToAdd:[]string{"chalin-infra", "HOGEHOGE"},
		GroupToDel:[]string{"FUGAFUGA", "h"},
	}
	data1 := BelongingGroup{
		Username:"kengoscal@gmail.com",
		GroupToAdd:[]string{"chalin-infra", "HOGEHOGE"},
		GroupToDel:[]string{"FUGAFUGA", "h"},
	}
	hoge := []BelongingGroup{data, data1} // hoge:=[...]Belonging{data}
	res, err = c.ChangeGroupsMembership(hoge)
	if err != nil {
		fmt.Println(err)
		return
	}
	var result Status
	err = decodeBody(res, &result)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result.Status + ": ")
	fmt.Println(result.Errors)
}

func NewClient(urlString string, logger *log.Logger) (*Client, error) {
	parsedURL, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", urlString)
	}

	var discardLogger = log.New(ioutil.Discard, "", log.LstdFlags)
	if logger == nil {
		logger = discardLogger
	}

	return &Client{
		URL:              parsedURL,
		HttpClient:       http.DefaultClient,
		CompanyId:        CompanyID,
		ProvisioningHash: Secret,
		Logger:           logger,
	}, err
}

// Have not used
func (c *Client) newRequest(ctx context.Context, method, spath string, body io.Reader) (*http.Request, error) {
	u := *c.URL
	u.Path = path.Join(c.URL.Path, spath)
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	//req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("User-Agent", userAgent)
	return req, nil
}

/*
The "batchadd" command is used to provision new or update existing users.
The "username" field is required while the "fullname", "groups", "duousername", "securidusername", "password" and "password_reset_required" fields are optional.
By setting the "password" field you can define a default password for the new user that could be temporary or permanent based on the "password_reset_required" field's value (default: true).

# Request
  {
    "cid": "8771312",
    "provhash": "<Your API Secret>",
    "cmd": "batchadd",
    "data": [
        {
            "username": "user0@lastpass.com"
        },
        {
            "username": "user1@lastpass.com",
            "fullname": "John Doe"
        },
        {
            "username": "user2@lastpass.com",
            "fullname": "Mark Hunter",
            "groups": [
                "Group1",
                "Group2",
                "Group3"
            ],
            "attribs": {
                "objectGUID": "d3b07384d113edec49eaa6238ad5ff00",
                "Department": "Finance",
                "EmployeeNumber": "192832"
            }
        },
        {
            "username": "user3@lastpass.com",
            "fullname": "John Smith",
            "password": "DefaultPassword"
        },
        {
            "username": "user4@lastpass.com",
            "fullname": "Jane Smith",
            "password": "DefaultPassword",
            "password_reset_required": false
        }
    ]
}

# Response
{
   "status": "OK"
}
*/
func (c *Client) BatchAddOrUpdateUsers(users []User) (*http.Response, error) {
	body, err := json.Marshal(Request{
		CompanyID:        c.CompanyId,
		ProvisioningHash: c.ProvisioningHash,
		Command:          "batchadd",
		Data:             users,
	})
	if err != nil {
		return nil, err
	}

	res, err := http.Post(EndPointURL, "application/json; charset=utf-8", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return res, nil
}

/*
Get Shared Folder Data returns a JSON object containing information on all Shared Folders in the enterprise and the permissions granted to them.
# Request
{
	"cid": "8771312",
	"provhash": "<Your API Secret>",
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
func (c *Client) GetSharedFolderData()(*http.Response, error) {
	body, err := json.Marshal(Request{
		CompanyID:        c.CompanyId,
		ProvisioningHash: c.ProvisioningHash,
		Command:          "getsfdata",
	})
	if err != nil {
		return nil, err
	}

	res, err := http.Post(EndPointURL, "application/json; charset=utf-8", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return res, nil
}

/* Batch Change Group (cmd = batchchangegrp)
group membership manipulation
# Request

  {
    "cid": "8771312",
    "provhash": "<Your API Secret>",
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
func (c *Client) ChangeGroupsMembership(groups []BelongingGroup)(*http.Response, error) {
	body, err := json.Marshal(Request{
		CompanyID:        c.CompanyId,
		ProvisioningHash: c.ProvisioningHash,
		Command:          "batchchangegrp",
		Data:			groups,
	})
	if err != nil {
		return nil, err
	}

	res, err := http.Post(EndPointURL, "application/json; charset=utf-8", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return res, nil
}

/*
Get information on users enterprise.
# Request
  {
    "cid": "8771312",
    "provhash": "<Your API Secret>",
    "cmd": "getuserdata",
    "data": {
        "username": "user1@lastpass.com" // This can be either username, disabled, or admin
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
func (c *Client) GetUserData(user string) (*http.Response, error) {
	body, err := json.Marshal(Request{
		CompanyID:        c.CompanyId,
		ProvisioningHash: c.ProvisioningHash,
		Command:          "getuserdata",
		Data:             User{UserName: user},
	})
	if err != nil {
		return nil, err
	}

	res, err := http.Post(EndPointURL, "application/json; charset=utf-8", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}
