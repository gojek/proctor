package slack

import (
	"github.com/go-resty/resty/v2"
)

type SlackClient interface {
	Publish(message Message) error
}

type slackClient struct {
	client *resty.Client
	config SlackConfig
}

func (s *slackClient) Publish(message Message) error {
	messageJson, err := message.JSON()
	if err != nil {
		return err
	}
	path := s.config.url
	_, err = s.client.R().
		SetBody(messageJson).
		SetHeader("Content-Type", "application/json").
		Post(path)
	return err
}

func NewSlackClient(client *resty.Client) SlackClient {
	return &slackClient{
		client: client,
		config: NewSlackConfig(),
	}
}
