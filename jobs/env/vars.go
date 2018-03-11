package env

type Vars struct {
	Secrets []VarMetadata `json:"secrets"`
	Args    []VarMetadata `json:"args"`
}

type VarMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
