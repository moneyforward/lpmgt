package lastpass_client

import (
	"lastpass_provisioning/lastpass_config"
	"os"
	"errors"
	"lastpass_provisioning/logger"
	"net/http"
	"lastpass_provisioning/client"
)

// LastPassClient is a Client that
type LastPassClient struct {
	Client    *client.Client
	CompanyID string
}

// NewLastPassClient returns LastPass Client from confFile
func NewLastPassClient(configFilePath string) *LastPassClient {
	apiKey := lastpass_config.LoadAPIKeyFromEnvOrConfig(configFilePath)
	companyID := lastpass_config.LoadCompanyIDFromEnvOrConfig(configFilePath)
	endPointURL := lastpass_config.LoadEndPointURL(configFilePath)

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
		endPointURL = lastpass_config.DefaultBaseURL
	}

	client, err := client.NewClient(apiKey, endPointURL, os.Getenv("DEBUG") != "")
	logger.DieIf(err)

	return &LastPassClient{
		Client:    client,
		CompanyID: companyID,
	}
}

// DoRequest executes LastPass specific request in JSON format and returns http Response
func (c *LastPassClient) DoRequest(command string, data interface{}) (*http.Response, error) {
	body := struct {
		CompanyID        string      `json:"cid"`
		ProvisioningHash string      `json:"provhash"`
		Command          string      `json:"cmd"`
		Data             interface{} `json:"data"`
	}{
		CompanyID:        c.CompanyID,
		ProvisioningHash: c.Client.APIKey,
		Command:          command,
	}

	if data != nil {
		body.Data = data
	}

	return c.Client.RequestJSON(http.MethodPost, "", body)
}