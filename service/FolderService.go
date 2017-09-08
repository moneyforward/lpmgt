package service

import (
	lc "lastpass_provisioning/lastpass_client"
	"net/http"
	"lastpass_provisioning/util"
	"fmt"
)

// SharedFolder is a LastPass Object in which users share accounts.
type SharedFolder struct {
	ShareFolderName string  `json:"sharedfoldername"`
	Score           float32 `json:"score"`
	Users           []User `json:"users"`
}

type folders struct {
	folders map[string]SharedFolder `json:"folders,omitempty"`
}

// FolderService is a service class that handles folder objects in LastPass.
type FolderService struct {
	client  *lc.LastPassClient
	command string
	data    interface{}
}

// NewEventService creates a new EventService
func NewFolderService(client *lc.LastPassClient) (s *FolderService) {
	return &FolderService{client: client}
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
func (s *FolderService) GetSharedFolders() (sf []SharedFolder, err error) {
	s.command = "getsfdata"
	s.data = SharedFolder{}
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var sharedFolders folders
	err = util.JSONBodyDecoder(res, &sharedFolders)
	fmt.Println(sharedFolders)
	if err != nil {
		return nil, err
	}
	return sharedFolders.GetFolders(), nil
}

func (s *FolderService) doRequest() (*http.Response, error) {
	return s.client.DoRequest(s.command, s.data)
}

func (fs *folders) GetFolders() []SharedFolder {
	folders := []SharedFolder{}
	for _, folder := range fs.folders {
		folders = append(folders, folder)
	}
	return folders
}