package message

import "encoding/json"

type Message interface {
	JSON() (string, error)
}

type standardMessage struct {
	Text string `json:"text"`
}

func (messageObject *standardMessage) JSON() (string, error) {
	byteMessage, err := json.Marshal(messageObject)
	if err != nil {
		return "", err
	}
	return string(byteMessage), nil
}

func NewStandardMessage(text string) Message {
	return &standardMessage{
		Text: text,
	}
}
