package metadata

import "github.com/gojektech/proctor/proctord/jobs/metadata/env"

type Metadata struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Author       string   `json:"author"`
	Contributors string   `json:"contributors"`
	Organization string   `json:"organization"`
	ImageName    string   `json:"image_name"`
	EnvVars      env.Vars `json:"env_vars"`
}
