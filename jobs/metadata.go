package jobs

import "github.com/gojekfarm/proctor/jobs/env"

type Metadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	EnvVars     env.Vars `json:"env_vars"`
}
