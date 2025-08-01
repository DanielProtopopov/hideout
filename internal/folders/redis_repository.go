package folders

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

	folders, errLoadFolders := m.Load(ctx)
	if errLoadFolders != nil {
		return 0, errors.Wrap(errLoadFolders, "Failed to load folders from Redis")
	}

	var maxID = uint(0)
	for _, folder := range folders {
		if folder.ID >= maxID {
			maxID = folder.ID
		}
	}

	return maxID, nil
}

func (m RedisRepository) Load(ctx context.Context) ([]Folder, error) {
	pattern := "folder:*"
	iter := m.conn.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	var results []Folder
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	values, errGetValues := m.conn.MGet(ctx, keys...).Result()
	if errGetValues != nil {
		return results, errors.Wrap(errGetValues, "Failed to retrieve values by keys in Redis")
	}
	for i, _ := range values {
		if values[i] != nil {
			var result = Folder{}
			var resultString = values[i].(string)
			errUnmarshal := json.Unmarshal([]byte(resultString), &result)
			if errUnmarshal != nil {
				return nil, errors.Wrapf(errUnmarshal, "Failed to unmarshal folder data in Redis")
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func (m RedisRepository) GetMapByID(ctx context.Context, params ListFolderParams) (map[uint]*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	resultsMap := make(map[uint]*Folder)
	for _, result := range results {
		resultsMap[result.ID] = result
	}

	return resultsMap, nil
}

func (m RedisRepository) GetMapByUID(ctx context.Context, params ListFolderParams) (map[string]*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByUID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	resultsMap := make(map[string]*Folder)
	for _, result := range results {
		resultsMap[result.UID] = result
	}

	return resultsMap, nil
}

func (m RedisRepository) Get(ctx context.Context, params ListFolderParams) ([]*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Get(ctx, params)
	}

	pattern := "folder:*"
	iter := m.conn.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	var results []Folder
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	values, errGetValues := m.conn.MGet(ctx, keys...).Result()
	if errGetValues != nil {
		return nil, errors.Wrap(errGetValues, "Failed to retrieve values by keys in Redis")
	}
	for i, _ := range values {
		if values[i] != nil {
			var result Folder
			var resultString = values[i].(string)
			errUnmarshal := json.Unmarshal([]byte(resultString), &result)
			if errUnmarshal != nil {
				return nil, errors.Wrapf(errUnmarshal, "Failed to unmarshal folder data in Redis")
			}
			results = append(results, result)
		}
	}

	inMemoryRepository := NewInMemoryRepository(&results)
	return inMemoryRepository.Get(ctx, params)
}

func (m RedisRepository) GetByID(ctx context.Context, id uint) (*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByID(ctx, id)
	}

	resultsMap, errGetResults := m.GetMapByID(ctx, ListFolderParams{ListParams: generics.ListParams{IDs: []uint{id}, Deleted: model.No}})
	if errGetResults != nil {
		return nil, errGetResults
	}

	result, exists := resultsMap[id]
	if !exists {
		return nil, apperror.ErrRecordNotFound
	}

	return result, nil
}

func (m RedisRepository) GetByUID(ctx context.Context, uid string) (*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByUID(ctx, uid)
	}

	resultsMap, errGetResults := m.GetMapByUID(ctx, ListFolderParams{ListParams: generics.ListParams{UIDs: []string{uid}, Deleted: model.No}})
	if errGetResults != nil {
		return nil, errGetResults
	}

	result, exists := resultsMap[uid]
	if !exists {
		return nil, apperror.ErrRecordNotFound
	}

	return result, nil
}

func (m RedisRepository) Update(ctx context.Context, folder Folder) (*Folder, error) {
	var updatedFolderEntry = &folder
	updatedFolderVal, errMarshal := json.Marshal(updatedFolderEntry)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing folder with ID of %d and name %s", updatedFolderEntry.ID, folder.Name)
	}
	_, errUpdate := m.conn.Set(ctx, fmt.Sprintf("folder:%d", folder.ID), updatedFolderVal, 0).Result()
	if errUpdate != nil && !errors.Is(errUpdate, redis.Nil) {
		return nil, errors.Wrapf(errUpdate, "Error updating folder with ID of %d in Redis", folder.ID)
	}

	if m.inMemoryRepository != nil {
		updatedFolder, errUpdateFolder := m.inMemoryRepository.Update(ctx, *updatedFolderEntry)
		if errUpdateFolder != nil {
			return nil, errors.Wrapf(errUpdateFolder, "Error updating folder with ID of %d and name %s in memory", folder.ID, folder.Name)
		}

		updatedFolderEntry = updatedFolder
	}

	return updatedFolderEntry, nil
}

func (m RedisRepository) Create(ctx context.Context, folder Folder) (*Folder, error) {
	var createdFolderEntry = &folder
	newFolderVal, errMarshal := json.Marshal(createdFolderEntry)
	if errMarshal != nil {
		return nil, errors.Wrapf(errMarshal, "Error serializing folder with ID of %d and name %s", createdFolderEntry.ID, createdFolderEntry.Name)
	}
	_, errCreate := m.conn.Set(ctx, fmt.Sprintf("folder:%d", createdFolderEntry.ID), newFolderVal, 0).Result()
	if errCreate != nil && !errors.Is(errCreate, redis.Nil) {
		return nil, errors.Wrapf(errCreate, "Error creating folder with ID of %d in memory", createdFolderEntry.ID)
	}

	if m.inMemoryRepository != nil {
		newFolder, errCreateFolder := m.inMemoryRepository.Create(ctx, *createdFolderEntry)
		if errCreateFolder != nil {
			return nil, errors.Wrapf(errCreateFolder, "Error creating folder with parent ID of %d and name %s in memory", folder.ParentID, folder.Name)
		}
		createdFolderEntry = newFolder
	}

	return createdFolderEntry, nil
}

func (m RedisRepository) Count(ctx context.Context, params ListFolderParams) (uint, error) {
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
		_, errDelete := m.conn.Del(ctx, fmt.Sprintf("folder:%d", id)).Result()
		if errDelete != nil && !errors.Is(errDelete, redis.Nil) {
			return errors.Wrapf(errDelete, "Error deleting folder with ID of %d in Redis", id)
		}
	} else {
		existingFolder, errGetFolder := m.GetByID(ctx, id)
		if errGetFolder != nil {
			return errors.Wrapf(errGetFolder, "Failed to retrieve folder with ID of %d in Redis", id)
		}
		existingFolder.DeletedAt = sql.NullTime{Time: time.Now(), Valid: true}
		updatedFolderVal, errMarshal := json.Marshal(existingFolder)
		if errMarshal != nil {
			return errors.Wrapf(errMarshal, "Error serializing folder with ID of %d and name %s", existingFolder.ID, existingFolder.Name)
		}
		_, errUpdate := m.conn.Set(ctx, fmt.Sprintf("folder:%d", id), updatedFolderVal, 0).Result()
		if errUpdate != nil {
			return errors.Wrapf(errUpdate, "Failed to update soft-deleted record in Redis")
		}
	}

	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Delete(ctx, id, forceDelete)
	}

	return nil
}
