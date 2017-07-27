package api

import "net/http"

type User struct {
	UserName               string   `json:"username"`
	FullName               string   `json:"fullname,omitempty"`
	MasterPasswordStrength string   `json:"mpstrength,omitempty"`
	Created                string   `json:"created,omitempty"`
	LastPasswordChange     string   `json:"last_pw_change,omitempty"`
	LastLogin              string   `json:"lastlogin,omitempty"`
	Disabled               bool     `json:"disabled,omitempty"`
	NeverLoggedIn          bool     `json:"neverloggedin,omitempty"`
	LinkedAccount          string   `json:"linked,omitempty"`
	NumberOfSites          int      `json:"sites,omitempty"`
	NumberOfNotes          int      `json:"notes,omitempty"`
	NumberOfFormFills      int      `json:"formfills,omitempty"`
	NumberOfApplications   int      `json:"applications,omitempty"`
	NumberOfAttachments    int      `json:"attachment,omitempty"`
	Groups                 []string `json:"groups,omitempty"`
	Readonly               string   `json:"readonly,omitempty"`       // ShareFolderの設定に利用. BooldでもなくIntでもない...
	Give                   string   `json:"give,omitempty"`           // ShareFolderの設定に利用
	Can_Administer         string   `json:"can_administer,omitempty"` // ShareFolderの設定に利用
	IsAdmin				   int      `json:"admin,omitempty"`
}

type UserService struct {
	Client *http.Client
	Command string
}

func (u *User) Contains(users []string) bool {
	for _, user := range users {
		if user == u.UserName {
			return true
		}
	}
	return false
}

func (s *UserService) DoRequest() (*http.Response, error) {

}

