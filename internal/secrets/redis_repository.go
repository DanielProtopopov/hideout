package secrets

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"hideout/internal/common/model"
	"time"
)

type RedisRepository struct {
	conn               *redis.Client
	inMemoryRepository InMemoryRepository
}

func NewRedisRepository(conn *redis.Client, inMemoryRep InMemoryRepository) RedisRepository {
	return RedisRepository{conn: conn, inMemoryRepository: inMemoryRep}
}

func (m RedisRepository) GetID(ctx context.Context) (uint, error) {
	return m.inMemoryRepository.GetID(ctx)
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

func (m RedisRepository) Update(ctx context.Context, secret Secret) (*Secret, error) {
	var updatedSecretEntry = Secret{Model: model.Model{ID: secret.ID, UpdatedAt: time.Now()}, UID: secret.UID, Value: secret.Value}
	updatedSecretVal, errMarshal := json.Marshal(updatedSecretEntry)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing secret with ID of %d and name %s", secret.ID, secret.Name)
	}
	_, errUpdate := m.conn.Set(ctx, fmt.Sprintf("secret:%d", secret.ID), updatedSecretVal, 0).Result()
	if errUpdate != nil && !errors.Is(errUpdate, redis.Nil) {
		return nil, errors.Wrapf(errUpdate, "Error updating secret with ID of %d in Redis", secret.ID)
	}
	updatedSecret, errUpdateSecret := m.inMemoryRepository.Update(ctx, updatedSecretEntry)
	if errUpdateSecret != nil {
		return nil, errors.Wrapf(errUpdateSecret, "Error updating secret with ID of %d in-memory", secret.ID)
	}
	return updatedSecret, nil
}

func (m RedisRepository) Create(ctx context.Context, secret Secret) (*Secret, error) {
	createdSecretVal, errMarshal := json.Marshal(secret)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing secret with ID of %d and name %s", secret.ID, secret.Name)
	}
	_, errCreate := m.conn.Set(ctx, fmt.Sprintf("secret:%d", secret.ID), createdSecretVal, 0).Result()
	if errCreate != nil && !errors.Is(errCreate, redis.Nil) {
		return nil, errors.Wrapf(errCreate, "Error creating secret with ID of %d", secret.ID)
	}
	newSecret, errCreateSecret := m.inMemoryRepository.Create(ctx, secret)
	if errCreateSecret != nil {
		return nil, errors.Wrapf(errCreateSecret, "Error creating secret with path ID of %d and name %s in-memory", secret.PathID, secret.Name)
	}
	return newSecret, nil
}

func (m RedisRepository) Count(ctx context.Context, params ListSecretParams) (uint, error) {
	return m.inMemoryRepository.Count(ctx, params)
}

func (m RedisRepository) Delete(ctx context.Context, id uint, forceDelete bool) error {
	if forceDelete {
		_, errDelete := m.conn.Del(ctx, fmt.Sprintf("secret:%d", id)).Result()
		if errDelete != nil && !errors.Is(errDelete, redis.Nil) {
			return errors.Wrapf(errDelete, "Error deleting secret with ID of %d", id)
		}
	} else {
		existingSecret, errGetSecret := m.GetByID(ctx, id)
		if errGetSecret != nil {
			return errors.Wrapf(errGetSecret, "Failed to retrieve secret with ID of %d", id)
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

	return m.inMemoryRepository.Delete(ctx, id, forceDelete)
}
