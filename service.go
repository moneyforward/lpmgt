package main

import "net/http"

type Service interface {
	DoRequest() (*http.Response, error)
}
