package http

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"proctor/proctord/config"
)

func TestNewClient(t *testing.T) {
	caCert, _ := base64.StdEncoding.DecodeString(config.KubeCACertEncoded())
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	expectedTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	httpClient, err := NewClient()

	assert.NoError(t, err)
	assert.Equal(t, expectedTransport, httpClient.Transport)
}
