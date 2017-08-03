package lastpassclient

import (
	"net/url"
	"net/http"
	"log"
	"encoding/json"
)

//https://lastpass.com/enterprise_apidoc.php
type client struct {
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

type Status struct {
	Status string   `json:"status"`
	Errors []string `json:"errors,omitempty"`
}


// DecodeBody
func DecodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}