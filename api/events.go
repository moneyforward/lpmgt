package api

import (
	"strings"
)

type Events struct {
	Events []Event `json:"events"`
}

type Event struct {
	Time      string `json:"Time"`
	Username  string `json:"Username,omitempty"`
	IPAddress string `json:"IP_Address,omitempty"`
	Action    string `json:"Action,omitempty"`
	Data      string `json:"Data,omitempty"`
	ID        string `json:"ID,omitempty"`
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
