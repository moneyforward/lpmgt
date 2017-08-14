package lastpass_time

import (
	"time"
)

const LastpassFormat = "2006-01-02 15:04:05"

type JsonLastPassTime struct {
	JsonTime time.Time
}

func (j JsonLastPassTime) Format() string {
	return j.JsonTime.Format(LastpassFormat)
}

func (j JsonLastPassTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + j.Format() + `"`), nil
}
