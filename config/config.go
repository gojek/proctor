package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigType("yaml")

	home := "$HOME/.proctor"
	viper.AddConfigPath(home)

	viper.SetConfigName("proctor")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error reading proctor config")
		os.Exit(1)
	}
}

func ProctorURL() string {
	InitConfig()
	proctorUrl := viper.GetString("PROCTOR_URL")
	if len(proctorUrl) == 0 {
		fmt.Println("proctor url not configured")
		os.Exit(1)
	}
	return proctorUrl
}
