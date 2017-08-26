package config

import (
	"io/ioutil"
	"os"
	"gopkg.in/yaml.v2"
)

// Config is config structure
type Config struct {
	EndPoint  string `yaml:"end_point_url"`
	Secret    string `yaml:"secret"` // API Key
	ConfFile  string
}

const (
	DefaultBaseURL   = "https://lastpass.com/enterpriseapi.php"
	DefaultUserAgent = "lastpass-client-go"
)

// LoadConfig loads config file in YAML format.
func LoadConfig(configFile string) (*Config, error) {
	config := &Config{}
	f, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		return nil, err
	}

	if config.EndPoint == "" {
		config.EndPoint = DefaultBaseURL
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