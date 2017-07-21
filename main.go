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
	"os"
	"lastpass_provisioning/api"
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

type Status struct {
	Status string   `json:"status"`
	Errors []string `json:"errors,omitempty"`
}

type DeactivationMode int
const (
	Deactivate DeactivationMode = iota
	Remove
	Delete
)

type Config struct {
	CompanyId string `yaml:"company_id"`
	EndPoint  string `yaml:"end_point_url"`
	Secret 	  string `yaml:"secret"`
}

func formUsers(ou, parentOU *api.OU) map[string]*api.User {
	users := make(map[string]*api.User)

	// Construct Members within Child OU
	if parentOU != nil {
		ou.Name = fmt.Sprintf("%v - %v", parentOU.Name, ou.Name)
	}
	for _, member := range ou.Members {
		if _, ok := users[member]; ok {
			users[member].Groups = append(users[member].Groups, ou.Name)
		} else {
			users[member] = &api.User{UserName:member, Groups:[]string{ou.Name}}
		}
	}

	// Construct Members within Child OU
	for _, child_ou := range ou.Children {
		childUsers := formUsers(child_ou, ou)
		for user, child := range childUsers {
			if v, ok := users[user]; ok {
				v.Groups = append(v.Groups, child.Groups...)
			} else {
				users[user] = child
			}
		}
	}
	return users
}

func main() {

	// Client作成
	c, err := NewClient(nil)
	if err != nil {
		fmt.Errorf("Failed Building Client: %v", err.Error())
		os.Exit(1)
	}

	f, err := ioutil.ReadFile("organization_structure.yaml")
	if err != nil {
		panic(err)
	}
	var orgs struct{
		Organizations []*api.OU `yaml:",flow"`
	}

	err = yaml.Unmarshal(f, &orgs)
	if err != nil {
		panic(err)
	}

	users := make(map[string]*api.User)
	for _, ou := range orgs.Organizations {
		for user, hoge := range formUsers(ou, nil) {
			if v, ok := users[user]; ok {
				v.Groups = append(v.Groups, hoge.Groups...)
			} else {
				users[user] = hoge
			}
		}
	}

	hoge :=  make([]*api.User, len(users))
	fmt.Println(len(users))
	i := 0
	for _, u := range users {
		fmt.Println(u)
		hoge[i] = u
		i++
	}

	// Add Users
	//fuga := []*api.User{{UserName:"suzuki.kengo@moneyforward.co.jp"}}
	res, err := c.BatchAddOrUpdateUsers(hoge)
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
	fmt.Println(status.Errors)

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

func NewClient(logger *log.Logger) (*Client, error) {
	var config Config
	f, err := ioutil.ReadFile("secret.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		panic(err)
	}

	parsedURL, err := url.ParseRequestURI(config.EndPoint)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", config.EndPoint)
	}

	var discardLogger = log.New(ioutil.Discard, "", log.LstdFlags)
	if logger == nil {
		logger = discardLogger
	}

	return &Client{
		URL:              parsedURL,
		HttpClient:       http.DefaultClient,
		CompanyId:        config.CompanyId,
		ProvisioningHash: config.Secret,
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
    "provhash": "<Your API secret>",
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
func (c *Client) BatchAddOrUpdateUsers(users []*api.User) (*http.Response, error) {
	return c.DoRequest("batchadd", users)
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
func (c *Client) GetSharedFolderData() (*http.Response, error) {
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
func (c *Client) ChangeGroupsMembership(groups []api.BelongingGroup) (*http.Response, error) {
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
func (c *Client) GetUserData(user string) (*http.Response, error) {
	return c.DoRequest("getuserdata", api.User{UserName: user})
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
	return c.DoRequest("disablemultifactor", api.User{UserName: user})
}

// ResetPassword
func (c *Client) ResetPassword(user string) (*http.Response, error) {
	return c.DoRequest("resetpassword", api.User{UserName: user})
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

	res, err := http.Post(c.URL.String(), "application/json; charset=utf-8", bytes.NewBuffer(body))
	fmt.Println(c.URL.String())
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
