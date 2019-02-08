package execution

type Job struct {
	Name        string            `json:"name"`
	Args        map[string]string `json:"args"`
	CallbackApi string            `json:"callback_api"`
}

