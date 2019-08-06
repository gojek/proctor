package config

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"proctor/internal/pkg/constant"

	"github.com/spf13/viper"
)

const (
	Environment                  = "ENVIRONMENT"
	ProctorHost                  = "PROCTOR_HOST"
	EmailId                      = "EMAIL_ID"
	AccessToken                  = "ACCESS_TOKEN"
	ConnectionTimeoutSecs        = "CONNECTION_TIMEOUT_SECS"
	ProcExecutionStatusPollCount = "PROC_EXECUTION_STATUS_POLL_COUNT"
)

type ProctorConfig struct {
	Host                         string
	Email                        string
	AccessToken                  string
	ConnectionTimeoutSecs        time.Duration
	ProcExecutionStatusPollCount int
}

type ConfigError struct {
	error
	Message string
}

func (c *ConfigError) RootError() error {
	return c.error
}

type Loader interface {
	Load() (ProctorConfig, ConfigError)
}

type loader struct{}

func NewLoader() Loader {
	return &loader{}
}

func (loader *loader) Load() (ProctorConfig, ConfigError) {
	viper.SetDefault(ConnectionTimeoutSecs, 10)
	viper.SetDefault(ProcExecutionStatusPollCount, 30)
	viper.AutomaticEnv()

	viper.AddConfigPath(ConfigFileDir())
	viper.SetConfigName("proctor")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()

	if err != nil {
		configFileUsed := viper.ConfigFileUsed()
		message := ""
		if _, err := os.Stat(configFileUsed); os.IsNotExist(err) {
			bytes, _ := dataConfig_templateYamlBytes()
			template := string(bytes)
			message = fmt.Sprintf("Config file not found in %s/proctor.yaml\n", ConfigFileDir())
			message += fmt.Sprintf("Setup config using `proctor config PROCTOR_HOST=some.host ...`\n\n")
			message += fmt.Sprintf("Alternatively create a config file with template:\n\n%s\n", template)
		}
		return ProctorConfig{}, ConfigError{error: err, Message: message}
	}

	proctorHost := viper.GetString(ProctorHost)
	if proctorHost == "" {
		return ProctorConfig{}, ConfigError{error: errors.New("Mandatory Config Missing"), Message: constant.ConfigProctorHostMissingError}
	}
	emailId := viper.GetString(EmailId)
	accessToken := viper.GetString(AccessToken)
	connectionTimeout := time.Duration(viper.GetInt(ConnectionTimeoutSecs)) * time.Second
	procExecutionStatusPollCount := viper.GetInt(ProcExecutionStatusPollCount)

	return ProctorConfig{Host: proctorHost, Email: emailId, AccessToken: accessToken, ConnectionTimeoutSecs: connectionTimeout, ProcExecutionStatusPollCount: procExecutionStatusPollCount}, ConfigError{}
}

// Returns Config file directory
// This allows to test on dev environment without conflicting with installed proctor config file
func ConfigFileDir() string {
	localConfigDir, localConfigAvailable := os.LookupEnv("LOCAL_CONFIG_DIR")
	if localConfigAvailable {
		return localConfigDir
	} else if os.Getenv(Environment) == "test" {
		return "/tmp"
	} else {
		return fmt.Sprintf("%s/.proctor", os.Getenv("HOME"))
	}
}
