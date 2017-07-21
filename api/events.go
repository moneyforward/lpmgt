package api

type Event struct {
	Time       string `json:"Time"`
	Username   string `json:"Username"`
	IP_Address string `json:"IP_Address"`
	Action     string `json:"Action"`
	Data       string `json:"Data"`
}