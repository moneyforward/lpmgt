package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"fmt"
	"os"
	"strings"
	"lastpass_provisioning/logger"
)

// DecodeBody
func DecodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

func JSONReader(v interface{}) (io.Reader, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(v)

	if err != nil {
		return nil, err
	}
	return buf, nil
}

// PrettyPrintJSON output indented json via stdout.
func PrettyPrintJSON(src interface{}) {
	fmt.Fprintln(os.Stdout, JSONMarshalIndent(src, "", "    "))
}

// JSONMarshalIndent call json.MarshalIndent and replace encoded angle brackets
func JSONMarshalIndent(src interface{}, prefix, indent string) string {
	dataRaw, err := json.MarshalIndent(src, prefix, indent)
	logger.DieIf(err)
	return replaceAngleBrackets(string(dataRaw))
}

func replaceAngleBrackets(s string) string {
	s = strings.Replace(s, "\\u003c", "<", -1)
	return strings.Replace(s, "\\u003e", ">", -1)
}