package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"hideout/internal/common/model"
)

type RedisRepository struct {
	conn               *redis.Client
	inMemoryRepository InMemoryRepository
}

func NewRedisRepository(conn *redis.Client, inMemoryRep InMemoryRepository) RedisRepository {
	return RedisRepository{conn: conn, inMemoryRepository: inMemoryRep}
}

func (m RedisRepository) getID() uint {
	return m.inMemoryRepository.getID()
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
		return results, errors.Wrap(errGetValues, "Failed to retrieve values by keys from Redis")
	}
	for i, _ := range values {
		if values[i] != nil {
			var result = Secret{}
			var resultString = values[i].(string)
			errUnmarshal := json.Unmarshal([]byte(resultString), &result)
			if errUnmarshal != nil {
				return nil, errors.Wrapf(errUnmarshal, "Failed to unmarshal secret data from Redis")
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func (m RedisRepository) GetMapByID(ctx context.Context, params ListSecretParams) (map[uint]*Secret, error) {
	return m.inMemoryRepository.GetMapByID(ctx, params)
}

func (m RedisRepository) GetMapByUID(ctx context.Context, params ListSecretParams) (map[string]*Secret, error) {
	return m.inMemoryRepository.GetMapByUID(ctx, params)
}

func (m RedisRepository) Get(ctx context.Context, params ListSecretParams) ([]*Secret, error) {
	return m.inMemoryRepository.Get(ctx, params)
}

func (m RedisRepository) GetMapByPath(ctx context.Context, params ListSecretParams) (map[uint][]*Secret, error) {
	return m.inMemoryRepository.GetMapByPath(ctx, params)
}

func (m RedisRepository) GetByID(ctx context.Context, id uint) (*Secret, error) {
	return m.inMemoryRepository.GetByID(ctx, id)
}

func (m RedisRepository) GetByUID(ctx context.Context, uid string) (*Secret, error) {
	return m.inMemoryRepository.GetByUID(ctx, uid)
}

func (m RedisRepository) Update(ctx context.Context, id uint, value string) (*Secret, error) {
	existingSecret, errGetSecret := m.GetByID(ctx, id)
	if errGetSecret != nil {
		return nil, errors.Wrapf(errGetSecret, "Failed to retrieve secret with ID of %d", id)
	}
	var existingSecretName = existingSecret.Name

	var updatedSecretEntry = Secret{Model: model.Model{ID: existingSecret.ID}, UID: existingSecret.UID, Value: value}
	updatedSecretVal, errMarshal := json.Marshal(updatedSecretEntry)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing secret with ID of %d and name %s", existingSecret.ID, existingSecretName)
	}
	_, errUpdate := m.conn.Set(ctx, fmt.Sprintf("secret:%d", id), updatedSecretVal, 0).Result()

	// Revert back to old name
	if errUpdate != nil && !errors.Is(errUpdate, redis.Nil) {
		_, _ = m.inMemoryRepository.Update(ctx, id, existingSecretName)
		return nil, errors.Wrapf(errUpdate, "Error updating secret with ID of %d", id)
	}

	// Update was successful, update in-memory
	updatedSecret, errUpdateSecret := m.inMemoryRepository.Update(ctx, id, value)
	if errUpdateSecret != nil {
		return nil, errors.Wrapf(errUpdateSecret, "Error updating secret with ID of %d", id)
	}

	return updatedSecret, nil
}

func (m RedisRepository) Create(ctx context.Context, pathID uint, name string, value string, valueType string) (*Secret, error) {
	newSecret, errCreateSecret := m.inMemoryRepository.Create(ctx, pathID, name, value, valueType)
	if errCreateSecret != nil {
		return nil, errors.Wrapf(errCreateSecret, "Error creating secret with path ID of %d and name %s", pathID, name)
	}
	createdSecretVal, errMarshal := json.Marshal(newSecret)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing secret with ID of %d and name %s", newSecret.ID, newSecret.Name)
	}

	_, errCreate := m.conn.Set(ctx, fmt.Sprintf("secret:%d", newSecret.ID), createdSecretVal, 0).Result()
	if errCreate != nil && !errors.Is(errCreate, redis.Nil) {
		_ = m.inMemoryRepository.Delete(ctx, newSecret.ID, true)
		return nil, errors.Wrapf(errCreate, "Error creating secret with ID of %d", newSecret.ID)
	}

	return newSecret, nil
}

func (m RedisRepository) Count(ctx context.Context, pathID uint, name string) (uint, error) {
	return m.inMemoryRepository.Count(ctx, pathID, name)
}

func (m RedisRepository) Delete(ctx context.Context, id uint, forceDelete bool) error {
	_, errDelete := m.conn.Del(ctx, fmt.Sprintf("secret:%d", id)).Result()
	if errDelete != nil && !errors.Is(errDelete, redis.Nil) {
		return errors.Wrapf(errDelete, "Error deleting secret with ID of %d", id)
	}

	return m.inMemoryRepository.Delete(ctx, id, true)
}
