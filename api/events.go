package api

import (
	"strings"
	"encoding/json"
	"time"
)

type Events struct {
	Events []Event `json:"events"`
}

type Event struct {
	Time      string `json:"JsonTime"`
	Username  string `json:"Username,omitempty"`
	IPAddress string `json:"IP_Address,omitempty"`
	Action    string `json:"Action,omitempty"`
	Data      string `json:"Data,omitempty"`
	ID        string `json:"ID,omitempty"`
}

func (e *Event) UnmarshalJSON(b []byte) error {
	var rawStrings map[string]string

	err := json.Unmarshal(b, &rawStrings)
	if err != nil {
		return err
	}

	for k, v := range rawStrings {
		if strings.ToLower(k) == "time" {
			t, err := time.Parse("2006-01-02 15:04:05", v)
			if err != nil {
				return err
			}
			origLoc, _ := time.LoadLocation("EST")
			asiaLoc, _ := time.LoadLocation("Asia/Tokyo")
			e.Time = t.In(origLoc).In(asiaLoc).String()
		}
		if strings.ToLower(k) == "username" {
			e.Username = v
		}
		if k == "IP_Address" {
			e.IPAddress = v
		}
		if k == "Action" {
			e.Action = v
		}
		if k == "Data" {
			e.Data = v
		}
		if k == "ID" {
			e.ID = v
		}
	}

	return nil
}

func (e *Event) IsAuditEvent() bool {
	switch e.Action {
	case "Employee Account Deleted":
	case "Employee Account Created":
	case "従業員のアカウントを作成しました":
	case "Edit Policy":
	case "ポリシーの編集":
	case "Deactivated User":
	case "Reactivated User":
	case "Make Admin":
	case "Remove Admin":
	case "Master Password Reuse":
	case "Require Password Change":
	case "Super Admin Password Reset":
	case "Add to Shared Folder":
		if !strings.Contains(e.Data, "Shared-Super-Admins") {
			return false
		}
		break
	default:
		return false
	}
	return true
}
