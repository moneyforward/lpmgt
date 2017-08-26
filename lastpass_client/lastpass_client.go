package lastpass_client

import (
	"lastpass_provisioning/lastpass_config"
	"os"
	"errors"
	"lastpass_provisioning/logger"
	"net/http"
	"lastpass_provisioning/client"
	"github.com/urfave/cli"
	lf "lastpass_provisioning/lastpass_format"
)

// LastPassClient is a Client that
type LastPassClient struct {
	Client    *client.Client
	CompanyID string
}

// NewLastPassClientFromContext creates LastpassClient.
// This method depends on urfave/cli.
func NewLastPassClientFromContext(c *cli.Context) *LastPassClient {
	confFile := c.GlobalString("config")
	apiKey := lastpass_config.LoadAPIKeyFromEnvOrConfig(confFile)
	companyID := lastpass_config.LoadCompanyIDFromEnvOrConfig(confFile)
	endPointURL := lastpass_config.LoadEndPointURL(confFile)

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

/*
GetSharedFolderData returns a JSON object containing information on all Shared Folders in the enterprise and the permissions granted to them.
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
func (c *LastPassClient) GetSharedFolderData() (*http.Response, error) {
	return c.DoRequest("getsfdata", nil)
}

// GetEventReport fetches event of an user in certain period of time.
// Filtering is also available by setting search string.
func (c *LastPassClient) GetEventReport(user, search string, from, to lf.JsonLastPassTime) (*http.Response, error) {
	data := struct {
		From   lf.JsonLastPassTime `json:"from"`
		To     lf.JsonLastPassTime `json:"to"`
		Search string                         `json:"search"`
		User   string                         `json:"user"`
		Format string                         `json:"format"`
	}{User: user, From: from, To: to, Format: "siem"}
	return c.DoRequest("reporting", data)
}

// GetAllEventReports fetches event of all users in certain period of time.
// Filtering is also available by setting search string.
func (c *LastPassClient) GetAllEventReports() (*http.Response, error) {
	data := struct {
		From   lf.JsonLastPassTime `json:"from"`
		To     lf.JsonLastPassTime `json:"to"`
		Search string                         `json:"search"`
		User   string                         `json:"user"`
		Format string                         `json:"format"`
	}{User: "allusers", Format: "siem"}
	return c.DoRequest("reporting", data)
}