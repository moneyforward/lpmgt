package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
	"gopkg.in/yaml.v2"
	"lastpass_provisioning/api"
	"os"
	lastpassTime "lastpass_provisioning/lastpass_time"
	"lastpass_provisioning/client"
	"sync"
)

//https://lastpass.com/enterprise_apidoc.php
type Client struct {
	URL              *url.URL
	HttpClient       *http.Client
	CompanyId        string
	ProvisioningHash string
	Logger           *log.Logger
}

type Status struct {
	Status string   `json:"status"`
	Errors []string `json:"errors,omitempty"`
}

type Config struct {
	CompanyId string `yaml:"company_id"`
	EndPoint  string `yaml:"end_point_url"`
	Secret 	  string `yaml:"secret"`
}

func main() {

	// Client作成
	c, err := NewClient(nil)
	if err != nil {
		fmt.Errorf("Failed Building Client: %v", err.Error())
		os.Exit(1)
	}

	fmt.Println(" --------------------------------------  Admin Users ---------------------------------------- ")
	res, err := c.GetAdminUserData()
	if err != nil {
		fmt.Println(err)
		return
	}

	var AdminUsers api.Users
	err = client.DecodeBody(res, &AdminUsers)
	i := 0
	for _, admin := range AdminUsers.Users {
		fmt.Println(admin.UserName)
		i++
	}


	// Get Shared Folder Data
	fmt.Println(" --------------------------------------  Super Shared Folders ------------------------------- ")
	res, err = c.GetSharedFolderData()
	if err != nil {
		fmt.Println(err)
		return
	}
	var sharedFolders map[string]api.SharedFolder
	err = client.DecodeBody(res, &sharedFolders)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, sf := range sharedFolders {
		if sf.ShareFolderName == "Super-Admins" {
			for _, user := range sf.Users {
				fmt.Println(sf.ShareFolderName + " : " + user.UserName)
			}
		}
	}

	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)
	dayAgo := now.Add(-time.Duration(1) * time.Hour * 24)
	t := lastpassTime.JsonLastPassTime{now}
	f := lastpassTime.JsonLastPassTime{dayAgo}
	res, err = c.GetEventReport("", "", f, t)
	if err != nil {
		fmt.Println(err)
		return
	}

	var result api.Events
	err = client.DecodeBody(res, &result)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(fmt.Sprintf(" -------------------------------- Events(%v ~ %v) -----------------------------------", f.Format(), t.Format()))
	for _, event := range result.Events {
		if event.IsAuditEvent() {
			fmt.Println(event)
		}
	}


	//q := make(chan string)

	hoge :=  func(q chan string) {
		for {
			userName, ok := <- q
			if !ok {
				return
			}
			res, err = c.GetEventReport(userName, "login", f, t)
			fmt.Println(fmt.Sprintf(" --------------------------------------  %v Login History ------------------------------- ", userName))
			if err != nil {
				fmt.Println(err)
				return
			}
			err = client.DecodeBody(res, &result)
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, event := range result.Events {
				fmt.Println(event)
			}
		}
		//fmt.Println(fmt.Sprintf(" --------------------------------------  %v Login ------------------------------- ", userName))
		//res, err = c.GetEventReport(userName, "login", f, t)
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
		//
		//err = client.DecodeBody(res, &result)
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
		//for _, event := range result.Events {
		//	//q <- event.Data
		//	fmt.Println(event.Data)
		//}
	}

	var wg sync.WaitGroup
	q := make(chan string, 5)
	for i:= 0; i < len(AdminUsers.Users); i++ {
		wg.Add(1)
		go hoge(q)
	}

	for _, admin := range AdminUsers.Users {
		q <- admin.UserName
	}
	close(q)
	wg.Wait()

	// Delete
	//_, err = c.DeleteUser("takizawa.naoto@moneyforward.co.jp", Delete)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//_, err = c.DeleteUser("suga.kosuke@moneyforward.co.jp", Delete)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

	// Move
	//_, err = c.ChangeGroupsMembership([]api.BelongingGroup{
	//	{
	//		"ikeuchi.kenichi@moneyforward.co.jp",
	//		[]string{"PFMサービス開発本部"},
	//		[]string{},
	//	},
	//})
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

	// Add
	//_, err = c.BatchAddOrUpdateUsers(
	//	[]*api.User{
	//		{UserName:"takahashi.yuto@moneyforward.co.jp",Groups:[]string{"MFクラウドサービス開発本部"}},
	//		{UserName:"ishii.hiroyuki@moneyforward.co.jp",Groups:[]string{"MFクラウドサービス開発本部"}},
	//		{UserName:"suzuki.shota.340@moneyforward.co.jp",Groups:[]string{"PFMサービス開発本部"}},
	//		{UserName:"oba.akitaka@moneyforward.co.jp",Groups:[]string{"PFMサービス開発本部"}},
	//		{UserName:"ono.yumemi@moneyforward.co.jp",Groups:[]string{"アカウントアグリゲーション本部"}},
	//		{UserName:"takenaka.kazumasa@moneyforward.co.jp",Groups:[]string{"MFクラウド事業推進本部 - 事業戦略部"}},
	//		{UserName:"ukon.yuto@@moneyforward.co.jp", Groups:[]string{"MFクラウド事業推進本部 - ダイレクトセールス部"}},
	//		{UserName:"lee.choonghaeng@moneyforward.co.jp",Groups: []string{"MFクラウド事業推進本部 - MFクラウド事業戦略部"}},
	//		{UserName:"furuhama.yusuke@moneyforward.co.jp", Groups: []string{"社長室 - Chalin", "MFクラウドサービス開発本部"}},
	//	},
	//)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
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

