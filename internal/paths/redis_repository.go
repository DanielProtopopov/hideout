package paths

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
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

func (m RedisRepository) Load(ctx context.Context) ([]Path, error) {
	pattern := "path:*"
	iter := m.conn.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	var results []Path
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	values, errGetValues := m.conn.MGet(ctx, keys...).Result()
	if errGetValues != nil {
		return results, errors.Wrap(errGetValues, "Failed to retrieve values by keys from Redis")
	}
	for i, _ := range values {
		if values[i] != nil {
			var result = Path{}
			var resultString = values[i].(string)
			errUnmarshal := json.Unmarshal([]byte(resultString), &result)
			if errUnmarshal != nil {
				return nil, errors.Wrapf(errUnmarshal, "Failed to unmarshal path data from Redis")
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func (m RedisRepository) GetMapByID(ctx context.Context, params ListPathParams) (map[uint]*Path, error) {
	return m.inMemoryRepository.GetMapByID(ctx, params)
}

func (m RedisRepository) GetMapByUID(ctx context.Context, params ListPathParams) (map[string]*Path, error) {
	return m.inMemoryRepository.GetMapByUID(ctx, params)
}

func (m RedisRepository) Get(ctx context.Context, params ListPathParams) ([]*Path, error) {
	return m.inMemoryRepository.Get(ctx, params)
}

func (m RedisRepository) GetByID(ctx context.Context, id uint) (*Path, error) {
	return m.inMemoryRepository.GetByID(ctx, id)
}

func (m RedisRepository) GetByUID(ctx context.Context, uid string) (*Path, error) {
	return m.inMemoryRepository.GetByUID(ctx, uid)
}

func (m RedisRepository) Update(ctx context.Context, id uint, name string) (*Path, error) {
	existingPath, errGetPath := m.GetByID(ctx, id)
	if errGetPath != nil {
		return nil, errors.Wrapf(errGetPath, "Failed to retrieve path with ID of %d", id)
	}
	var existingPathName = existingPath.Name
	var updatedPathEntry = Path{
		ID: existingPath.ID, ParentID: existingPath.ParentID, UID: existingPath.UID, Name: name,
	}
	updatedPathVal, errMarshal := json.Marshal(updatedPathEntry)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing path with ID of %d and name %s", existingPath.ID, name)
	}
	_, errUpdate := m.conn.Set(ctx, fmt.Sprintf("path:%d", id), updatedPathVal, 0).Result()
	// Revert back to old name
	if errUpdate != nil && !errors.Is(errUpdate, redis.Nil) {
		_, _ = m.inMemoryRepository.Update(ctx, id, existingPathName)
		return nil, errors.Wrapf(errUpdate, "Error updating path with ID of %d", id)
	}
	// Update was successful, update in-memory
	updatedPath, errUpdatePath := m.inMemoryRepository.Update(ctx, id, name)
	if errUpdatePath != nil {
		return nil, errors.Wrapf(errUpdatePath, "Error updating path with ID of %d and name %s", id, name)
	}

	return updatedPath, nil
}

func (m RedisRepository) Create(ctx context.Context, parentPathID uint, name string) (*Path, error) {
	newPath, errCreatePath := m.inMemoryRepository.Create(ctx, parentPathID, name)
	if errCreatePath != nil {
		return nil, errors.Wrapf(errCreatePath, "Error creating path with parent ID of %d and name %s", parentPathID, name)
	}

	newPathVal, errMarshal := json.Marshal(newPath)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing path with ID of %d and name %s", newPath.ID, name)
	}

	_, errCreate := m.conn.Set(ctx, fmt.Sprintf("path:%d", newPath.ID), newPathVal, 0).Result()
	if errCreate != nil && !errors.Is(errCreate, redis.Nil) {
		_ = m.inMemoryRepository.Delete(ctx, newPath.ID)
		return nil, errors.Wrapf(errCreate, "Error creating path with ID of %d", newPath.ID)
	}

	return newPath, nil
}

func (m RedisRepository) Count(ctx context.Context, name string) (uint, error) {
	return m.inMemoryRepository.Count(ctx, name)
}

func (m RedisRepository) Delete(ctx context.Context, id uint) error {
	_, errDelete := m.conn.Del(ctx, fmt.Sprintf("path:%d", id)).Result()
	if errDelete != nil && !errors.Is(errDelete, redis.Nil) {
		return errors.Wrapf(errDelete, "Error deleting path with ID of %d", id)
	}

	return m.inMemoryRepository.Delete(ctx, id)
}
