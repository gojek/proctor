package message

import (
	"encoding/json"
	"proctor/pkg/notification/event"
)

type standardMessage struct {
	Text string `json:"text"`
	evt  event.Event
}

func (messageObject *standardMessage) JSON() (string, error) {
	evtDataJSON, err := json.Marshal(messageObject.evt.Content())
	if err != nil {
		return "", err
	}
	textMessage := "User: " + messageObject.evt.User().Email + "\n"
	textMessage += "Emit event" + messageObject.evt.Type() + " with detail: "
	textMessage += string(evtDataJSON)
	messageObject.Text = textMessage
	byteMessage, err := json.Marshal(messageObject)
	if err != nil {
		return "", err
	}
	return string(byteMessage), nil
}

func NewStandardMessage(evt event.Event) Message {
	return &standardMessage{
		evt: evt,
	}
}
