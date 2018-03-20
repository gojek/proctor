package metadata

import "github.com/gojektech/proctor/proctord/jobs/metadata/env"

type Metadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ImageName   string   `json:"image_name"`
	EnvVars     env.Vars `json:"env_vars"`
}
