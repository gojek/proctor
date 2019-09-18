package message

type Message interface {
	JSON() (string, error)
}
