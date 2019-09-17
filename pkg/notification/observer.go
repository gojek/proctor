package notification

import "proctor/pkg/notification/event"

type Observer interface {
	OnNotify(evt event.Event)
}
