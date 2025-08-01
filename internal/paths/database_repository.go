package paths

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"hideout/internal/common/apperror"
	"hideout/internal/common/ordering"
	"hideout/internal/common/pagination"
	"time"
)

type DatabaseRepository struct {
	conn               *gorm.DB
	inMemoryRepository *InMemoryRepository
}

func NewDatabaseRepository(conn *gorm.DB, inMemoryRep *InMemoryRepository) DatabaseRepository {
	return DatabaseRepository{conn: conn, inMemoryRepository: inMemoryRep}
}

func (m DatabaseRepository) GetID(ctx context.Context) (uint, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetID(ctx)
	}

	id := uint(0)
	errScan := m.conn.Table(TableName).Select("COALESCE(MAX(id), 0)").Row().Scan(&id)
	return id + 1, errScan
}

func (m DatabaseRepository) Load(ctx context.Context) ([]Path, error) {
	var results []Path
	errGetRecords := m.conn.Table(TableName).Select([]string{TableName + ".*"}).Find(&results).Error
	if errGetRecords != nil {
		return results, errors.Wrap(errGetRecords, "Failed to obtain records in database")
	}

	return results, nil
}

func (m DatabaseRepository) GetMapByID(ctx context.Context, params ListPathParams) (map[uint]*Path, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	var mapResults = make(map[uint]*Path)
	for _, result := range results {
		mapResults[result.ID] = result
	}

	return mapResults, nil
}

func (m DatabaseRepository) GetMapByUID(ctx context.Context, params ListPathParams) (map[string]*Path, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByUID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	var mapResults = make(map[string]*Path)
	for _, result := range results {
		mapResults[result.UID] = result
	}

	return mapResults, nil
}

func (m DatabaseRepository) Get(ctx context.Context, params ListPathParams) ([]*Path, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Get(ctx, params)
	}

	var results []*Path
	Query := m.GetQuery(m.conn, []string{TableName + ".*"}, params)
	errQuery := Query.Find(&results).Error
	if errQuery != nil {
		if errors.Is(errQuery, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrRecordNotFound
		}
		return nil, errQuery
	}

	return results, nil
}

func (m DatabaseRepository) GetByID(ctx context.Context, id uint) (*Path, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByID(ctx, id)
	}

	var result Path
	Query := m.conn.Table(TableName).Select([]string{TableName + ".*"}).Where(TableName+".id = ? AND "+TableName+".deleted_at IS NULL", id)

	errQuery := Query.First(&result).Error
	if errQuery != nil {
		if errors.Is(errQuery, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrRecordNotFound
		}
		return nil, errQuery
	}

	return &result, nil
}

func (m DatabaseRepository) GetByUID(ctx context.Context, uid string) (*Path, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByUID(ctx, uid)
	}

	var result Path
	Query := m.conn.Table(TableName).Select([]string{TableName + ".*"}).Where(TableName+".uid = ? AND "+TableName+".deleted_at IS NULL", uid)

	errQuery := Query.First(&result).Error
	if errQuery != nil {
		if errors.Is(errQuery, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrRecordNotFound
		}
		return nil, errQuery
	}

	return &result, nil
}

func (m DatabaseRepository) Update(ctx context.Context, path Path) (*Path, error) {
	var updatedPathEntry = &path
	errUpdate := m.conn.Table(TableName).Model(&path).Updates(updatedPathEntry).Error
	if errUpdate != nil {
		return nil, errors.Wrapf(errUpdate, "Error updating path with ID of %d in database", path.ID)
	}

	if m.inMemoryRepository != nil {
		updatedPath, errUpdatePath := m.inMemoryRepository.Update(ctx, path)
		if errUpdatePath != nil {
			return nil, errors.Wrapf(errUpdatePath, "Error updating path with ID of %d in memory", path.ID)
		}

		updatedPathEntry = updatedPath
	}

	return updatedPathEntry, nil
}

func (m DatabaseRepository) Create(ctx context.Context, path Path) (*Path, error) {
	var createdPathEntry = &path
	path.CreatedAt = time.Now()
	errCreate := m.conn.Table(TableName).Create(&path).Error
	if errCreate != nil {
		return nil, errors.Wrapf(errCreate, "Error creating path with ID of %d in database", path.ID)
	}

	if m.inMemoryRepository != nil {
		newPathEntry, errCreatePath := m.inMemoryRepository.Create(ctx, *createdPathEntry)
		if errCreatePath != nil {
			return nil, errors.Wrapf(errCreatePath, "Error creating path with parent path ID of %d and name %s in memory", path.ParentID, path.Name)
		}

		createdPathEntry = newPathEntry
	}

	return createdPathEntry, nil
}

func (m DatabaseRepository) Count(ctx context.Context, params ListPathParams) (uint, error) {
	// These are not needed when performing filtering and counting
	params.Pagination = pagination.Pagination{PerPage: 0, Page: 0}
	params.Order = []ordering.Order{}

	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Count(ctx, params)
	}

	var count = uint(0)
	Query := m.GetQuery(m.conn, []string{"count(" + TableName + ".id) as count"}, params)
	errQuery := Query.Find(&count).Error
	return 0, errQuery
}

func (m DatabaseRepository) Delete(ctx context.Context, id uint, forceDelete bool) error {
	if forceDelete {
		errDelete := m.conn.Table(TableName).Unscoped().Delete(&Path{}, id).Error
		if errDelete != nil {
			return errors.Wrapf(errDelete, "Error deleting path with ID of %d in database", id)
		}
	} else {
		errUpdate := m.conn.Table(TableName).Where("id = ?", id).Update("deleted_at",
			sql.NullTime{Valid: true, Time: time.Now()}).Error
		if errUpdate != nil {
			return errors.Wrapf(errUpdate, "Error marking path with ID of %d deleted in database", id)
		}
	}

	if m.inMemoryRepository != nil {
		errDelete := m.inMemoryRepository.Delete(ctx, id, forceDelete)
		if errDelete != nil {
			return errors.Wrapf(errDelete, "Error deleting path with ID of %d in memory", id)
		}
	}

	return nil
}

func (m DatabaseRepository) GetQuery(tx *gorm.DB, selectedColumnNames []string, params ListPathParams) (Query *gorm.DB) {
	conn := m.conn
	if tx != nil {
		conn = tx
	}
	Query = conn.Table(TableName).Select(selectedColumnNames)
	Query = params.DatabaseFilter(TableName, Query)
	return params.DatabaseOrder(TableName, Query, OrderMap)
}
