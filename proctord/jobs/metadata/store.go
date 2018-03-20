package metadata

import (
	"encoding/json"

	"github.com/gojektech/proctor/proctord/redis"
)

const JobNameKeySuffix = "-metadata"

type Store interface {
	CreateOrUpdateJobMetadata(metadata Metadata) error
	GetAllJobsMetadata() ([]Metadata, error)
	GetJobMetadata(jobName string) (*Metadata, error)
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

func (store *store) CreateOrUpdateJobMetadata(metadata Metadata) error {
	jobNameKey := jobMetadataKey(metadata.Name)

	binaryJobMetadata, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return store.redisClient.SET(jobNameKey, binaryJobMetadata)
}

func (store *store) GetAllJobsMetadata() ([]Metadata, error) {
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

	jobsMetadata := make([]Metadata, len(values))
	for i := range values {
		err = json.Unmarshal(values[i], &jobsMetadata[i])
		if err != nil {
			return nil, err
		}
	}

	return jobsMetadata, nil
}

func (store *store) GetJobMetadata(jobName string) (*Metadata, error) {
	binaryJobMetadata, err := store.redisClient.GET(jobMetadataKey(jobName))
	if err != nil {
		return nil, err
	}

	var jobMetadata Metadata
	err = json.Unmarshal(binaryJobMetadata, &jobMetadata)
	if err != nil {
		return nil, err
	}

	return &jobMetadata, nil
}
