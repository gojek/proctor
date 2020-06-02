package message

import (
	"encoding/json"
	"errors"
	"proctor/pkg/notification/event"
	"sort"
)

type executionMessage struct {
	evt    event.Event
	Blocks []sectionComponent `json:"blocks"`
}

func (messageObject *executionMessage) JSON() (string, error) {
	evt := messageObject.evt
	if evt.Type() != event.ExecutionEventType {
		return "", errors.New("event type mismatch")
	}
	userEmail := evt.User().Email
	messageObject.Blocks = []sectionComponent{}

	section := sectionComponent{}
	section.Type = "section"
	section.Text = textComponent{
		Type: "plain_text",
		Text: userEmail + " execute job with details:",
	}
	section.Fields = []textComponent{}
	contents := evt.Content()
	var keys []string
	for key := range contents {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := contents[key]
		keyComponent := textComponent{}
		keyComponent.Type = "mrkdwn"
		keyComponent.Text = "*" + key + "*"
		valueComponent := textComponent{}
		valueComponent.Type = "plain_text"
		valueComponent.Text = value

		section.Fields = append(section.Fields, keyComponent, valueComponent)
	}

	messageObject.Blocks = append(messageObject.Blocks, section)
	byteMessage, err := json.Marshal(messageObject)
	if err != nil {
		return "", err
	}
	return string(byteMessage), nil
}

func NewExecutionMessage(evt event.Event) Message {
	return &executionMessage{
		evt: evt,
	}
}
