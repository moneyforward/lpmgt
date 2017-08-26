package api

import "lastpass_provisioning/service"

type SharedFolder struct {
	ShareFolderName string  `json:"sharedfoldername"`
	Score           float32 `json:"score"`
	Users           []service.User `json:"users"`
}
