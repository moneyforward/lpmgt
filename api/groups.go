package api

type OU struct {
	Name string
	Members []string `yaml:",flow"`
	Children	[]*OU
}

type BelongingGroup struct {
	Username   string   `json:"username"`
	GroupToAdd []string `json:"add,omitempty"`
	GroupToDel []string `json:"del,omitempty"`
}