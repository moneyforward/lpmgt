package api

type BelongingGroup struct {
	Username   string   `json:"username"`
	GroupToAdd []string `json:"add,omitempty"`
	GroupToDel []string `json:"del,omitempty"`
}
