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

type Status struct {
	S string `json:"status"`
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
	var hoge Status
	err = decodeBody(res, &hoge)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(hoge.S)

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

/*
  {
    "cid": "8771312",
    "provhash": "359fdfbc93bc6b8f1963c84e9db3539a5f3d688f394bd536e1ca6b77f8d5f101",
    "cmd": "getuserdata",
    "data": {
        "username": "user1@lastpass.com"
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
