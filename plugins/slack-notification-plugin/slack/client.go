package slack

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"proctor/plugins/slack-notification-plugin/slack/message"
)

type SlackClient interface {
	Publish(messageObject message.Message) error
}

type slackClient struct {
	client *resty.Client
	config SlackConfig
}

func (s *slackClient) Publish(messageObject message.Message) error {
	messageJson, err := messageObject.JSON()
	if err != nil {
		return err
	}
	path := s.config.url
	response, err := s.client.R().
		SetBody(messageJson).
		SetHeader("Content-Type", "application/json").
		Post(path)
	if err != nil {
		return err
	}
	if response.IsError() {
		return errors.New(response.String())
	}
	return err
}

func NewSlackClient(client *resty.Client) SlackClient {
	return &slackClient{
		client: client,
		config: NewSlackConfig(),
	}
}
