package client

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"lastpass_provisioning/util"
	"lastpass_provisioning/lastpass_config"
)

// Client is a general client structure.
type Client struct {
	URL       *url.URL
	APIKey    string
	Verbose   bool
	UserAgent string
	Logger    *log.Logger
	Headers   http.Header
}

//// LastpassClient is a Client that
//type LastpassClient struct {
//	Client    *Client
//	CompanyID string
//}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// NewClient returns a general Client structure.
func NewClient(apiKey string, endpointURL string, verbose bool) (*Client, error) {
	parsedURL, err := url.ParseRequestURI(endpointURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		URL:       parsedURL,
		APIKey:    apiKey,
		Verbose:   verbose,
		UserAgent: lastpass_config.DefaultUserAgent,
		Headers:   http.Header{},
		Logger:    nil,
	}, nil
}

// RequestJSON is a general http request in JSON form.
func (c *Client) RequestJSON(method string, path string, payload interface{}) (*http.Response, error) {
	body, err := util.JSONReader(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, c.URL.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	return c.ExecRequest(req)
}

// ExecRequest executes general http request
func (c *Client) ExecRequest(req *http.Request) (resp *http.Response, err error) {
	for header, values := range c.Headers {
		for _, v := range values {
			req.Header.Add(header, v)
		}
	}
	req.Header.Set("X-Api-Key", c.APIKey)
	req.Header.Set("User-Agent", c.UserAgent)

	if c.Verbose {
		dump, err := httputil.DumpRequest(req, true)
		if err == nil {
			log.Printf("%s", dump)
		}
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if c.Verbose {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Printf("%s", dump)
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp, fmt.Errorf("API result failed: %s", resp.Status)
	}
	return resp, nil
}

