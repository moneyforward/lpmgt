package api

import "strings"

type Event struct {
	Time       string `json:"Time"`
	Username   string `json:"Username"`
	IP_Address string `json:"IP_Address"`
	Action     string `json:"Action"`
	// {YYYY-MM-DD MM:DD:SS(US/Eastern time zone) USER IP ACTION}
	// {2017-07-25 09:40:56 suzuki.kengo@moneyforward.co.jp 210.138.23.111 Require Password Change kengo-admin@moneyforward.co.jp}
	Data       string `json:"Data"`
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
		if strings.Contains(e.Data, "Shared-Super-Admins") {
			return false
		}
	default:
		return false
	}
	return true
}