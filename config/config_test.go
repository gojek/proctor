package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setUp() {
	os.Setenv(Environment, "test")
}

func TestReturnsConfigDirAsTmpIfEnvironmentIsTest(t *testing.T) {
	os.Setenv(Environment, "test")
	dir := ConfigFileDir()
	assert.Equal(t, "/tmp", dir)
}

func TestReturnsConfigDirAsHomeDotProctorIfEnvironmentIsNotSet(t *testing.T) {
	os.Unsetenv(Environment)

	dir := ConfigFileDir()
	expectedDir := fmt.Sprintf("%s/.proctor", os.Getenv("HOME"))
	assert.Equal(t, expectedDir, dir)
}

func TestLoadConfigsFromEnvironmentVariables(t *testing.T) {
	setUp()
	proctorHost := "test.example.com"
	email := "user@example.com"
	accessToken := "test-token"
	os.Setenv(Host, proctorHost)
	os.Setenv(Email, email)
	os.Setenv(Token, accessToken)
	configFilePath := createProctorConfigFile(t, "")
	defer os.Remove(configFilePath)

	proctorConfig, err := LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, ProctorConfig{Host: proctorHost, Email: email, AccessToken: accessToken}, proctorConfig)
}

func TestLoadConfigFromFile(t *testing.T) {
	setUp()
	os.Unsetenv(Host)
	os.Unsetenv(Email)
	os.Unsetenv(Token)

	configFilePath := createProctorConfigFile(t, "PROCTOR_HOST: file.example.com\nEMAIL_ID: file@example.com\nACCESS_TOKEN: file-token")
	defer os.Remove(configFilePath)

	proctorConfig, err := LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, ProctorConfig{Host: "file.example.com", Email: "file@example.com", AccessToken: "file-token"}, proctorConfig)
}

func createProctorConfigFile(t *testing.T, content string) string {
	proctorHost := []byte(fmt.Sprintf(content))
	configFilePath := fmt.Sprintf("%s/proctor.yaml", ConfigFileDir())
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

	InitConfig()
	configuredProctorHost := ProctorHost()

	assert.Equal(t, "any-random-host.com", configuredProctorHost)
}

func TestProctorEmailId(t *testing.T) {
	proctorConfigFilePath := "/tmp/proctor.yaml"
	emailId := []byte("EMAIL_ID: foobar@gmail.com")
	err := ioutil.WriteFile(proctorConfigFilePath, emailId, 0644)
	defer os.Remove(proctorConfigFilePath)
	assert.NoError(t, err)

	InitConfig()
	configuredEmailId := EmailId()

	assert.Equal(t, "foobar@gmail.com", configuredEmailId)
}

func TestProctorAccessToken(t *testing.T) {
	proctorConfigFilePath := "/tmp/proctor.yaml"

	accessToken := []byte("ACCESS_TOKEN: access-token")
	err := ioutil.WriteFile(proctorConfigFilePath, accessToken, 0644)
	defer os.Remove(proctorConfigFilePath)
	assert.NoError(t, err)

	InitConfig()
	configuredAccessToken := AccessToken()

	assert.Equal(t, "access-token", configuredAccessToken)
}
