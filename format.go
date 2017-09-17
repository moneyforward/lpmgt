package lastpass_provisioning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const LastPassFormat = "2006-01-02 15:04:05"
const LastPassTimeZone = "US/Eastern"

type JsonLastPassTime struct {
	JsonTime time.Time
}

// Format returns a textual representation of the time value formatted in LastPass Format
func (j JsonLastPassTime) Format() string {
	return j.JsonTime.Format(LastPassFormat)
}

// MarshalJSON encodes golang structure into json format
func (j JsonLastPassTime) MarshalJSON() ([]byte, error) {
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
		return nil, fmt.Errorf("Failed to encode %v into JSON" , v)
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
