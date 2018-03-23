package proc

import "github.com/gojektech/proctor/proc/env"

type Metadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	EnvVars     env.Vars `json:"env_vars"`
}
