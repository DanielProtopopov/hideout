package paths

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

	paths, errLoadPaths := m.Load(ctx)
	if errLoadPaths != nil {
		return 0, errors.Wrap(errLoadPaths, "Failed to load paths from Redis")
	}

	var maxID = uint(0)
	for _, path := range paths {
		if path.ID >= maxID {
			maxID = path.ID
		}
	}

	return maxID, nil
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
		return results, errors.Wrap(errGetValues, "Failed to retrieve values by keys in Redis")
	}
	for i, _ := range values {
		if values[i] != nil {
			var result = Path{}
			var resultString = values[i].(string)
			errUnmarshal := json.Unmarshal([]byte(resultString), &result)
			if errUnmarshal != nil {
				return nil, errors.Wrapf(errUnmarshal, "Failed to unmarshal path data in Redis")
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func (m RedisRepository) GetMapByID(ctx context.Context, params ListPathParams) (map[uint]*Path, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	resultsMap := make(map[uint]*Path)
	for _, result := range results {
		resultsMap[result.ID] = result
	}

	return resultsMap, nil
}

func (m RedisRepository) GetMapByUID(ctx context.Context, params ListPathParams) (map[string]*Path, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByUID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	resultsMap := make(map[string]*Path)
	for _, result := range results {
		resultsMap[result.UID] = result
	}

	return resultsMap, nil
}

func (m RedisRepository) Get(ctx context.Context, params ListPathParams) ([]*Path, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Get(ctx, params)
	}

	pattern := "path:*"
	iter := m.conn.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	var results []Path
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	values, errGetValues := m.conn.MGet(ctx, keys...).Result()
	if errGetValues != nil {
		return nil, errors.Wrap(errGetValues, "Failed to retrieve values by keys in Redis")
	}
	for i, _ := range values {
		if values[i] != nil {
			var result Path
			var resultString = values[i].(string)
			errUnmarshal := json.Unmarshal([]byte(resultString), &result)
			if errUnmarshal != nil {
				return nil, errors.Wrapf(errUnmarshal, "Failed to unmarshal path data in Redis")
			}
			results = append(results, result)
		}
	}

	inMemoryRepository := NewInMemoryRepository(&results)
	return inMemoryRepository.Get(ctx, params)
}

func (m RedisRepository) GetByID(ctx context.Context, id uint) (*Path, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByID(ctx, id)
	}

	resultsMap, errGetResults := m.GetMapByID(ctx, ListPathParams{ListParams: generics.ListParams{IDs: []uint{id}, Deleted: model.No}})
	if errGetResults != nil {
		return nil, errGetResults
	}

	result, exists := resultsMap[id]
	if !exists {
		return nil, apperror.ErrRecordNotFound
	}

	return result, nil
}

func (m RedisRepository) GetByUID(ctx context.Context, uid string) (*Path, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByUID(ctx, uid)
	}

	resultsMap, errGetResults := m.GetMapByUID(ctx, ListPathParams{ListParams: generics.ListParams{UIDs: []string{uid}, Deleted: model.No}})
	if errGetResults != nil {
		return nil, errGetResults
	}

	result, exists := resultsMap[uid]
	if !exists {
		return nil, apperror.ErrRecordNotFound
	}

	return result, nil
}

func (m RedisRepository) Update(ctx context.Context, path Path) (*Path, error) {
	var updatedPathEntry = &path
	updatedPathVal, errMarshal := json.Marshal(updatedPathEntry)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing path with ID of %d and name %s", updatedPathEntry.ID, path.Name)
	}
	_, errUpdate := m.conn.Set(ctx, fmt.Sprintf("path:%d", path.ID), updatedPathVal, 0).Result()
	if errUpdate != nil && !errors.Is(errUpdate, redis.Nil) {
		return nil, errors.Wrapf(errUpdate, "Error updating path with ID of %d in Redis", path.ID)
	}

	if m.inMemoryRepository != nil {
		updatedPath, errUpdatePath := m.inMemoryRepository.Update(ctx, *updatedPathEntry)
		if errUpdatePath != nil {
			return nil, errors.Wrapf(errUpdatePath, "Error updating path with ID of %d and name %s in memory", path.ID, path.Name)
		}

		updatedPathEntry = updatedPath
	}

	return updatedPathEntry, nil
}

func (m RedisRepository) Create(ctx context.Context, path Path) (*Path, error) {
	var createdPathEntry = &path
	newPathVal, errMarshal := json.Marshal(createdPathEntry)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing path with ID of %d and name %s", createdPathEntry.ID, createdPathEntry.Name)
	}
	_, errCreate := m.conn.Set(ctx, fmt.Sprintf("path:%d", createdPathEntry.ID), newPathVal, 0).Result()
	if errCreate != nil && !errors.Is(errCreate, redis.Nil) {
		return nil, errors.Wrapf(errCreate, "Error creating path with ID of %d in memory", createdPathEntry.ID)
	}

	if m.inMemoryRepository != nil {
		newPath, errCreatePath := m.inMemoryRepository.Create(ctx, *createdPathEntry)
		if errCreatePath != nil {
			return nil, errors.Wrapf(errCreatePath, "Error creating path with parent ID of %d and name %s in memory", path.ParentID, path.Name)
		}
		createdPathEntry = newPath
	}

	return createdPathEntry, nil
}

func (m RedisRepository) Count(ctx context.Context, params ListPathParams) (uint, error) {
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
		_, errDelete := m.conn.Del(ctx, fmt.Sprintf("path:%d", id)).Result()
		if errDelete != nil && !errors.Is(errDelete, redis.Nil) {
			return errors.Wrapf(errDelete, "Error deleting path with ID of %d in Redis", id)
		}
	} else {
		existingPath, errGetPath := m.GetByID(ctx, id)
		if errGetPath != nil {
			return errors.Wrapf(errGetPath, "Failed to retrieve path with ID of %d in Redis", id)
		}
		existingPath.DeletedAt = sql.NullTime{Time: time.Now(), Valid: true}
		updatedPathVal, errMarshal := json.Marshal(existingPath)
		if errMarshal != nil {
			return errors.Wrapf(errMarshal, "Error serializing path with ID of %d and name %s", existingPath.ID, existingPath.Name)
		}
		_, errUpdate := m.conn.Set(ctx, fmt.Sprintf("path:%d", id), updatedPathVal, 0).Result()
		if errUpdate != nil {
			return errors.Wrapf(errUpdate, "Failed to update soft-deleted record in Redis")
		}
	}

	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Delete(ctx, id, forceDelete)
	}

	return nil
}
