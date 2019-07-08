package secrets

type Secret struct {
	JobName string            `json:"job_name"`
	Secrets map[string]string `json:"secrets"`
}
