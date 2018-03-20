package execution

type Job struct {
	Name string            `json:"name"`
	Args map[string]string `json:"args"`
}
