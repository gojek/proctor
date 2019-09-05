package http

import (
	"net/http"
)

func NewClient() (*http.Client, error) {
	httpClient := &http.Client{}
	return httpClient, nil
}
