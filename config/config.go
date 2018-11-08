package config

import (
	"fmt"
	"github.com/gojektech/proctor/proctord/utility"
	"github.com/pkg/errors"
	"os"
	"time"

	"github.com/spf13/viper"
)

const (
	Environment           = "ENVIRONMENT"
	ProctorHost           = "PROCTOR_HOST"
	EmailId               = "EMAIL_ID"
	AccessToken           = "ACCESS_TOKEN"
	ConnectionTimeoutSecs = "CONNECTION_TIMEOUT_SECS"
)

type ProctorConfig struct {
	Host                  string
	Email                 string
	AccessToken           string
	ConnectionTimeoutSecs time.Duration
}

type ConfigError struct {
	error
	Message string
}

func (c *ConfigError) RootError() error {
	return c.error
}

func LoadConfig() (ProctorConfig, ConfigError) {
	viper.SetDefault(ConnectionTimeoutSecs, 10)
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
			message += fmt.Sprintf("Create a config file with template:\n\n%s\n", template)
		}
		return ProctorConfig{}, ConfigError{error: err, Message: message}
	}

	proctorHost := viper.GetString(ProctorHost)
	if proctorHost == "" {
		return ProctorConfig{}, ConfigError{error: errors.New("Mandatory Config Missing"), Message: utility.ConfigProctorHostMissingError}
	}
	emailId := viper.GetString(EmailId)
	accessToken := viper.GetString(AccessToken)
	connectionTimeout := time.Duration(viper.GetInt(ConnectionTimeoutSecs)) * time.Second
	return ProctorConfig{Host: proctorHost, Email: emailId, AccessToken: accessToken, ConnectionTimeoutSecs: connectionTimeout}, ConfigError{}
}

// Returns Config file directory
// This allows to test on dev environment without conflicting with installed proctor config file
func ConfigFileDir() string {
	if os.Getenv(Environment) == "test" {
		return "/tmp"
	} else {
		return fmt.Sprintf("%s/.proctor", os.Getenv("HOME"))
	}
}
