package lpmgt

import (
	"net/http"
	"strings"
	"time"
	"encoding/json"
	"github.com/pkg/errors"
)

// Events structure represents LastPass events.
type Events struct {
	Events []Event `json:"events"`
}

// GetUserEvents get events from users
func (es *Events) GetUserEvents(username string) *Events {
	var events []Event
	for _, event := range es.Events {
		if username == event.Username {
			events = append(events, event)
		}
	}

	return &Events{Events:events}
}

// ConvertTimezone overwrite events in new timezone.
func (es *Events) ConvertTimezone(timezone *time.Location) {
	for index, event := range es.Events {
		es.Events[index].Time = event.Time.UTC().In(timezone)
	}
}

// Event represents event data in LastPass
type Event struct {
	Time      time.Time `json:"JSONTime"`
	Username  string    `json:"Username,omitempty"`
	IPAddress string    `json:"IP_Address,omitempty"`
	Action    string    `json:"Action,omitempty"`
	Data      string    `json:"Data,omitempty"`
	ID        string    `json:"ID,omitempty"`
}

func (e *Event) String(timezone *time.Location) string {
	return e.Time.UTC().In(timezone).String() + " " + e.Username + " " + e.IPAddress + " " + e.Action + " " + e.Data
}

// UnmarshalJSON is written because it has a value(time) in a special format.
func (e *Event) UnmarshalJSON(b []byte) error {
	var rawStrings map[string]string

	err := json.Unmarshal(b, &rawStrings)
	if err != nil {
		return errors.Wrapf(err, "Failed UnMarshalling object b: %v", b)
	}

	for k, v := range rawStrings {
		switch strings.ToLower(k) {
		case "time":
			eastLoc, err := time.LoadLocation(LastPassTimeZone)
			if err != nil {
				return err
			}
			t, err := time.ParseInLocation(LastPassFormat, v, eastLoc)
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

// IsAuditEvent checks whether Event is one to be audited.
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

// EventService is a service class that handles event objects in LastPass.
type EventService struct {
	client  *LastPassClient
	command string
	data    interface{}
}

// NewEventService creates a new EventService
func NewEventService(client *LastPassClient) (s *EventService) {
	return &EventService{client: client}
}

func (s *EventService) doRequest() (*http.Response, error) {
	res, err := s.client.DoRequest(s.command, s.data)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetEventReport fetches event of an user in certain period of time.
// Filtering is also available by setting search string.
func (s *EventService) GetEventReport(username, search string, from, to JSONLastPassTime) (*Events, error) {
	s.command = "reporting"
	s.data = struct {
		From   JSONLastPassTime `json:"from"`
		To     JSONLastPassTime `json:"to"`
		Search string                         `json:"search"`
		User   string                         `json:"user"`
		Format string                         `json:"format"`
	}{User: username, From: from, To: to, Format: "siem"}

	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var events Events
	err = JSONBodyDecoder(res, &events)
	if err != nil {
		return nil, err
	}

	return &events, nil
}

// GetAllEventReports fetches event of all users in certain period of time.
// Filtering is also available by setting search string.
func (s *EventService) GetAllEventReports(from, to JSONLastPassTime) (*Events, error) {
	s.GetEventReport("allusers", "", from, to)
	res, err := s.doRequest()
	if err != nil {
		return nil, err
	}

	var events Events
	err = JSONBodyDecoder(res, &events)
	if err != nil {
		return nil, err
	}

	return &events, nil
}

// GetAPIEventReports retrieves events triggered by API.
// We first call
// s.GetEventReport("api", "", from, to) will return error "Please select a valid user."
func (s *EventService) GetAPIEventReports(from, to JSONLastPassTime) (*Events, error) {
	events, err := s.GetAllEventReports(from, to)
	if err != nil {
		return nil, err
	}
	return events.GetUserEvents("API"), nil
}