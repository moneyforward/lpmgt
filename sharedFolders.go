package lpmgt

import (
	"net/http"
)

// SharedFolder is a LastPass Object in which users share accounts.
type SharedFolder struct {
	ShareFolderName string  `json:"sharedfoldername"`
	Score           float32 `json:"score"`
	Users           []User `json:"users"`
}

// FolderService is a service class that handles folder objects in LastPass.
type FolderService struct {
	client  *LastPassClient
	command string
	data    interface{}
}

// NewFolderService creates a new NewFolderService
func NewFolderService(client *LastPassClient) (s *FolderService) {
	return &FolderService{client: client}
}

/*
GetSharedFolders returns a JSON object containing information on all Shared Folders in the enterprise and the permissions granted to them.
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
func (s *FolderService) GetSharedFolders() ([]SharedFolder, error) {
	s.command = "getsfdata"
	s.data = nil

	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var sharedFolders map[string]SharedFolder
	err = JSONBodyDecoder(res, &sharedFolders)
	if err != nil {
		return nil, err
	}

	sf := []SharedFolder{}
	for _, folder := range sharedFolders {
		sf = append(sf, folder)
	}
	return sf, nil
}

func (s *FolderService) doRequest() (*http.Response, error) {
	res, err := s.client.DoRequest(s.command, s.data)
	if err != nil {
		return nil, err
	}
	return res, nil
}