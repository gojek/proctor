package secrets

import (
	"encoding/json"

	"proctor/proctord/redis"
)

const JobsSecretsKeySuffix = "-secret"

type Store interface {
	CreateOrUpdateJobSecret(Secret) error
	GetJobSecrets(string) (map[string]string, error)
}

type store struct {
	redisClient redis.Client
}

func NewStore(redisClient redis.Client) Store {
	return &store{
		redisClient: redisClient,
	}
}

func jobSecretsKey(jobName string) string {
	return jobName + JobsSecretsKeySuffix
}

func (store *store) CreateOrUpdateJobSecret(secret Secret) error {
	jobSecretsKey := jobSecretsKey(secret.JobName)

	binaryJobSecrets, err := json.Marshal(secret.Secrets)
	if err != nil {
		return err
	}
	return store.redisClient.SET(jobSecretsKey, binaryJobSecrets)
}

func (store *store) GetJobSecrets(jobName string) (map[string]string, error) {
	var secrets map[string]string
	binarySecrets, err := store.redisClient.GET(jobSecretsKey(jobName))
	if err != nil {
		return secrets, err
	}

	err = json.Unmarshal(binarySecrets, &secrets)
	return secrets, err
}
