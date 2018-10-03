package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	var configFileDir string

	if viper.GetString("ENVIRONMENT") == "test" {
		configFileDir = "/tmp"
	} else {
		configFileDir = "$HOME/.proctor"
	}

	viper.AddConfigPath(configFileDir)
	viper.SetConfigName("proctor")

	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("Error reading proctor config")
		os.Exit(1)
	}
}

func ProctorURL() string {
	InitConfig()
	proctorUrl := viper.GetString("PROCTOR_URL")
	return proctorUrl
}

func EmailId() string {
	InitConfig()
	emailId := viper.GetString("EMAIL_ID")
	return emailId
}

func AccessToken() string {
	InitConfig()
	accessToken := viper.GetString("ACCESS_TOKEN")
	return accessToken
}
