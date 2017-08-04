package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
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
