package execution

type Job struct {
	Name        string            `json:"name"`
	Args        map[string]string `json:"args"`
	CallbackURL string            `json:"callback_url"`
}

