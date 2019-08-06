package execution

type ExecutionResult struct {
	ExecutionId   uint64 `json:"id"`
	JobName       string `json:"job_name"`
	ExecutionName string `json:"name"`
	ImageTag      string `json:"image_tag"`
	CreatedAt     string `json:"created_at"`
	Status        string `json:"status"`
}
