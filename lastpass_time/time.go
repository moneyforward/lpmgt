package lastpass_time

import "time"

type JsonLastPassTime struct {
	time.Time
}

func (j JsonLastPassTime) Format() string {
	return j.Time.Format("2006-01-02 15:04:05")
}

func (j JsonLastPassTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + j.Format() + `"`), nil
}
