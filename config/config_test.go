package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gojektech/proctor/config"
	"github.com/stretchr/testify/assert"
)

func setUp() {
	os.Setenv(config.Environment, "test")
}

func TestReturnsConfigDirAsTmpIfEnvironmentIsTest(t *testing.T) {
	os.Setenv(config.Environment, "test")
	dir := config.ConfigFileDir()
	assert.Equal(t, "/tmp", dir)
}

func TestReturnsConfigDirAsHomeDotProctorIfEnvironmentIsNotSet(t *testing.T) {
	os.Unsetenv(config.Environment)

	dir := config.ConfigFileDir()
	expectedDir := fmt.Sprintf("%s/.proctor", os.Getenv("HOME"))
	assert.Equal(t, expectedDir, dir)
}

func TestLoadConfigsFromEnvironmentVariables(t *testing.T) {
	setUp()
	proctorHost := "test.example.com"
	email := "user@example.com"
	accessToken := "test-token"
	os.Setenv(config.Host, proctorHost)
	os.Setenv(config.Email, email)
	os.Setenv(config.Token, accessToken)
	configFilePath := createProctorConfigFile(t, "")
	defer os.Remove(configFilePath)

	proctorConfig, err := config.LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, config.ProctorConfig{Host: proctorHost, Email: email, AccessToken: accessToken}, proctorConfig)
}

func TestLoadConfigFromFile(t *testing.T) {
	setUp()
	os.Unsetenv(config.Host)
	os.Unsetenv(config.Email)
	os.Unsetenv(config.Token)

	configFilePath := createProctorConfigFile(t, "PROCTOR_HOST: file.example.com\nEMAIL_ID: file@example.com\nACCESS_TOKEN: file-token")
	defer os.Remove(configFilePath)

	proctorConfig, err := config.LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, config.ProctorConfig{Host: "file.example.com", Email: "file@example.com", AccessToken: "file-token"}, proctorConfig)
}

func createProctorConfigFile(t *testing.T, content string) string {
	proctorHost := []byte(fmt.Sprintf(content))
	configFilePath := fmt.Sprintf("%s/proctor.yaml", config.ConfigFileDir())
	err := ioutil.WriteFile(configFilePath, proctorHost, 0644)
	assert.NoError(t, err)
	return configFilePath
}

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
