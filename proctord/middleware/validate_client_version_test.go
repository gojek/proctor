package middleware

import (
	"proctor/proctord/utility"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func getTestHandler() http.HandlerFunc {
	fn := func(rw http.ResponseWriter, req *http.Request) {
	}
	return http.HandlerFunc(fn)
}

func TestValidClientVersionHttpHeader(t *testing.T) {

	os.Setenv("PROCTOR_MIN_CLIENT_VERSION","0.2.0")

	ts := httptest.NewServer(ValidateClientVersion(getTestHandler()))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL+"/jobs/metadata", nil)
	req.Header.Add(utility.ClientVersionHeaderKey, "0.2.0")

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestEmptyClientVersionHttpHeader(t *testing.T) {

	os.Setenv("PROCTOR_MIN_CLIENT_VERSION","0.2.0")

	ts := httptest.NewServer(ValidateClientVersion(getTestHandler()))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL+"/jobs/metadata", nil)

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestInvalidClientVersionHttpHeader(t *testing.T) {

	os.Setenv("PROCTOR_MIN_CLIENT_VERSION","0.3.0")

	ts := httptest.NewServer(ValidateClientVersion(getTestHandler()))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL+"/jobs/metadata", nil)
	req.Header.Add(utility.ClientVersionHeaderKey, "0.1.0")

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(body)
	assert.Equal(t,bodyString,"Your Proctor client is using an outdated version: 0.1.0 . To continue using proctor, please upgrade to latest version.")
}
