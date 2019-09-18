package main

import (
	"encoding/json"
	"proctor/pkg/notification"
	"proctor/pkg/notification/event"
	"proctor/plugins/slack-notification-plugin/slack"
)

type slackNotification struct {
	slackClient slack.SlackClient
}

func (notification *slackNotification) OnNotify(evt event.Event) error {
	evtDataJSON, _ := json.Marshal(evt.Content())
	textMessage := "User: " + evt.User().Email + "\n"
	textMessage += "Execute job with detail: "
	message := slack.NewStandardMessage(string(evtDataJSON))
	_ = notification.slackClient.Publish(message)
	return nil
}

func NewSlackNotification(slackClient slack.SlackClient) notification.Observer {
	return &slackNotification{
		slackClient: slackClient,
	}
}
