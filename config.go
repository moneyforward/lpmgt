package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// LastPassConfig is config structure for LastPass
type LastPassConfig struct {
	CompanyID string `yaml:"company_id"`
	EndPoint  string `yaml:"end_point_url"`
	Secret    string `yaml:"secret"` // API Key
	ConfFile  string
}

const (
	defaultBaseURL   = "https://lastpass.com/enterprise_apidoc.php"
	defaultUserAgent = "lastpass-client-go"
)

// LoadConfig loads config file in YAML format.
func LoadConfig(configFile string) (*LastPassConfig, error) {
	config := &LastPassConfig{}
	f, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		return nil, err
	}

	if config.EndPoint == "" {
		config.EndPoint = defaultBaseURL
	}

	return config, nil
}

// LoadEndPointURL returns endpoint url
func (config *LastPassConfig) LoadEndPointURL() string {
	return config.EndPoint
}

// LoadAPIKeyFromEnvOrConfig returns API Key from either Env or LastPassConfig file
// If Env `LASTPASS_APIKEY` exists, that will be prioritized.
func (config *LastPassConfig) LoadAPIKeyFromEnvOrConfig() string {
	if secret := os.Getenv("LASTPASS_APIKEY"); secret != "" {
		return secret
	}
	return config.Secret
}

// LoadCompanyID returns Company ID provided by Lastpass.
func (config *LastPassConfig) LoadCompanyID() string {
	if id := os.Getenv("LASTPASS_COMPANY_ID"); id != "" {
		return id
	}
	return config.CompanyID
}
