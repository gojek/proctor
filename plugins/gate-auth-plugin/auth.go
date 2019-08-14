package main

import (
	"proctor/internal/app/service/infra/logger"
	"proctor/pkg/auth"
	"proctor/plugins/gate-auth-plugin/gate"
)

type gateAuth struct {
	gateClient gate.GateClient
}

func (g *gateAuth) Auth(email string, token string) (*auth.UserDetail, error) {
	userDetail, err := g.gateClient.GetUserProfile(email, token)
	logger.LogErrors(err, "get user detail with email: ", email)
	if err != nil {
		return nil, err
	}

	return userDetail, nil
}

//Verify check whether user has all required groups
func (g *gateAuth) Verify(userDetail auth.UserDetail, requiredGroups []string) (bool, error) {
	if !userDetail.Active {
		return false, nil
	}

	for _, requiredGroup := range requiredGroups {
		if !contains(userDetail.Groups, requiredGroup) {
			return false, nil
		}
	}

	return true, nil
}

func contains(groups []string, value string) bool {
	for _, group := range groups {
		if group == value {
			return true
		}
	}
	return false
}

func newGateAuth() auth.Auth {
	return &gateAuth{
		gateClient: gate.NewGateClient(),
	}
}

var Auth = newGateAuth()
