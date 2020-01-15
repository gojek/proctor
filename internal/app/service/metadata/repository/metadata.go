package repository

import (
	"encoding/json"
	"proctor/internal/app/service/infra/db/redis"
	"proctor/internal/pkg/model/metadata"
)

const KeySuffix = "-metadata"

type MetadataRepository interface {
	Save(metadata metadata.Metadata) error
	GetAll() ([]metadata.Metadata, error)
	GetAllByGroups(groups []string) ([]metadata.Metadata, error)
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

func (repository *metadataRepository) Save(metadata metadata.Metadata) error {
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

func (repository *metadataRepository) GetAllByGroups(groups []string) ([]metadata.Metadata, error) {
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

	var filteredMetadata []metadata.Metadata
	for _, meta := range metadataSlice {
		if len(meta.AuthorizedGroups) == 0 {
			filteredMetadata = append(filteredMetadata, meta)
		} else if duplicateItemExists(meta.AuthorizedGroups, groups) {
			filteredMetadata = append(filteredMetadata, meta)
		}
	}

	return filteredMetadata, nil
}

func duplicateItemExists(first []string, second []string) bool {
	for _, firstString := range first {
		for _, secondString := range second {
			if firstString == secondString {
				return true
			}
		}
	}
	return false
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
