package config_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gojektech/proctor/config"
	"github.com/stretchr/testify/assert"
)

func TestProctorURL(t *testing.T) {
	proctorConfigFilePath := "/tmp/proctor.yaml"
	proctorUrl := []byte("PROCTOR_URL: any-random-url.com")
	err := ioutil.WriteFile(proctorConfigFilePath, proctorUrl, 0644)
	defer os.Remove(proctorConfigFilePath)
	assert.NoError(t, err)

	config.InitConfig()
	configuredProctorURL := config.ProctorURL()

	assert.Equal(t, "any-random-url.com", configuredProctorURL)
}

func TestProctorEmailId(t *testing.T) {
	proctorConfigFilePath := "/tmp/proctor.yaml"
	EmailId := []byte("EMAIL_ID: foobar@gmail.com")
	err := ioutil.WriteFile(proctorConfigFilePath, EmailId, 0644)
	defer os.Remove(proctorConfigFilePath)
	assert.NoError(t, err)

	config.InitConfig()
	configuredEmailId := config.EmailId()

	assert.Equal(t, "foobar@gmail.com", configuredEmailId)
}

func TestProctorAccessToken(t *testing.T) {
	proctorConfigFilePath := "/tmp/proctor.yaml"

	AccessToken := []byte("ACCESS_TOKEN: access-token")
	err := ioutil.WriteFile(proctorConfigFilePath, AccessToken, 0644)
	defer os.Remove(proctorConfigFilePath)
	assert.NoError(t, err)

	config.InitConfig()
	configuredAccessToken := config.AccessToken()

	assert.Equal(t, "access-token", configuredAccessToken)
}
