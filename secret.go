package main

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type Config struct {
	CompanyId string `yaml:"company_id"`
	EndPoint  string `yaml:"end_point_url"`
	Secret 	  string `yaml:"secret"`
}

func NewConfig() Config {
	var config Config
	f, err := ioutil.ReadFile("secret.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		panic(err)
	}

	return config
}