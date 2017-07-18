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

type Client struct {
	URL              *url.URL
	HttpClient       *http.Client
	CompanyId        string
	ProvisioningHash string
	Command          string
	Logger           *log.Logger
}

type User struct {
	UserName               string              `json:"username"`
	FullName               string              `json:"fullname,omitempty"`
	MasterPasswordStrength string              `json:"mpstrength,omitempty"`
	Created                string              `json:"created,omitempty"`
	LastPasswordChange     string              `json:"last_pw_change,omitempty"`
	LastLogin              string              `json:"lastlogin,omitempty"`
	Disabled               bool                `json:"disabled,omitempty"`
	NeverLoggedIn          bool                `json:"neverloggedin,omitempty"`
	LinkedAccount          string              `json:"linked,omitempty"`
	NumberOfSites          int                 `json:"sites,omitempty"`
	NumberOfNotes          int                 `json:"notes,omitempty"`
	NumberOfFormFills      int                 `json:"formfills,omitempty"`
	NumberOfApplications   int                 `json:"applications,omitempty"`
	NumberOfAttachments    int                 `json:"attachment,omitempty"`
	Groups                 []map[string]string `json:"groups,omitempty"`
}

const (
	CompanyID        = "8771312"
	ProvisioningHash = "359fdfbc93bc6b8f1963c84e9db3539a5f3d688f394bd536e1ca6b77f8d5f101"
	EndPointURL      = "https://lastpass.com/enterpriseapi.php"
)

func main() {
	//noinspection SpellCheckingInspection
	c, err := NewClient(EndPointURL, nil)

	if err != nil {
		fmt.Errorf(err.Error())
		return
	}

	user, err := c.GetUserData("suzuki.kengo@moneyforward.co.jp")
	if err != nil {
		fmt.Println(err)
		return
	}
	hoge := User{}
	decodeBody(user, hoge)
	fmt.Println(hoge)
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
		ProvisioningHash: ProvisioningHash,
		Logger:           logger,
	}, err
}

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

	fmt.Println(string(body))

	res, err := http.Post(EndPointURL, "application/json; charset=utf-8", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	c.Logger.Println(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type Request struct {
	CompanyID        string `json:"cid"`
	ProvisioningHash string `json:"provhash"`
	Command          string `json:"cmd"`
	Data             User	`json:"data"`
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(out)
}
