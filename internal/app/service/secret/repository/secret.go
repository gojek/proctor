package repository

import (
	"encoding/json"
	"proctor/internal/app/service/infra/db/redis"
	"proctor/internal/app/service/secret/model"
)

const KeySuffix = "-secret"

type SecretRepository interface {
	Save(secret model.Secret) error
	GetByJobName(jobName string) (map[string]string, error)
}

type secretRepository struct {
	redisClient redis.Client
}

func NewSecretRepository(redisClient redis.Client) SecretRepository {
	return &secretRepository{
		redisClient: redisClient,
	}
}

func applySuffix(jobName string) string {
	return jobName + KeySuffix
}

func (store *secretRepository) Save(secret model.Secret) error {
	key := applySuffix(secret.JobName)

	jsonSecret, err := json.Marshal(secret.Secrets)
	if err != nil {
		return err
	}
	return store.redisClient.SET(key, jsonSecret)
}

func (store *secretRepository) GetByJobName(jobName string) (map[string]string, error) {
	var secrets map[string]string
	jsonSecrets, err := store.redisClient.GET(applySuffix(jobName))
	if err != nil {
		return secrets, err
	}

	err = json.Unmarshal(jsonSecrets, &secrets)
	return secrets, err
}
