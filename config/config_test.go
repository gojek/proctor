package config_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gojektech/proctor/config"
	"github.com/stretchr/testify/assert"
)

func TestProctorHost(t *testing.T) {
	proctorConfigFilePath := "/tmp/proctor.yaml"
	proctorHost := []byte("PROCTOR_HOST: any-random-host.com")
	err := ioutil.WriteFile(proctorConfigFilePath, proctorHost, 0644)
	defer os.Remove(proctorConfigFilePath)
	assert.NoError(t, err)

	config.InitConfig()
	configuredProctorHost := config.ProctorHost()

	assert.Equal(t, "any-random-host.com", configuredProctorHost)
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
