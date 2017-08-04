package api

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
	NumberOfSites          float64  `json:"sites,omitempty"`
	NumberOfNotes          float64  `json:"notes,omitempty"`
	NumberOfFormFills      float64  `json:"formfills,omitempty"`
	NumberOfApplications   float64  `json:"applications,omitempty"`
	NumberOfAttachments    float64  `json:"attachment,omitempty"`
	Groups                 []string `json:"groups,omitempty"`
	Readonly               string   `json:"readonly,omitempty"`       // ShareFolderの設定に利用. BooldでもなくIntでもない...
	Give                   string   `json:"give,omitempty"`           // ShareFolderの設定に利用
	Can_Administer         string   `json:"can_administer,omitempty"` // ShareFolderの設定に利用
	IsAdmin                bool     `json:"admin,omitempty"`
	Duousername            string   `json:"duousername,omitempty"`
	LastPwChange           string   `json:"last_pw_change,omitempty"`
	Mpstrength             string   `json:"mpstrength,omitempty"`
	Multifactor            string   `json:"multifactor,omitempty"`
}

type Users struct {
	Users   map[string]User     `json:"Users,omitempty"`
	Groups  map[string][]string `json:"Groups,omitempty"`
	Invited []string            `json:"invited,omitempty"`
}

func ConstructUser(email string, groupName... string) *User {
	u := &User{UserName: email}
	u.Groups = groupName
	return u
}

func (u *User) Contains(users []string) bool {
	for _, user := range users {
		if user == u.UserName {
			return true
		}
	}
	return false
}

func (us *Users) GetUsers() []User {
	var users []User
	for _, user := range us.Users {
		users = append(users, user)
	}
	return users
}
