package http

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"net/http"

	"github.com/gojektech/proctor/proctord/config"
)

func NewClient() (*http.Client, error) {
	caCert, err := base64.StdEncoding.DecodeString(config.KubeCACertEncoded())
	if err != nil {
		return &http.Client{}, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
	return httpClient, err
}
