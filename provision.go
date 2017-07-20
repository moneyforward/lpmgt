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
	"time"
	"gopkg.in/yaml.v2"
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
	Readonly               string   `json:"readonly,omitempty"`       // ShareFolderの設定に利用. BooldでもなくIntでもない...
	Give                   string   `json:"give,omitempty"`           // ShareFolderの設定に利用
	Can_Administer         string   `json:"can_administer,omitempty"` // ShareFolderの設定に利用
}

type SharedFolder struct {
	ShareFolderName string  `json:"sharedfoldername"`
	Score           float32 `json:"score"`
	Users           []User  `json:"users"`
}

type BelongingGroup struct {
	Username   string   `json:"username"`
	GroupToAdd []string `json:"add,omitempty"`
	GroupToDel []string `json:"del,omitempty"`
}

type Event struct {
	Time       string `json:"Time"`
	Username   string `json:"Username"`
	IP_Address string `json:"IP_Address"`
	Action     string `json:"Action"`
	Data       string `json:"Data"`
}

type Status struct {
	Status string   `json:"status"`
	Errors []string `json:"errors,omitempty"`
}

type DeactivationMode int

const (
	CompanyID   = "8771312"
	Secret      = "359fdfbc93bc6b8f1963c84e9db3539a5f3d688f394bd536e1ca6b77f8d5f101"
	EndPointURL = "https://lastpass.com/enterpriseapi.php"
)
const (
	Deactivate DeactivationMode = iota
	Remove
	Delete
)

type OU struct {
	Name string
	Members []string `yaml:",flow"`
	Children	[]*OU
}

func formUser(email string, groups ...string) User {
	return User{UserName: email, Groups: groups}
}

