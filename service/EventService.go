package service

import (
	lc "lastpass_provisioning/lastpass_client"
	lp "lastpass_provisioning"
	"net/http"
	"strings"
	"time"
	"encoding/json"
	"github.com/pkg/errors"
)

type Events struct {
	Events []Event `json:"events"`
}

// GetUserEvents get events from users
func (es *Events) GetUserEvents(username string) *Events {
	events := make([]Event, 0)
	for _, event := range es.Events {
		if username == event.Username {
			events = append(events, event)
		}
	}

	return &Events{Events:events}
}

func (es *Events) ConvertTimezone(timezone *time.Location) {
	for index, event := range es.Events {
		es.Events[index].Time = event.Time.UTC().In(timezone)
	}
}

type Event struct {
	Time      time.Time `json:"JsonTime"`
	Username  string    `json:"Username,omitempty"`
	IPAddress string    `json:"IP_Address,omitempty"`
	Action    string    `json:"Action,omitempty"`
	Data      string    `json:"Data,omitempty"`
	ID        string    `json:"ID,omitempty"`
}

func (e *Event) String(timezone *time.Location) string {
	return e.Time.UTC().In(timezone).String() + " " + e.Username + " " + e.IPAddress + " " + e.Action + " " + e.Data
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
			eastLoc, err := time.LoadLocation(lp.LastPassTimeZone)
			if err != nil {
				return err
			}
			t, err := time.ParseInLocation(lp.LastPassFormat, v, eastLoc)
			if err != nil {
				return err
			}
			e.Time = t.UTC()
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
func (s *EventService) GetEventReport(username, search string, from, to lp.JsonLastPassTime) (*Events, error) {
	s.command = "reporting"
	s.data = struct {
		From   lp.JsonLastPassTime `json:"from"`
		To     lp.JsonLastPassTime `json:"to"`
		Search string                         `json:"search"`
		User   string                         `json:"user"`
		Format string                         `json:"format"`
	}{User: username, From: from, To: to, Format: "siem"}

	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var events Events
	err = lp.JSONBodyDecoder(res, &events)
	if err != nil {
		err = errors.New("Failed parsing response body.")
		return nil, err
	}

	return &events, nil
}

// GetAllEventReports fetches event of all users in certain period of time.
// Filtering is also available by setting search string.
func (s *EventService) GetAllEventReports(from, to lp.JsonLastPassTime) (*Events, error) {
	s.GetEventReport("allusers", "", from, to)
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var events Events
	err = lp.JSONBodyDecoder(res, &events)
	if err != nil {
		return nil, err
	}

	return &events, nil
}

// GetAPIEventReports retrieves events triggered by API.
// We first call
// s.GetEventReport("api", "", from, to) will return error "Please select a valid user."
func (s *EventService) GetAPIEventReports(from, to lp.JsonLastPassTime) (*Events, error) {
	events, err := s.GetAllEventReports(from, to)
	if err != nil {
		return nil, err
	}
	return events.GetUserEvents("API"), nil
}