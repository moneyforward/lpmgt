package lastpass_provisioning

import (
	"os"
	"errors"
	"lastpass_provisioning/logger"
	"net/http"
	"net/url"
	"log"
	"net/http/httputil"
	"fmt"
)

// LastPassClient is a Client that
type LastPassClient struct {
	URL       *url.URL
	APIKey    string
	Verbose   bool
	UserAgent string
	Logger    *log.Logger
	Headers   http.Header
	CompanyID string
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// NewClient returns a general Client structure.
func NewClient(apiKey, endpointURL, companyID string, verbose bool) (*LastPassClient, error) {
	parsedURL, err := url.ParseRequestURI(endpointURL)
	if err != nil {
		return nil, err
	}
	return &LastPassClient{
		URL:       parsedURL,
		APIKey:    apiKey,
		Verbose:   verbose,
		UserAgent: defaultUserAgent,
		Headers:   http.Header{},
		Logger:    nil,
		CompanyID: companyID,
	}, nil
}

// NewLastPassClient returns LastPass Client from confFile
func NewLastPassClient(configFilePath string) *LastPassClient {
	apiKey := LoadAPIKeyFromEnvOrConfig(configFilePath)
	companyID := LoadCompanyIDFromEnvOrConfig(configFilePath)
	endPointURL := LoadEndPointURL(configFilePath)

	if apiKey == "" {
		err := errors.New(`
    LASTPASS_APIKEY environment variable is not set. (Try "export LASTPASS_APIKEY='<Your apikey>'")
`)
		logger.DieIf(err)
	}

	if companyID == "" {
		err := errors.New(`
    LASTPASS_COMPANY_ID environment variable is not set. (Try "export LASTPASS_COMPANY_ID='<Your lastpass company id>'")
`)
		logger.DieIf(err)
	}

	if endPointURL == "" {
		endPointURL = defaultBaseURL
	}

	client, err := NewClient(apiKey, endPointURL, companyID, os.Getenv("DEBUG") != "")
	logger.DieIf(err)

	return client
}

// DoRequest executes LastPass specific request in JSON format and returns http Response
func (c *LastPassClient) DoRequest(command string, payload interface{}) (*http.Response, error) {
	data := struct {
		CompanyID        string      `json:"cid"`
		ProvisioningHash string      `json:"provhash"`
		Command          string      `json:"cmd"`
		Payload             interface{} `json:"payload"`
	}{
		CompanyID:        c.CompanyID,
		ProvisioningHash: c.APIKey,
		Command:          command,
	}

	if payload != nil {
		data.Payload = payload
	}

	// Form body.
	body, err := JSONReader(data)
	if err != nil {
		return nil, err
	}

	// Form Request.
	req, err := http.NewRequest(http.MethodPost, c.URL.String(), body)
	if err != nil {
		return nil, err
	}

	// Add Header.
	req.Header.Add("Content-Type", "application/json")
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

	// Do Request
	resp, err := http.DefaultClient.Do(req)
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

	return resp, err
}