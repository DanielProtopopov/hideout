package paths

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"hideout/internal/common/model"
	"time"
)

type DatabaseRepository struct {
	conn               *gorm.DB
	inMemoryRepository InMemoryRepository
}

func NewDatabaseRepository(conn *gorm.DB, inMemoryRep InMemoryRepository) DatabaseRepository {
	return DatabaseRepository{conn: conn, inMemoryRepository: inMemoryRep}
}

func (m DatabaseRepository) GetID(ctx context.Context) (uint, error) {
	id := uint(0)
	errScan := m.conn.Table(TableName).Select("COALESCE(MAX(id), 0)").Row().Scan(&id)
	return id + 1, errScan
}

func (m DatabaseRepository) Load(ctx context.Context) ([]Path, error) {
	var results []Path
	errGetRecords := m.conn.Table(TableName).Select([]string{TableName + ".*"}).Find(&results).Error
	if errGetRecords != nil {
		return results, errors.Wrap(errGetRecords, "Failed to obtain records from the database")
	}

	return results, nil
}

func (m DatabaseRepository) GetMapByID(ctx context.Context, params ListPathParams) (map[uint]*Path, error) {
	return m.inMemoryRepository.GetMapByID(ctx, params)
}

func (m DatabaseRepository) GetMapByUID(ctx context.Context, params ListPathParams) (map[string]*Path, error) {
	return m.inMemoryRepository.GetMapByUID(ctx, params)
}

func (m DatabaseRepository) Get(ctx context.Context, params ListPathParams) ([]*Path, error) {
	return m.inMemoryRepository.Get(ctx, params)
}

func (m DatabaseRepository) GetByID(ctx context.Context, id uint) (*Path, error) {
	return m.inMemoryRepository.GetByID(ctx, id)
}

func (m DatabaseRepository) GetByUID(ctx context.Context, uid string) (*Path, error) {
	return m.inMemoryRepository.GetByUID(ctx, uid)
}

func (m DatabaseRepository) Update(ctx context.Context, id uint, name string) (*Path, error) {
	existingPath, errGetPath := m.GetByID(ctx, id)
	if errGetPath != nil {
		return nil, errors.Wrapf(errGetPath, "Failed to retrieve path with ID of %d", id)
	}
	var existingPathName = existingPath.Name

	var updatedPathEntry = Path{Model: model.Model{ID: existingPath.ID, UpdatedAt: time.Now()}, UID: existingPath.UID, Name: name}
	errUpdate := m.conn.Table(TableName).Model(&updatedPathEntry).Update("name", name).Error

	// Revert back to old name
	if errUpdate != nil {
		_, _ = m.inMemoryRepository.Update(ctx, id, existingPathName)
		return nil, errors.Wrapf(errUpdate, "Error updating path with ID of %d", id)
	}

	// Update was successful, update in-memory
	updatedPath, errUpdatePath := m.inMemoryRepository.Update(ctx, id, name)
	if errUpdatePath != nil {
		return nil, errors.Wrapf(errUpdatePath, "Error updating path with ID of %d", id)
	}

	return updatedPath, nil
}

func (m DatabaseRepository) Create(ctx context.Context, id uint, uid string, parentPathID uint, name string) (*Path, error) {
	newPath := Path{Model: model.Model{ID: id, CreatedAt: time.Now()}, ParentID: parentPathID, UID: uid, Name: name}
	errCreate := m.conn.Table(TableName).Create(&newPath).Error
	if errCreate != nil {
		return nil, errors.Wrapf(errCreate, "Error creating path with ID of %d", newPath.ID)
	}

	newPathEntry, errCreatePath := m.inMemoryRepository.Create(ctx, newPath.ID, newPath.UID, newPath.ParentID, newPath.Name)
	if errCreatePath != nil {
		return nil, errors.Wrapf(errCreatePath, "Error creating path with parent path ID of %d and name %s", parentPathID, name)
	}

	return newPathEntry, nil
}

func (m DatabaseRepository) Count(ctx context.Context, name string) (uint, error) {
	return m.inMemoryRepository.Count(ctx, name)
}

func (m DatabaseRepository) Delete(ctx context.Context, id uint, forceDelete bool) error {
	if forceDelete {
		errDelete := m.conn.Table(TableName).Unscoped().Delete(&Path{}, id).Error
		if errDelete != nil {
			return errors.Wrapf(errDelete, "Error deleting path with ID of %d", id)
		}
	} else {
		errUpdate := m.conn.Table(TableName).Where("id = ?", id).Update("deleted_at",
			sql.NullTime{Valid: true, Time: time.Now()}).Error
		if errUpdate != nil {
			return errors.Wrapf(errUpdate, "Error marking path with ID of %d deleted", id)
		}
	}

	return m.inMemoryRepository.Delete(ctx, id, forceDelete)
}
