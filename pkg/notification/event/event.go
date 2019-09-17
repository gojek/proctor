package event

type Event interface {
	Type() string
	User() UserData
	Content() map[string]string
}

type UserData struct {
	Email string
}
