package lastpass_time

import (
	"time"
	//"encoding/json"
	"encoding/json"
	"fmt"
)

const LastpassFormat  = "2006-01-02 15:04:05"

type JsonLastPassTime struct {
	time.Time
}

func (j JsonLastPassTime) Format() string {
	return j.Time.Format(LastpassFormat)
}

func (j JsonLastPassTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + j.Format() + `"`), nil
}

func (j JsonLastPassTime) UnmarshalJSON(b []byte) error {
	var s string

	loc, _ := time.LoadLocation("EST")
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	t, err := time.Parse(LastpassFormat, s)
	if err == nil {
		j.Time = t.In(loc)
		fmt.Println(j)
	}

	return err
}