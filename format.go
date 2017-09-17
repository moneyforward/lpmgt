package lpmgt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// LastPassFormat is a format used LastPass Provisioning API
	LastPassFormat = "2006-01-02 15:04:05"
	// LastPassTimeZone is only location acceptable to  LastPass Provisioning API
	LastPassTimeZone = "US/Eastern"
)

// JSONLastPassTime is a golang structure used in LastPass
type JSONLastPassTime struct {
	JSONTime time.Time
}

// Format returns a textual representation of the time value formatted in LastPass Format
func (j JSONLastPassTime) Format() string {
	return j.JSONTime.Format(LastPassFormat)
}

// MarshalJSON encodes golang structure into json format
func (j JSONLastPassTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + j.Format() + `"`), nil
}

// JSONBodyDecoder reads the next JSON-encoded value from its
// input and stores it in the value pointed to by out.
func JSONBodyDecoder(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

// JSONReader reads the value and converts it to JSON-encoded value
func JSONReader(v interface{}) (io.Reader, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(v)

	if err != nil {
		return nil, fmt.Errorf("Failed to encode %v into JSON\n" , v)
	}

	return buf, nil
}

// PrintIndentedJSON output indented json via stdout.
func PrintIndentedJSON(originalJSON interface{}) error {
	dataRaw, err := IndentedJSON(originalJSON)

	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(dataRaw))
	return nil
}

// IndentedJSON returns api.
func IndentedJSON(originalJSON interface{}) ([]byte, error) {
	return json.MarshalIndent(originalJSON, "", "    ")
}
