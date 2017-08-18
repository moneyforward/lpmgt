package api

type SharedFolder struct {
	ShareFolderName string  `json:"sharedfoldername"`
	Score           float32 `json:"score"`
	Users           []User `json:"users"`
}
