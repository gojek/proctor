package gate

import "github.com/spf13/viper"

type GateConfig struct {
	Protocol    string
	Host        string
	ProfilePath string
	viper       viper.Viper
}

func NewGateConfig() GateConfig {
	fang := viper.New()
	fang.AutomaticEnv()
	fang.SetEnvPrefix("GATE_PLUGIN")
	fang.SetDefault("PROTOCOL", "https")

	fang.SetDefault("HOST", "gate.gojek.co.id")
	fang.SetDefault("PROFILE_PATH", "api/v1/users/profile")
	config := GateConfig{
		Protocol:    fang.GetString("PROTOCOL"),
		Host:        fang.GetString("HOST"),
		ProfilePath: fang.GetString("PROFILE_PATH"),
		viper:       viper.Viper{},
	}

	return config
}
