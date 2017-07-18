package main

import (
	"net/url"
	"net/http"
	"log"
	"github.com/pkg/errors"
	"io/ioutil"
	"fmt"
	"context"
	"io"
	"path"
	"encoding/json"
)

type Client struct {
	URL *url.URL
	HttpClient *http.Client
	Logger *log.Logger
}

var cid = "8771312"
var provhash = "359fdfbc93bc6b8f1963c84e9db3539a5f3d688f394bd536e1ca6b77f8d5f101"

func main() {
	//noinspection SpellCheckingInspection
	lastpass_url := "https://lastpass.com/enterpriseapi.php"
	c, err := NewClient(lastpass_url, nil)

	if err != nil {
		fmt.Errorf(err.Error())
		return
	}

	user, err := c.GetUserData("suzuki.kengo@moneyforward.co.jp")
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
	fmt.Println(user)
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

	return &Client{parsedURL, http.DefaultClient, logger}, err
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
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
func (c *Client) GetUserData(user string) (string, error) {
	c.URL.Path = path.Join(c.URL.Path, user)
	req, err:= http.NewRequest(http.MethodGet, c.URL.String(), nil)
	if err != nil {
		return "",err
	}

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return "",err
	}

	data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	return string(data), nil
}


func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(out)
}