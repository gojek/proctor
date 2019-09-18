package main

import (
	"github.com/go-resty/resty/v2"
	"proctor/pkg/notification"
	"proctor/pkg/notification/event"
	"proctor/plugins/slack-notification-plugin/slack"
	"proctor/plugins/slack-notification-plugin/slack/message"
)

type slackNotification struct {
	slackClient slack.SlackClient
}

func (notification *slackNotification) OnNotify(evt event.Event) error {
	messageObject := message.NewExecutionMessage(evt)
	err := notification.slackClient.Publish(messageObject)
	return err
}

func newSlackNotification() notification.Observer {
	slackClient := slack.NewSlackClient(resty.New())
	return &slackNotification{
		slackClient: slackClient,
	}
}

var SlackNotification = newSlackNotification()
