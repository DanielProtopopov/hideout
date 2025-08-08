package secrets

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"hideout/internal/common/apperror"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/common/ordering"
	"hideout/internal/common/pagination"
	"time"
)

type RedisRepository struct {
	conn               *redis.Client
	inMemoryRepository *InMemoryRepository
}

func NewRedisRepository(conn *redis.Client, inMemoryRep *InMemoryRepository) RedisRepository {
	return RedisRepository{conn: conn, inMemoryRepository: inMemoryRep}
}

func (m RedisRepository) GetID(ctx context.Context) (uint, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetID(ctx)
	}

	secrets, errLoadSecrets := m.Load(ctx)
	if errLoadSecrets != nil {
		return 0, errors.Wrap(errLoadSecrets, "Failed to load secrets from Redis")
	}

	var maxID = uint(0)
	for _, secret := range secrets {
		if secret.ID >= maxID {
			maxID = secret.ID
		}
	}

	return maxID + 1, nil
}

func (m RedisRepository) Load(ctx context.Context) ([]Secret, error) {
	pattern := "secret:*"
	iter := m.conn.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	var results []Secret
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	values, errGetValues := m.conn.MGet(ctx, keys...).Result()
	if errGetValues != nil {
		return results, errors.Wrap(errGetValues, "Failed to retrieve values by keys in Redis")
	}
	for i, _ := range values {
		if values[i] != nil {
			var result = Secret{}
			var resultString = values[i].(string)
			errUnmarshal := json.Unmarshal([]byte(resultString), &result)
			if errUnmarshal != nil {
				return nil, errors.Wrapf(errUnmarshal, "Failed to unmarshal secret data")
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func (m RedisRepository) GetMapByID(ctx context.Context, params ListSecretParams) (map[uint]*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	var mapResults = make(map[uint]*Secret)
	for _, result := range results {
		mapResults[result.ID] = result
	}

	return mapResults, nil
}

func (m RedisRepository) GetMapByUID(ctx context.Context, params ListSecretParams) (map[string]*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByUID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	var mapResults = make(map[string]*Secret)
	for _, result := range results {
		mapResults[result.UID] = result
	}

	return mapResults, nil
}

func (m RedisRepository) Get(ctx context.Context, params ListSecretParams) ([]*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Get(ctx, params)
	}

	pattern := "secret:*"
	iter := m.conn.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	var results []Secret
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	values, errGetValues := m.conn.MGet(ctx, keys...).Result()
	if errGetValues != nil {
		return nil, errors.Wrap(errGetValues, "Failed to retrieve values by keys in Redis")
	}
	for i, _ := range values {
		if values[i] != nil {
			var result Secret
			var resultString = values[i].(string)
			errUnmarshal := json.Unmarshal([]byte(resultString), &result)
			if errUnmarshal != nil {
				return nil, errors.Wrapf(errUnmarshal, "Failed to unmarshal secret data in Redis")
			}
			results = append(results, result)
		}
	}

	inMemoryRepository := NewInMemoryRepository(&results)
	return inMemoryRepository.Get(ctx, params)
}

func (m RedisRepository) GetMapByFolder(ctx context.Context, params ListSecretParams) (map[uint][]*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByFolder(ctx, params)
	}

	return nil, apperror.ErrNotImplemented
}

func (m RedisRepository) GetByID(ctx context.Context, id uint) (*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByID(ctx, id)
	}

	results, errGetResults := m.GetMapByID(ctx, ListSecretParams{
		ListParams: generics.ListParams{IDs: []uint{id}, Deleted: model.No},
	})
	if errGetResults != nil {
		return nil, errGetResults
	}

	result, exists := results[id]
	if !exists {
		return nil, apperror.ErrRecordNotFound
	}

	return result, nil
}

func (m RedisRepository) GetByUID(ctx context.Context, uid string) (*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByUID(ctx, uid)
	}

	results, errGetResults := m.GetMapByUID(ctx, ListSecretParams{
		ListParams: generics.ListParams{UIDs: []string{uid}, Deleted: model.No},
	})
	if errGetResults != nil {
		return nil, errGetResults
	}

	result, exists := results[uid]
	if !exists {
		return nil, apperror.ErrRecordNotFound
	}

	return result, nil
}

