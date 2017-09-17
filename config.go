package lpmgt

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
	TimeZone string  `yaml:"timezone,omitempty"`
}

const (
	defaultBaseURL   = "https://lastpass.com/enterpriseapi.php"
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

	if config.TimeZone == "" {
		config.TimeZone = "UTC"
	}

	return config, nil
}

// LoadEndPointURL returns endpoint url
func LoadEndPointURL(configFile string) string {
	config, err := LoadConfig(configFile)
	if err != nil {
		return ""
	}
	return config.EndPoint
}

// LoadAPIKeyFromEnvOrConfig returns API Key from either Env or LastPassConfig file
// If Env `LASTPASS_APIKEY` exists, that will be prioritized.
func LoadAPIKeyFromEnvOrConfig(configFile string) string {
	if secret := os.Getenv("LASTPASS_APIKEY"); secret != "" {
		return secret
	}
	config, err := LoadConfig(configFile)
	if err != nil {
		return ""
	}
	return config.Secret
}

// LoadCompanyIDFromEnvOrConfig returns Company ID provided by Lastpass.
func LoadCompanyIDFromEnvOrConfig(configFile string) string {
	if id := os.Getenv("LASTPASS_COMPANY_ID"); id != "" {
		return id
	}
	config, err := LoadConfig(configFile)
	if err != nil {
		return ""
	}
	return config.CompanyID
}
