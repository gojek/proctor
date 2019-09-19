package service

import (
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/infra/plugin"
	"proctor/pkg/notification"
	"proctor/pkg/notification/event"
	"strings"
	"sync"
)

type NotificationService interface {
	Notify(evt event.Event)
}

type notificationService struct {
	observers           []notification.Observer
	goPlugin            plugin.GoPlugin
	pluginsBinary       []string
	pluginsExportedName string
	once                sync.Once
}

func (s *notificationService) Notify(evt event.Event) {
	s.initializePlugin()
	for _, observer := range s.observers {
		err := observer.OnNotify(evt)
		logger.LogErrors(err, "notify event to observer", evt, observer)
	}
}

func (s *notificationService) initializePlugin() {
	s.once.Do(func() {
		s.observers = []notification.Observer{}
		pluginsBinary := s.pluginsBinary
		pluginsExported := strings.Split(s.pluginsExportedName, ",")

		for idx, pluginBinary := range pluginsBinary {
			raw, err := s.goPlugin.Load(pluginBinary, pluginsExported[idx])
			logger.LogErrors(err, "Load GoPlugin binary")
			if err != nil {
				return
			}
			observer := *raw.(*notification.Observer)
			if observer == nil {
				logger.Error("Failed to convert exported plugin to notification.Observer type")
				return
			}
			s.observers = append(s.observers, observer)
		}
		logger.Info("Number of notification plugin ", len(s.observers))
	})
}

func NewNotificationService(pluginsBinary []string, pluginsExportedName string, goPlugin plugin.GoPlugin) NotificationService {
	return &notificationService{
		goPlugin:            goPlugin,
		pluginsBinary:       pluginsBinary,
		pluginsExportedName: pluginsExportedName,
		once:                sync.Once{},
	}
}
