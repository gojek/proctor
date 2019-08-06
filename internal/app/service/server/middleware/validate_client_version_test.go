package middleware

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"proctor/internal/app/service/server/middleware/parameter"
	"testing"
)

func getTestHandler() http.HandlerFunc {
	fn := func(rw http.ResponseWriter, req *http.Request) {
	}
	return http.HandlerFunc(fn)
}

func TestValidClientVersionHTTPHeader(t *testing.T) {

	_ = os.Setenv("PROCTOR_MIN_CLIENT_VERSION", "0.2.0")

	ts := httptest.NewServer(ValidateClientVersion(getTestHandler()))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL+"/metadata", nil)
	req.Header.Add(parameter.ClientVersionHeader, "0.2.0")

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestEmptyClientVersionHTTPHeader(t *testing.T) {

	_ = os.Setenv("PROCTOR_MIN_CLIENT_VERSION", "0.2.0")

	ts := httptest.NewServer(ValidateClientVersion(getTestHandler()))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL+"/metadata", nil)

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestInvalidClientVersionHTTPHeader(t *testing.T) {

	_ = os.Setenv("PROCTOR_MIN_CLIENT_VERSION", "0.3.0")

	ts := httptest.NewServer(ValidateClientVersion(getTestHandler()))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL+"/metadata", nil)
	req.Header.Add(parameter.ClientVersionHeader, "0.1.0")

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(body)
	assert.Equal(t, bodyString, "Your Proctor client is using an outdated version: 0.1.0. To continue using proctor, please upgrade to latest version.")
}