func (c *Client) InitialBatchAdd() (*http.Response, error) {
	org := api.ReadOrg("organization_structure.yaml")
	userMap := make(map[string]*api.User)
	for _, ou := range org.OUs {
		for userName, user := range api.FormUsers(ou, nil) {
			if v, ok := userMap[userName]; ok {
				v.Groups = append(v.Groups, user.Groups...)
			} else {
				userMap[userName] = user
			}
		}
	}

	i := 0
	users := make([]*api.User, len(userMap))
	for _, u := range userMap {
		users[i] = u
		i++
	}

	return c.BatchAddOrUpdateUsers(users)
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

// GetAdminUserData
func (c *Client) GetAdminUserData() (*http.Response, error) {
	return c.DoRequest("getuserdata", api.User{IsAdmin:1})
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
type DeactivationMode int
const (
	Deactivate DeactivationMode = iota
	Remove
	Delete
)

// DisableMultifactor
func (c *Client) DisableMultifactor(user string) (*http.Response, error) {
	return c.DoRequest("disablemultifactor", api.User{UserName: user})
}

// ResetPassword
func (c *Client) ResetPassword(user string) (*http.Response, error) {
	return c.DoRequest("resetpassword", api.User{UserName: user})
}

// GetEventReport
func (c *Client) GetEventReport(user, search string, from, to lastpassTime.JsonLastPassTime) (*http.Response, error) {
	data := struct {
		From   lastpassTime.JsonLastPassTime `json:"from"`
		To     lastpassTime.JsonLastPassTime `json:"to"`
		Search string           `json:"search"`
		User   string           `json:"user"`
		Format string           `json:"format"`
	}{User: user, Search: search, From: from, To: to, Format: "siem"}
	return c.DoRequest("reporting", data)
}

// GetAllEventReports
func (c *Client) GetAllEventReports() (*http.Response, error) {
	data := struct {
		From   lastpassTime.JsonLastPassTime `json:"from"`
		To     lastpassTime.JsonLastPassTime `json:"to"`
		Search string           `json:"search"`
		User   string           `json:"user"`
		Format string           `json:"format"`
	}{User: "allusers", Format: "siem"}
	return c.DoRequest("reporting", data)
}

// DoRequest
func (c *Client) DoRequest(command string, data interface{}) (*http.Response, error) {
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

func JSONReader(v interface{}) (io.Reader, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(v)

	if err != nil {
		return nil, err
	}
	return buf, nil
}