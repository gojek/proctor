package main

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"proctor/pkg/notification"
	"proctor/pkg/notification/event"
	"proctor/plugins/slack-notification-plugin/slack"
)

type slackNotification struct {
	slackClient slack.SlackClient
}

func (notification *slackNotification) OnNotify(evt event.Event) error {
	evtDataJSON, err := json.Marshal(evt.Content())
	if err != nil {
		return err
	}
	textMessage := "User: " + evt.User().Email + "\n"
	textMessage += "Execute job with detail: "
	textMessage += string(evtDataJSON)
	message := slack.NewStandardMessage(textMessage)
	err = notification.slackClient.Publish(message)
	return err
}

func newSlackNotification() notification.Observer {
	slackClient := slack.NewSlackClient(resty.New())
	return &slackNotification{
		slackClient: slackClient,
	}
}

var SlackNotification = newSlackNotification()
