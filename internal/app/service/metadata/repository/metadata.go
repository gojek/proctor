package repository

import (
	"encoding/json"
	"proctor/internal/app/service/infra/db/redis"
	"proctor/internal/pkg/model/metadata"
)

const KeySuffix = "-metadata"

type MetadataRepository interface {
	Save(metadata *metadata.Metadata) error
	GetAll() ([]metadata.Metadata, error)
	GetByName(name string) (*metadata.Metadata, error)
}

type metadataRepository struct {
	redisClient redis.Client
}

func applySuffix(name string) string {
	return name + KeySuffix
}

func NewMetadataRepository(client redis.Client) MetadataRepository {
	return &metadataRepository{
		redisClient: client,
	}
}

func (repository *metadataRepository) Save(metadata *metadata.Metadata) error {
	key := applySuffix(metadata.Name)

	jsonMetadata, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return repository.redisClient.SET(key, jsonMetadata)
}

func (repository *metadataRepository) GetAll() ([]metadata.Metadata, error) {
	searchKey := "*" + KeySuffix

	keys, err := repository.redisClient.KEYS(searchKey)
	if err != nil {
		return nil, err
	}

	availableKeys := make([]interface{}, len(keys))
	for i := range keys {
		availableKeys[i] = keys[i]
	}

	values, err := repository.redisClient.MGET(availableKeys...)
	if err != nil {
		return nil, err
	}

	metadataSlice := make([]metadata.Metadata, len(values))
	for i := range values {
		err = json.Unmarshal(values[i], &metadataSlice[i])
		if err != nil {
			return nil, err
		}
	}

	return metadataSlice, nil
}

func (repository *metadataRepository) GetByName(name string) (*metadata.Metadata, error) {
	binaryMetadata, err := repository.redisClient.GET(applySuffix(name))
	if err != nil {
		return nil, err
	}

	var jobMetadata metadata.Metadata
	err = json.Unmarshal(binaryMetadata, &jobMetadata)
	if err != nil {
		return nil, err
	}

	return &jobMetadata, nil
}
