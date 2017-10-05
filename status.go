package lpmgt

import (
	"github.com/pkg/errors"
)

// APIResultStatus is a status of response from LastPass API
// {"status":"FAIL","error":"No such user: masayoshi.umeda@moneyforward.co.jp"}
type APIResultStatus struct {
	Status   string   `json:"status,omitempty"`
	Errors string `json:"error,omitempty"`
}

// APIResultStatusForPasswordResetting has different format...
// {"status":"FAIL","error":["user not found: masayoshi.umeda@moneyforward.co.jp"]}
type APIResultStatusForPasswordResetting struct {
	Status   string   `json:"status,omitempty"`
	Errors []string `json:"error,omitempty"`
}

// IsOK checks status of response from LastPass
func (s *APIResultStatus)IsOK() bool {
	return s.Status == "OK"
}

func (s *APIResultStatus) Error() error {
	if s.IsOK() {
		return nil
	}
	b, e := IndentedJSON(s.Errors)
	if e != nil {
		return e
	}
	return errors.New(string(b))
}

func (s *APIResultStatus) String() string {
	return s.Status
}

func (s *APIResultStatusForPasswordResetting) String() string {
	return s.Status
}

func (s *APIResultStatusForPasswordResetting) Error() error {
	if s.Status == "OK" {
		return nil
	}
	b, e := IndentedJSON(s.Errors)
	if e != nil {
		return e
	}
	return errors.New(string(b))
}