package lastpass_format

import (
	"time"
)

const LastPassFormat = "2006-01-02 15:04:05"
const LastPassTimeZone = "US/Eastern"

type JsonLastPassTime struct {
	JsonTime time.Time
}

func (j JsonLastPassTime) Format() string {
	return j.JsonTime.Format(LastPassFormat)
}

func (j JsonLastPassTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + j.Format() + `"`), nil
}
