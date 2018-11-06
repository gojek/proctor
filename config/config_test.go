package config

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gojektech/proctor/io"
	"io/ioutil"
	"os"
	"testing"
	"time"

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
	os.Setenv(ProctorHost, proctorHost)
	os.Setenv(EmailId, email)
	os.Setenv(AccessToken, accessToken)
	os.Setenv(ConnectionTimeoutSecs, "20")
	configFilePath := createProctorConfigFile(t, "")
	defer os.Remove(configFilePath)

	proctorConfig, err := LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, ProctorConfig{Host: proctorHost, Email: email, AccessToken: accessToken, ConnectionTimeoutSecs: time.Duration(20 * time.Second)}, proctorConfig)
}

func TestLoadConfigFromFile(t *testing.T) {
	setUp()
	unsetEnvs()

	configFilePath := createProctorConfigFile(t, "PROCTOR_HOST: file.example.com\nEMAIL_ID: file@example.com\nACCESS_TOKEN: file-token\nCONNECTION_TIMEOUT_SECS: 30")
	defer os.Remove(configFilePath)

	proctorConfig, err := LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, ProctorConfig{Host: "file.example.com", Email: "file@example.com", AccessToken: "file-token", ConnectionTimeoutSecs: time.Duration(30 * time.Second)}, proctorConfig)
}

func TestTakesDefaultValueForConfigs(t *testing.T) {
	setUp()
	unsetEnvs()
	configFilePath := createProctorConfigFile(t, "PROCTOR_HOST: file.example.com\nEMAIL_ID: file@example.com\nACCESS_TOKEN: file-token")
	defer os.Remove(configFilePath)

	proctorConfig, err := LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, time.Duration(10*time.Second), proctorConfig.ConnectionTimeoutSecs)
}

func TestShouldPrintInstructionsForConfigFileIfFileNotFound(t *testing.T) {
	setUp()
	configFilePath := fmt.Sprintf("%s/proctor.yaml", ConfigFileDir())
	os.Remove(configFilePath)

	mockPrinter := &io.MockPrinter{}
	io.SetupMockPrinter(mockPrinter)
	defer io.ResetPrinter()
	mockPrinter.On("Println", fmt.Sprintf("Config file not found in %s", configFilePath), color.FgRed).Once()
	helpMessage := "Create a config file with template:\n\nPROCTOR_HOST: <host>\nEMAIL_ID: <email>\nACCESS_TOKEN: <access-token>\n\n"
	mockPrinter.On("Println", helpMessage, color.FgGreen).Once()

	_, err := LoadConfig()

	assert.Error(t, err)
	mockPrinter.AssertExpectations(t)
}

func unsetEnvs() {
	os.Unsetenv(ProctorHost)
	os.Unsetenv(EmailId)
	os.Unsetenv(AccessToken)
	os.Unsetenv(ConnectionTimeoutSecs)
}

func createProctorConfigFile(t *testing.T, content string) string {
	proctorHost := []byte(fmt.Sprintf(content))
	configFilePath := fmt.Sprintf("%s/proctor.yaml", ConfigFileDir())
	err := ioutil.WriteFile(configFilePath, proctorHost, 0644)
	assert.NoError(t, err)
	return configFilePath
}