func (m RedisRepository) Update(ctx context.Context, secret Secret) (*Secret, error) {
	var updatedSecretEntry = &secret
	updatedSecretVal, errMarshal := json.Marshal(updatedSecretEntry)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing secret with ID of %d and name %s", updatedSecretEntry.ID, updatedSecretEntry.Name)
	}
	_, errUpdate := m.conn.Set(ctx, fmt.Sprintf("secret:%d", updatedSecretEntry.ID), updatedSecretVal, 0).Result()
	if errUpdate != nil && !errors.Is(errUpdate, redis.Nil) {
		return nil, errors.Wrapf(errUpdate, "Error updating secret with ID of %d in Redis", updatedSecretEntry.ID)
	}
	if m.inMemoryRepository != nil {
		updatedSecret, errUpdateSecret := m.inMemoryRepository.Update(ctx, *updatedSecretEntry)
		if errUpdateSecret != nil {
			return nil, errors.Wrapf(errUpdateSecret, "Error updating secret with ID of %d in memory", secret.ID)
		}

		updatedSecretEntry = updatedSecret
	}

	return updatedSecretEntry, nil
}

func (m RedisRepository) Create(ctx context.Context, secret Secret) (*Secret, error) {
	var createdFolderEntry = &secret
	createdSecretVal, errMarshal := json.Marshal(createdFolderEntry)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing secret with ID of %d and name %s", createdFolderEntry.ID, createdFolderEntry.Name)
	}
	_, errCreate := m.conn.Set(ctx, fmt.Sprintf("secret:%d", createdFolderEntry.ID), createdSecretVal, 0).Result()
	if errCreate != nil && !errors.Is(errCreate, redis.Nil) {
		return nil, errors.Wrapf(errCreate, "Error creating secret with ID of %d in Redis", createdFolderEntry.ID)
	}

	if m.inMemoryRepository != nil {
		newSecret, errCreateSecret := m.inMemoryRepository.Create(ctx, secret)
		if errCreateSecret != nil {
			return nil, errors.Wrapf(errCreateSecret, "Error creating secret with folder ID of %d and name %s in memory", secret.FolderID, secret.Name)
		}

		createdFolderEntry = newSecret
	}

	return createdFolderEntry, nil
}

func (m RedisRepository) Count(ctx context.Context, params ListSecretParams) (uint, error) {
	// These are not needed when performing filtering and counting
	params.Pagination = pagination.Pagination{PerPage: 0, Page: 0}
	params.Order = []ordering.Order{}

	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Count(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return 0, errGetResults
	}

	return uint(len(results)), nil
}

func (m RedisRepository) Delete(ctx context.Context, id uint, forceDelete bool) error {
	if forceDelete {
		_, errDelete := m.conn.Del(ctx, fmt.Sprintf("secret:%d", id)).Result()
		if errDelete != nil && !errors.Is(errDelete, redis.Nil) {
			return errors.Wrapf(errDelete, "Error deleting secret with ID of %d in Redis", id)
		}
	} else {
		existingSecret, errGetSecret := m.GetByID(ctx, id)
		if errGetSecret != nil {
			return errors.Wrapf(errGetSecret, "Failed to retrieve secret with ID of %d in Redis", id)
		}
		existingSecret.DeletedAt = sql.NullTime{Time: time.Now(), Valid: true}
		updatedSecretVal, errMarshal := json.Marshal(existingSecret)
		if errMarshal != nil {
			return errors.Wrapf(errMarshal, "Error serializing secret with ID of %d and name %s", existingSecret.ID, existingSecret.Name)
		}
		_, errUpdate := m.conn.Set(ctx, fmt.Sprintf("secret:%d", id), updatedSecretVal, 0).Result()
		if errUpdate != nil {
			return errors.Wrapf(errUpdate, "Failed to update soft-deleted record in Redis")
		}
	}

	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Delete(ctx, id, forceDelete)
	}

	return nil
}
