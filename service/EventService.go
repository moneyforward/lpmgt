package service

import (
	lc "lastpass_provisioning/lastpass_client"
	format "lastpass_provisioning/lastpass_format"
	"net/http"
	"strings"
	"time"
	"encoding/json"
	"lastpass_provisioning/util"
)
type Events struct {
	Events []Event `json:"events"`
}

// GetUserEvents get events from users
func (es *Events) GetUserEvents(username string) []Event {
	events := make([]Event, len(es.Events))
	for _, event := range es.Events {
		if username == event.Username {
			events = append(events, event)
		}
	}

	return events
}

type Event struct {
	Time      time.Time `json:"JsonTime"`
	Username  string    `json:"Username,omitempty"`
	IPAddress string    `json:"IP_Address,omitempty"`
	Action    string    `json:"Action,omitempty"`
	Data      string    `json:"Data,omitempty"`
	ID        string    `json:"ID,omitempty"`
}

func (e *Event) UnmarshalJSON(b []byte) error {
	var rawStrings map[string]string

	err := json.Unmarshal(b, &rawStrings)
	if err != nil {
		return err
	}

	for k, v := range rawStrings {
		switch strings.ToLower(k) {
		case "time":
			// LastPass's timestamp is in EST
			t, err := time.Parse("2006-01-02 15:04:05", v)
			if err != nil {
				return err
			}

			// Add 5 hours to convert time zone from EST to UTC
			// Then convert to asia/tokyo time zone
			asiaLoc, _ := time.LoadLocation("Asia/Tokyo")
			e.Time = t.Add(time.Duration(5) * time.Hour).In(asiaLoc)
		case "username":
			e.Username = v
		case "ip_address":
			e.IPAddress = v
		case "action":
			e.Action = v
		case "data":
			e.Data = v
		case "id":
			e.ID = v
		}
	}

	return nil
}

func (e Event) IsAuditEvent() bool {
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

type EventService struct {
	client  *lc.LastPassClient
	command string
	data    interface{}
}

// NewEventService creates a new EventService
func NewEventService(client *lc.LastPassClient) (s *EventService) {
	return &EventService{client: client}
}

func (s *EventService) doRequest() (*http.Response, error) {
	return s.client.DoRequest(s.command, s.data)
}

// GetEventReport fetches event of an user in certain period of time.
// Filtering is also available by setting search string.
func (s *EventService) GetEventReport(username, search string, from, to format.JsonLastPassTime) ([]Event, error) {
	s.command = "reporting"
	s.data = struct {
		From   format.JsonLastPassTime `json:"from"`
		To     format.JsonLastPassTime `json:"to"`
		Search string                         `json:"search"`
		User   string                         `json:"user"`
		Format string                         `json:"format"`
	}{User: username, From: from, To: to, Format: "siem"}

	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var events Events
	err = util.JSONBodyDecoder(res, &events)
	if err != nil {
		return nil, err
	}

	return events.Events, nil
}

// GetAllEventReports fetches event of all users in certain period of time.
// Filtering is also available by setting search string.
func (s *EventService) GetAllEventReports(from, to format.JsonLastPassTime) ([]Event, error) {
	s.GetEventReport("allusers", "", from, to)
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var events Events
	err = util.JSONBodyDecoder(res, &events)
	if err != nil {
		return nil, err
	}

	return events.Events, nil
}
