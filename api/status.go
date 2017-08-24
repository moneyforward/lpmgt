package api

type ApiResultStatus struct {
Status   string   `json:"status,omitempty"`
}

func (s *ApiResultStatus)IsOK() bool {
	return s.Status == "OK"
}