package service

import (
	"fmt"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/infra/plugin"
	"proctor/pkg/auth"
	"sync"
)

type SecurityService interface {
	initializePlugin() error
	auth.Auth
}

type securityService struct {
	goPlugin           plugin.GoPlugin
	pluginBinary       string
	pluginExportedName string
	authInstance       auth.Auth
	once               sync.Once
}

func (s *securityService) Verify(userDetail auth.UserDetail, requiredGroups []string) (bool, error) {
	err := s.initializePlugin()
	logger.LogErrors(err, "initialize plugin")
	if err != nil {
		return false, err
	}
	return s.authInstance.Verify(userDetail, requiredGroups)
}

func (s *securityService) Auth(email string, token string) (*auth.UserDetail, error) {
	err := s.initializePlugin()
	logger.LogErrors(err, "initialize plugin")
	if err != nil {
		return nil, err
	}
	return s.authInstance.Auth(email, token)
}

func (s *securityService) initializePlugin() error {
	s.once.Do(func() {
		raw, err := s.goPlugin.Load(s.pluginBinary, s.pluginExportedName)
		logger.LogErrors(err, "Load GoPlugin binary")
		if err != nil {
			return
		}
		authInstance := *raw.(*auth.Auth)
		if authInstance == nil {
			logger.Error("Failed to convert exported plugin to auth.Auth type")
			return
		}
		s.authInstance = authInstance
	})

	if s.authInstance == nil {
		return fmt.Errorf("failed to load and instantiate *auth.Auth from plugin location %s and exported name %s", s.pluginBinary, s.pluginExportedName)
	} else {
		return nil
	}
}

func NewSecurityService(pluginBinary string, exportedName string, goPlugin plugin.GoPlugin) SecurityService {
	return &securityService{
		pluginBinary:       pluginBinary,
		pluginExportedName: exportedName,
		goPlugin:           goPlugin,
		once:               sync.Once{},
	}
}