func main() {
	f, err := ioutil.ReadFile("organization_structure.yaml")
	if err != nil {
		panic(err)
	}

	var ou struct{
		Organizations []*OU `yaml:",flow"`
	}

	err = yaml.Unmarshal(f, &ou)
	if err != nil {
		panic(err)
	}

	fmt.Println(ou.Organizations[1])

	//c, err := NewClient(EndPointURL, nil)
	//
	//if err != nil {
	//	fmt.Errorf(err.Error())
	//	return
	//}

	//var res *http.Response
	//// Get an User
	//res, err := c.GetUserData("suzuki.kengo@moneyforward.co.jp")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//var user struct {
	//	Users  map[string]User     `json:"Users,omitempty"`
	//	Groups map[string][]string `json:"Groups,omitempty"`
	//}
	//decodeBody(res, &user)
	//fmt.Println(user)
	//
	// Add Users
	//var users []User
	//emails := []string{
	//	"akiyama.yoshio@moneyforward.co.jp",
	//	"motokawa.daisuke@moneyforward.co.jp",
	//	"ueno.takeshi@moneyforward.co.jp",
	//	"kanebako.ryo@moneyforward.co.jp",
	//	"hanafusa.nobuhiro@moneyforward.co.jp",
	//
	//	"hanafusa.nobuhiro@moneyforward.co.jp",
	//	"kirihara.toyoaki@moneyforward.co.jp",
	//	"kimura.hisashi@moneyforward.co.jp",
	//
	//	"kanebako.ryo@moneyforward.co.jp",
	//	"yoshimoto.masaki@moneyforward.co.jp",
	//	"oyama.yasuhiro@moneyforward.co.jp",
	//
	//	"yamashita.manato@moneyforward.co.jp",
	//	"fukumoto.kazuhiro@moneyforward.co.jp",
	//	"sawada.tsuyoshi@moneyforward.co.jp",
	//	"sato.kimiaki@moneyforward.co.jp",
	//}
	//
	//for _, email := range emails {
	//	users = append(users, User{
	//		UserName: email,
	//		Groups: []string{"デザイン戦略室"},
	//	})
	//}
	//
	//res, err = c.BatchAddOrUpdateUsers(users)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//var status Status
	//err = decodeBody(res, &status)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(status.Status)
	//
	//// Get Shared Folder Data
	//res, err = c.GetSharedFolderData()
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//var sharedFolders map[string]SharedFolder
	//err = decodeBody(res, &sharedFolders)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//for _, sf := range sharedFolders {
	//	fmt.Println(sf)
	//}
	//
	//// BatchChange Group
	//data := BelongingGroup{
	//	Username:"suzuki.kengo@moneyforward.co.jp",
	//	GroupToAdd:[]string{"CISO室"},
	//}
	//data1 := BelongingGroup{
	//	Username:"kengoscal@gmail.com",
	//	GroupToAdd:[]string{"chalin-infra", "HOGEHOGE"},
	//	GroupToDel:[]string{"FUGAFUGA", "h"},
	//}
	//hoge := []BelongingGroup{data, data1} // hoge:=[...]Belonging{data}
	//res, err := c.ChangeGroupsMembership(hoge)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//var result Status
	//err = decodeBody(res, &result)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(result.Status + ": ")
	//fmt.Println(result.Errors)

	//res, err := c.DeleteUser("teramoto.tomoya@moneyforward.co.jp", Deactivate)
	//res, err := c.DisableMultifactor("teramoto.tomoya@moneyforward.co.jp")
	//res, err := c.ResetPassword("teramoto.tomoya@moneyforward.co.jp")
	//res, err := c.GetAllEventReports()
	//loc, _ := time.LoadLocation("Asia/Tokyo")
	//now := time.Now().In(loc)
	//weekAgo := now.Add(-time.Duration(7) * time.Hour * 24)
	//t := jsonLastPassTime{now}
	//f := jsonLastPassTime{weekAgo}
	//res, err := c.GetEventReport("allusers", "", f, t)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//var result struct {
	//	Events []Event `json:"events"`
	//}
	//err = decodeBody(res, &result)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(result.Events)
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
The "UserName" field is required while the "fullname", "groups", "duousername", "securidusername", "password" and "password_reset_required" fields are optional.
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
	return c.DoRequest("batchadd", users)
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
func (c *Client) GetSharedFolderData() (*http.Response, error) {
	return c.DoRequest("getsfdata", nil)
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
func (c *Client) ChangeGroupsMembership(groups []BelongingGroup) (*http.Response, error) {
	return c.DoRequest("batchchangegrp", groups)
}

/*
Get information on users enterprise.
# Request
  {
    "cid": "8771312",
    "provhash": "<Your API Secret>",
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
func (c *Client) GetUserData(user string) (*http.Response, error) {
	return c.DoRequest("getuserdata", User{UserName: user})
}

// DeleteUser - delete individual users.
/*
0 - Deactivate user. This blocks logins but retains data and enterprise membership
1 - Remove user. This removed the user from the enterprise but otherwise keeps the account itself active.
2 - Delete user. This will delete the account entirely.
*/
func (c *Client) DeleteUser(user string, mode DeactivationMode) (*http.Response, error) {
	data := struct {
		UserName     string `json:"username"`
		DeleteAction int    `json:"deleteaction"`
	}{UserName: user, DeleteAction: int(mode)}
	return c.DoRequest("deluser", data)
}

// DisableMultifactor
func (c *Client) DisableMultifactor(user string) (*http.Response, error) {
	return c.DoRequest("disablemultifactor", User{UserName: user})
}

// ResetPassword
func (c *Client) ResetPassword(user string) (*http.Response, error) {
	return c.DoRequest("resetpassword", User{UserName: user})
}

// GetEventReport
func (c *Client) GetEventReport(user, search string, from, to jsonLastPassTime) (*http.Response, error) {
	data := struct {
		From   jsonLastPassTime `json:"from"`
		To     jsonLastPassTime `json:"to"`
		Search string           `json:"search"`
		User   string           `json:"user"`
		Format string           `json:"format"`
	}{User: user, Search: search, From: from, To: to, Format: "siem"}
	return c.DoRequest("reporting", data)
}

// GetAllEventReports
func (c *Client) GetAllEventReports() (*http.Response, error) {
	data := struct {
		From   jsonLastPassTime `json:"from"`
		To     jsonLastPassTime `json:"to"`
		Search string           `json:"search"`
		User   string           `json:"user"`
		Format string           `json:"format"`
	}{User: "allusers", Format: "siem"}
	return c.DoRequest("reporting", data)
}

// DoRequest
func (c *Client) DoRequest(command string, data interface{}) (*http.Response, error) {
	req := Request{
		CompanyID:        c.CompanyId,
		ProvisioningHash: c.ProvisioningHash,
		Command:          command,
	}

	if data != nil {
		req.Data = data
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(body))

	res, err := http.Post(EndPointURL, "application/json; charset=utf-8", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return res, nil
}

// DecodeBody
func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

type jsonLastPassTime struct {
	time.Time
}

func (j jsonLastPassTime) format() string {
	return j.Time.Format("2006-01-02 15:04:05")
}

func (j jsonLastPassTime) MarshalJSON() ([]byte, error) {
	fmt.Println(j.format())
	return []byte(`"` + j.format() + `"`), nil
}
