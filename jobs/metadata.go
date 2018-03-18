package jobs

import "github.com/gojektech/proctor/jobs/env"

type Metadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	EnvVars     env.Vars `json:"env_vars"`
}
