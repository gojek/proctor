package message

import (
	"encoding/json"
	"errors"
	"proctor/pkg/notification/event"
)

type executionMessage struct {
	evt    event.Event
	Blocks []sectionComponent `json:"blocks"`
}

func (messageObject *executionMessage) JSON() (string, error) {
	evt := messageObject.evt
	if evt.Type() != string(event.ExecutionEventType) {
		return "", errors.New("event type mismatch")
	}
	userEmail := evt.User().Email
	messageObject.Blocks = []sectionComponent{}

	section := sectionComponent{}
	section.Type = "section"
	section.Text = textComponent{
		Type: "text",
		Text: userEmail + " execute job with details:",
	}
	section.Fields = []textComponent{}
	for key, value := range evt.Content() {
		keyComponent := textComponent{}
		keyComponent.Type = "mrkdwn"
		keyComponent.Text = "*" + key + "*"
		valueComponent := textComponent{}
		valueComponent.Type = "text"
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
