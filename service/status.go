package service

import (
	"github.com/pkg/errors"
	lp "lastpass_provisioning"
)

// APIResultStatus is a status of response from LastPass API
type APIResultStatus struct {
	Status   string   `json:"status,omitempty"`
	Errors []string `json:"errors,omitempty"`
}

// IsOK checks status of response from LastPass
func (s *APIResultStatus)IsOK() bool {
	return s.Status == "OK"
}

func (s *APIResultStatus) Error() error {
	if s.IsOK() {
		return nil
	}
	b, e := lp.IndentedJSON(s.Errors)
	if e != nil {
		return e
	}
	return errors.New(string(b))
}

func (s *APIResultStatus) String() string {
	return s.Status
}