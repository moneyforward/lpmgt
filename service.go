package lpmgt

import "net/http"

// Service is the interface that executes business logic
type Service interface {
	DoRequest() (*http.Response, error)
}
