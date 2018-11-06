package config

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gojektech/proctor/io"
	"os"

	"github.com/spf13/viper"
)

const (
	Environment = "ENVIRONMENT"
	ProctorHost = "PROCTOR_HOST"
	EmailId     = "EMAIL_ID"
	AccessToken = "ACCESS_TOKEN"
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
		configFileUsed := viper.ConfigFileUsed()
		if _, err := os.Stat(configFileUsed); os.IsNotExist(err) {
			bytes, _ := dataConfig_templateYamlBytes()
			template := string(bytes)
			io.GetPrinter().Println(fmt.Sprintf("Config file not found in %s/proctor.yaml", ConfigFileDir()), color.FgRed)
			io.GetPrinter().Println(fmt.Sprintf("Create a config file with template:\n\n%s\n\n", template), color.FgGreen)
		}
		return ProctorConfig{}, err
	}

	proctorHost := viper.GetString(ProctorHost)
	emailId := viper.GetString(EmailId)
	accessToken := viper.GetString(AccessToken)
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
