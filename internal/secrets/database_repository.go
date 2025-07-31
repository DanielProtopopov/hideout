package secrets

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

func (m DatabaseRepository) Load(ctx context.Context) ([]Secret, error) {
	var results []Secret
	errGetRecords := m.conn.Table(TableName).Select([]string{TableName + ".*"}).Find(&results).Error
	if errGetRecords != nil {
		return results, errors.Wrap(errGetRecords, "Failed to obtain records from the database")
	}

	return results, nil
}

func (m DatabaseRepository) GetMapByID(ctx context.Context, params ListSecretParams) (map[uint]*Secret, error) {
	return m.inMemoryRepository.GetMapByID(ctx, params)
}

func (m DatabaseRepository) GetMapByUID(ctx context.Context, params ListSecretParams) (map[string]*Secret, error) {
	return m.inMemoryRepository.GetMapByUID(ctx, params)
}

func (m DatabaseRepository) Get(ctx context.Context, params ListSecretParams) ([]*Secret, error) {
	return m.inMemoryRepository.Get(ctx, params)
}

func (m DatabaseRepository) GetMapByPath(ctx context.Context, params ListSecretParams) (map[uint][]*Secret, error) {
	return m.inMemoryRepository.GetMapByPath(ctx, params)
}

func (m DatabaseRepository) GetByID(ctx context.Context, id uint) (*Secret, error) {
	return m.inMemoryRepository.GetByID(ctx, id)
}

func (m DatabaseRepository) GetByUID(ctx context.Context, uid string) (*Secret, error) {
	return m.inMemoryRepository.GetByUID(ctx, uid)
}

func (m DatabaseRepository) Update(ctx context.Context, id uint, value string) (*Secret, error) {
	existingSecret, errGetSecret := m.GetByID(ctx, id)
	if errGetSecret != nil {
		return nil, errors.Wrapf(errGetSecret, "Failed to retrieve secret with ID of %d", id)
	}

	var updatedSecretEntry = Secret{Model: model.Model{ID: existingSecret.ID}, UID: existingSecret.UID, Value: value}
	errUpdate := m.conn.Table(TableName).Model(&updatedSecretEntry).Update("value", value).Error

	// Revert back to old name
	if errUpdate != nil {
		return nil, errors.Wrapf(errUpdate, "Error updating secret with ID of %d", id)
	}

	// Update was successful, update in-memory
	updatedSecret, errUpdateSecret := m.inMemoryRepository.Update(ctx, id, value)
	if errUpdateSecret != nil {
		return nil, errors.Wrapf(errUpdateSecret, "Error updating secret with ID of %d", id)
	}

	return updatedSecret, nil
}

func (m DatabaseRepository) Create(ctx context.Context, id uint, uid string, pathID uint, name string, value string, valueType string) (*Secret, error) {
	newSecret := Secret{Model: model.Model{ID: id}, PathID: pathID, UID: uid, Name: name, Value: value, Type: valueType}
	errCreate := m.conn.Table(TableName).Create(&newSecret).Error
	if errCreate != nil {
		return nil, errors.Wrapf(errCreate, "Error creating secret with ID of %d", newSecret.ID)
	}

	newSecretEntry, errCreateSecret := m.inMemoryRepository.Create(ctx, newSecret.ID, newSecret.UID, newSecret.PathID,
		newSecret.Name, newSecret.Value, newSecret.Type)
	if errCreateSecret != nil {
		return nil, errors.Wrapf(errCreateSecret, "Error creating secret with path ID of %d and name %s", pathID, name)
	}
	return newSecretEntry, nil
}

func (m DatabaseRepository) Count(ctx context.Context, params ListSecretParams) (uint, error) {
	return m.inMemoryRepository.Count(ctx, params)
}

func (m DatabaseRepository) Delete(ctx context.Context, id uint, forceDelete bool) error {
	if forceDelete {
		errDelete := m.conn.Table(TableName).Unscoped().Delete(&Secret{}, id).Error
		if errDelete != nil {
			return errors.Wrapf(errDelete, "Error deleting secret with ID of %d", id)
		}
	} else {
		errUpdate := m.conn.Table(TableName).Where("id = ?", id).Update("deleted_at",
			sql.NullTime{Valid: true, Time: time.Now()}).Error
		if errUpdate != nil {
			return errors.Wrapf(errUpdate, "Error marking secret with ID of %d deleted", id)
		}
	}

	return m.inMemoryRepository.Delete(ctx, id, forceDelete)
}
