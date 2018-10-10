package proc

import "github.com/gojektech/proctor/proc/env"

type Metadata struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Author       string   `json:"author"`
	Contributors string   `json:"contributors"`
	Organization string   `json:"organization"`
	EnvVars      env.Vars `json:"env_vars"`
}
