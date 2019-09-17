package service

import (
	"proctor/pkg/notification/event"
)

type NotificationService interface {
	Notify(evt event.Event)
}
