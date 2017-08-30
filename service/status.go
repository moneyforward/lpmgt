package service

import (
	"github.com/pkg/errors"
	"lastpass_provisioning/util"
)

type ApiResultStatus struct {
	Status   string   `json:"status,omitempty"`
	Errors []string `json:"errors,omitempty"`
}

func (s *ApiResultStatus)IsOK() bool {
	return s.Status == "OK"
}

func (s *ApiResultStatus) Error() error {
	if s.IsOK() {
		return nil
	}
	b, e := util.IndentedJSON(s.Errors)
	if e != nil {
		return e
	}
	return errors.New(string(b))
}

func (s *ApiResultStatus) String() string {
	return s.Status
}