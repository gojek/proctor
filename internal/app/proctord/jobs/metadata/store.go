package metadata

import (
	"encoding/json"
	"proctor/internal/app/proctord/redis"
	modelMetadata "proctor/internal/pkg/model/metadata"
)

const JobNameKeySuffix = "-metadata"

type Store interface {
	CreateOrUpdateJobMetadata(metadata modelMetadata.Metadata) error
	GetAllJobsMetadata() ([]modelMetadata.Metadata, error)
	GetJobMetadata(jobName string) (*modelMetadata.Metadata, error)
}

type store struct {
	redisClient redis.Client
}

func NewStore(redisClient redis.Client) Store {
	return &store{
		redisClient: redisClient,
	}
}

func jobMetadataKey(jobName string) string {
	return jobName + JobNameKeySuffix
}

func (store *store) CreateOrUpdateJobMetadata(metadata modelMetadata.Metadata) error {
	jobNameKey := jobMetadataKey(metadata.Name)

	binaryJobMetadata, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return store.redisClient.SET(jobNameKey, binaryJobMetadata)
}

func (store *store) GetAllJobsMetadata() ([]modelMetadata.Metadata, error) {
	jobNameKeyRegex := "*" + JobNameKeySuffix

	keys, err := store.redisClient.KEYS(jobNameKeyRegex)
	if err != nil {
		return nil, err
	}

	jobKeys := make([]interface{}, len(keys))
	for i := range keys {
		jobKeys[i] = keys[i]
	}
	values, err := store.redisClient.MGET(jobKeys...)
	if err != nil {
		return nil, err
	}

	jobsMetadata := make([]modelMetadata.Metadata, len(values))
	for i := range values {
		err = json.Unmarshal(values[i], &jobsMetadata[i])
		if err != nil {
			return nil, err
		}
	}

	return jobsMetadata, nil
}

func (store *store) GetJobMetadata(jobName string) (*modelMetadata.Metadata, error) {
	binaryJobMetadata, err := store.redisClient.GET(jobMetadataKey(jobName))
	if err != nil {
		return nil, err
	}

	var jobMetadata modelMetadata.Metadata
	err = json.Unmarshal(binaryJobMetadata, &jobMetadata)
	if err != nil {
		return nil, err
	}

	return &jobMetadata, nil
}
