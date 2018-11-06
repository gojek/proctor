package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const (
	Environment = "ENVIRONMENT"
	Host        = "PROCTOR_HOST"
	Email       = "EMAIL_ID"
	Token       = "ACCESS_TOKEN"
)

type ProctorConfig struct {
	Host        string
	Email       string
	AccessToken string
}

func LoadConfig() (ProctorConfig, error) {
	viper.AutomaticEnv()

	viper.AddConfigPath(ConfigFileDir())
	viper.SetConfigName("proctor")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()

	if err != nil {
		return ProctorConfig{}, err
	}

	proctorHost := viper.GetString(Host)
	emailId := viper.GetString(Email)
	accessToken := viper.GetString(Token)
	return ProctorConfig{Host: proctorHost, Email: emailId, AccessToken: accessToken}, nil
}

// Returns Config file directory
// Returns /tmp if environment variable `ENVIRONMENT` is set to test, otherwise returns $HOME/.proctor
// This allows to test on dev environment without conflicting with installed proctor config file
func ConfigFileDir() string {
	if os.Getenv("ENVIRONMENT") == "test" {
		return "/tmp"
	} else {
		return fmt.Sprintf("%s/.proctor", os.Getenv("HOME"))
	}
}
