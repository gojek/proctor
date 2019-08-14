package gate

import "github.com/spf13/viper"

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GATE_PLUGIN")
}

func Protocol() string {
	viper.SetDefault("PROTOCOL", "https")
	return viper.GetString("PROTOCOL")
}

func Host() string {
	viper.SetDefault("HOST", "gate.gojek.co.id")
	return viper.GetString("HOST")
}

func ProfilePath() string {
	viper.SetDefault("PROFILE_PATH", "api/v1/users/profile")
	return viper.GetString("PROFILE_PATH")
}
