package folders

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

func (m DatabaseRepository) Load(ctx context.Context) ([]Folder, error) {
	var results []Folder
	errGetRecords := m.conn.Table(TableName).Select([]string{TableName + ".*"}).Find(&results).Error
	if errGetRecords != nil {
		return results, errors.Wrap(errGetRecords, "Failed to obtain records in database")
	}

	return results, nil
}

func (m DatabaseRepository) GetMapByID(ctx context.Context, params ListFolderParams) (map[uint]*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	var mapResults = make(map[uint]*Folder)
	for _, result := range results {
		mapResults[result.ID] = result
	}

	return mapResults, nil
}

func (m DatabaseRepository) GetMapByUID(ctx context.Context, params ListFolderParams) (map[string]*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByUID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	var mapResults = make(map[string]*Folder)
	for _, result := range results {
		mapResults[result.UID] = result
	}

	return mapResults, nil
}

func (m DatabaseRepository) Get(ctx context.Context, params ListFolderParams) ([]*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Get(ctx, params)
	}

	var results []*Folder
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

func (m DatabaseRepository) GetByID(ctx context.Context, id uint) (*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByID(ctx, id)
	}

	var result Folder
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

func (m DatabaseRepository) GetByUID(ctx context.Context, uid string) (*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByUID(ctx, uid)
	}

	var result Folder
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

func (m DatabaseRepository) Update(ctx context.Context, folder Folder) (*Folder, error) {
	var updatedFolderEntry = &folder
	errUpdate := m.conn.Table(TableName).Model(&folder).Updates(updatedFolderEntry).Error
	if errUpdate != nil {
		return nil, errors.Wrapf(errUpdate, "Error updating folder with ID of %d in database", folder.ID)
	}

	if m.inMemoryRepository != nil {
		updatedFolder, errUpdateFolder := m.inMemoryRepository.Update(ctx, folder)
		if errUpdateFolder != nil {
			return nil, errors.Wrapf(errUpdateFolder, "Error updating folder with ID of %d in memory", folder.ID)
		}

		updatedFolderEntry = updatedFolder
	}

	return updatedFolderEntry, nil
}

func (m DatabaseRepository) Create(ctx context.Context, folder Folder) (*Folder, error) {
	var createdFolderEntry = &folder
	folder.CreatedAt = time.Now()
	errCreate := m.conn.Table(TableName).Create(&folder).Error
	if errCreate != nil {
		return nil, errors.Wrapf(errCreate, "Error creating folder with ID of %d in database", folder.ID)
	}

	if m.inMemoryRepository != nil {
		newFolderEntry, errCreateFolder := m.inMemoryRepository.Create(ctx, *createdFolderEntry)
		if errCreateFolder != nil {
			return nil, errors.Wrapf(errCreateFolder, "Error creating folder with parent folder ID of %d and name %s in memory", folder.ParentID, folder.Name)
		}

		createdFolderEntry = newFolderEntry
	}

	return createdFolderEntry, nil
}

func (m DatabaseRepository) Count(ctx context.Context, params ListFolderParams) (uint, error) {
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
		errDelete := m.conn.Table(TableName).Unscoped().Delete(&Folder{}, id).Error
		if errDelete != nil {
			return errors.Wrapf(errDelete, "Error deleting folder with ID of %d in database", id)
		}
	} else {
		errUpdate := m.conn.Table(TableName).Where("id = ?", id).Update("deleted_at",
			sql.NullTime{Valid: true, Time: time.Now()}).Error
		if errUpdate != nil {
			return errors.Wrapf(errUpdate, "Error marking folder with ID of %d deleted in database", id)
		}
	}

	if m.inMemoryRepository != nil {
		errDelete := m.inMemoryRepository.Delete(ctx, id, forceDelete)
		if errDelete != nil {
			return errors.Wrapf(errDelete, "Error deleting folder with ID of %d in memory", id)
		}
	}

	return nil
}

func (m DatabaseRepository) GetQuery(tx *gorm.DB, selectedColumnNames []string, params ListFolderParams) (Query *gorm.DB) {
	conn := m.conn
	if tx != nil {
		conn = tx
	}
	Query = conn.Table(TableName).Select(selectedColumnNames)
	Query = params.DatabaseFilter(TableName, Query)
	return params.DatabaseOrder(TableName, Query, OrderMap)
}
