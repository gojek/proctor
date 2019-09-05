package auth

type Auth interface {
	Auth(email string, token string) (*UserDetail, error)
	Verify(userDetail UserDetail, requiredGroups []string) (bool, error)
}

type UserDetail struct {
	Name   string
	Email  string
	Active bool
	Groups  []string
}
