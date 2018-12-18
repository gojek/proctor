package proc

import "github.com/gojektech/proctor/proc/env"

type Metadata struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Author           string   `json:"author"`
	Contributors     string   `json:"contributors"`
	Organization     string   `json:"organization"`
	AuthorizedGroups []string `json:"authorized_groups"`
	EnvVars          env.Vars `json:"env_vars"`
}
