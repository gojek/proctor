package slack

type Message interface {
	JSON() (string, error)
}
