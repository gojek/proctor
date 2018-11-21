package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestReturnsConfigDirAsHomeDotProctorIfEnvironmentIsNotSet(t *testing.T) {
	os.Unsetenv(Environment)

	dir := ConfigFileDir()
	expectedDir := fmt.Sprintf("%s/.proctor", os.Getenv("HOME"))
	assert.Equal(t, expectedDir, dir)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

type ConfigTestSuite struct {
	suite.Suite
	configLoader   Loader
	configFilePath string
}

func (s *ConfigTestSuite) SetupTest() {
	os.Setenv(Environment, "test")
	s.configLoader = NewLoader()
	s.configFilePath = fmt.Sprintf("%s/proctor.yaml", ConfigFileDir())
}

func (s *ConfigTestSuite) TearDownTest() {
	os.Unsetenv(ProctorHost)
	os.Unsetenv(EmailId)
	os.Unsetenv(AccessToken)
	os.Unsetenv(ConnectionTimeoutSecs)
	os.Remove(s.configFilePath)
}

func (s *ConfigTestSuite) TestReturnsConfigDirAsTmpIfEnvironmentIsTest() {
	dir := ConfigFileDir()
	assert.Equal(s.T(), "/tmp", dir)
}

func (s *ConfigTestSuite) createProctorConfigFile(content string) {
	fileContent := []byte(fmt.Sprintf(content))
	err := ioutil.WriteFile(s.configFilePath, fileContent, 0644)
	assert.NoError(s.T(), err)
}

func (s *ConfigTestSuite) TestLoadConfigsFromEnvironmentVariables() {
	t := s.T()

	proctorHost := "test.example.com"
	email := "user@example.com"
	accessToken := "test-token"
	os.Setenv(ProctorHost, proctorHost)
	os.Setenv(EmailId, email)
	os.Setenv(AccessToken, accessToken)
	os.Setenv(ConnectionTimeoutSecs, "20")
	s.createProctorConfigFile("")

	proctorConfig, err := s.configLoader.Load()

	assert.Empty(t, err)
	assert.Equal(t, ProctorConfig{Host: proctorHost, Email: email, AccessToken: accessToken, ConnectionTimeoutSecs: time.Duration(20 * time.Second)}, proctorConfig)
}

func (s *ConfigTestSuite) TestLoadConfigFromFile() {
	t := s.T()

	s.createProctorConfigFile("PROCTOR_HOST: file.example.com\nEMAIL_ID: file@example.com\nACCESS_TOKEN: file-token\nCONNECTION_TIMEOUT_SECS: 30")

	proctorConfig, err := s.configLoader.Load()

	assert.Empty(t, err)
	assert.Equal(t, ProctorConfig{Host: "file.example.com", Email: "file@example.com", AccessToken: "file-token", ConnectionTimeoutSecs: time.Duration(30 * time.Second)}, proctorConfig)
}

func (s *ConfigTestSuite) TestCheckForMandatoryConfig() {
	t := s.T()

	s.createProctorConfigFile("EMAIL_ID: file@example.com\nACCESS_TOKEN: file-token\nCONNECTION_TIMEOUT_SECS: 30")

	_, err := s.configLoader.Load()

	assert.Error(t, err, "Config Error!!!\nMandatory config PROCTOR_HOST is missing in Proctor Config file.")
}

func (s *ConfigTestSuite) TestTakesDefaultValueForConfigs() {
	t := s.T()
	s.createProctorConfigFile("PROCTOR_HOST: file.example.com\nEMAIL_ID: file@example.com\nACCESS_TOKEN: file-token")

	proctorConfig, err := s.configLoader.Load()

	assert.Empty(t, err)
	assert.Equal(t, time.Duration(10*time.Second), proctorConfig.ConnectionTimeoutSecs)
}

func (s *ConfigTestSuite) TestShouldPrintInstructionsForConfigFileIfFileNotFound() {
	t := s.T()
	expectedMessage := fmt.Sprintf("Config file not found in %s\nSetup config using `proctor config PROCTOR_HOST=some.host ...`\n\nAlternatively create a config file with template:\n\nPROCTOR_HOST: <host>\nEMAIL_ID: <email>\nACCESS_TOKEN: <access-token>\n", s.configFilePath)

	_, err := s.configLoader.Load()

	assert.Equal(t, expectedMessage, err.Message)
}
