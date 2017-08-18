package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Config struct {
	CompanyId string `yaml:"company_id"`
	EndPoint  string `yaml:"end_point_url"`
	Secret    string `yaml:"secret"` // API Key
	ConfFile  string
}

const (
	defaultBaseURL   = "https://lastpass.com/enterprise_apidoc.php"
	defaultUserAgent = "lastpass-client-go"
)

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
		config.EndPoint = defaultBaseURL
	}

	return config, nil
}

func (config *Config) LoadEndPointURL() string {
	return config.EndPoint
}

func (config *Config) LoadApiKeyFromEnvOrConfig() string {
	if secret := os.Getenv("LASTPASS_APIKEY"); secret != "" {
		return secret
	}
	return config.Secret
}

func (config *Config) LoadCompanyId() string {
	if id := os.Getenv("LASTPASS_COMPANY_ID"); id != "" {
		return id
	}
	return config.CompanyId
}
