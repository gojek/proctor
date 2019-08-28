package gate

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"proctor/internal/app/service/infra/logger"
	"proctor/pkg/auth"
)

type GateClient interface {
	GetUserProfile(email string, token string) (*auth.UserDetail, error)
}

type gateClient struct {
	config      GateConfig
	host        string
	profilePath string
	protocol    string
	restClient  *resty.Client
}

func (g *gateClient) GetUserProfile(email string, token string) (*auth.UserDetail, error) {
	path := fmt.Sprintf("%s://%s/%s", g.protocol, g.host, g.profilePath)
	formData := map[string]string{
		"email":        email,
		"access_token": token,
	}

	profile := &Profile{}
	response, err := g.restClient.
		R().
		SetHeader("Accept", "application/json").
		SetQueryParams(formData).
		SetResult(profile).
		Get(path)

	logger.LogErrors(err, "get request to %s, with email %s", path, email)
	if err != nil {
		return nil, err
	}

	if response.StatusCode() == 401 {
		return nil, fmt.Errorf("authentication failed, please check your access token")
	}
	if response.StatusCode() == 404 {
		return nil, fmt.Errorf("user not found for email %s", email)
	}

	if response.IsSuccess() {
		userDetail := auth.UserDetail{
			Name:   profile.Name,
			Email:  profile.Email,
			Active: profile.Active,
			Groups: profile.getGroups(),
		}

		return &userDetail, nil

	} else {
		return nil, fmt.Errorf("request to %s with email %s failed, returned body %s", path, email, response.String())
	}
}

func NewGateClient(client *resty.Client) GateClient {
	config := NewGateConfig()
	return &gateClient{
		config:      config,
		protocol:    config.Protocol,
		host:        config.Host,
		profilePath: config.ProfilePath,
		restClient:  client,
	}
}

type Profile struct {
	Email  string  `json:"email"`
	Name   string  `json:"name"`
	Active bool    `json:"active"`
	Groups []Group `json:"groups"`
}

type Group struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func (profile *Profile) getGroups() []string {
	var groups []string
	for _, group := range profile.Groups {
		groups = append(groups, group.Name)
	}

	return groups
}
