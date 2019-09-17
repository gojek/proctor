package slack

import "github.com/spf13/viper"

type SlackConfig struct {
	url string
}

func NewSlackConfig() SlackConfig {
	fang := viper.New()
	fang.AutomaticEnv()
	fang.SetEnvPrefix("SLACK_PLUGIN")
	fang.SetDefault("URL", "https://hooks.slack.com/services")
	config := SlackConfig{
		url: fang.GetString("URL"),
	}
	return config
}
