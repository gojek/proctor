package event

type Event interface {
	Type() Type
	User() UserData
	Content() map[string]string
}

type UserData struct {
	Email string
}
