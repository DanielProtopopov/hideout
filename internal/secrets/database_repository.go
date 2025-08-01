package secrets

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

func (m DatabaseRepository) Load(ctx context.Context) ([]Secret, error) {
	var results []Secret
	errGetRecords := m.conn.Table(TableName).Select([]string{TableName + ".*"}).Find(&results).Error
	if errGetRecords != nil {
		return results, errors.Wrap(errGetRecords, "Failed to obtain records in database")
	}

	return results, nil
}

func (m DatabaseRepository) GetMapByID(ctx context.Context, params ListSecretParams) (map[uint]*Secret, error) {
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

func (m DatabaseRepository) GetMapByUID(ctx context.Context, params ListSecretParams) (map[string]*Secret, error) {
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

func (m DatabaseRepository) Get(ctx context.Context, params ListSecretParams) ([]*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Get(ctx, params)
	}

	var results []*Secret
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

func (m DatabaseRepository) GetMapByPath(ctx context.Context, params ListSecretParams) (map[uint][]*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByPath(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	var mapResults = make(map[uint][]*Secret)
	for _, secret := range results {
		secretsInPath, secretExists := mapResults[secret.PathID]
		if !secretExists {
			secretsInPath = []*Secret{}
		}
		secretsInPath = append(secretsInPath, secret)
		mapResults[secret.PathID] = secretsInPath
	}

	return mapResults, nil
}

func (m DatabaseRepository) GetByID(ctx context.Context, id uint) (*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByID(ctx, id)
	}

	var result Secret
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

func (m DatabaseRepository) GetByUID(ctx context.Context, uid string) (*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByUID(ctx, uid)
	}

	var result Secret
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

func (m DatabaseRepository) Update(ctx context.Context, secret Secret) (*Secret, error) {
	var updatedSecretEntry = &secret
	errUpdate := m.conn.Table(TableName).Model(&secret).Updates(updatedSecretEntry).Error
	if errUpdate != nil {
		return nil, errors.Wrapf(errUpdate, "Error updating secret with ID of %d in database", secret.ID)
	}

	if m.inMemoryRepository != nil {
		updatedSecret, errUpdateSecret := m.inMemoryRepository.Update(ctx, secret)
		if errUpdateSecret != nil {
			return nil, errors.Wrapf(errUpdateSecret, "Error updating secret with ID of %d in memory", secret.ID)
		}

		updatedSecretEntry = updatedSecret
	}

	return updatedSecretEntry, nil
}

func (m DatabaseRepository) Create(ctx context.Context, secret Secret) (*Secret, error) {
	var createdSecretEntry = &secret
	secret.CreatedAt = time.Now()
	errCreate := m.conn.Table(TableName).Create(&secret).Error
	if errCreate != nil {
		return nil, errors.Wrapf(errCreate, "Error creating secret with ID of %d in database", secret.ID)
	}

	if m.inMemoryRepository != nil {
		newSecretEntry, errCreateSecret := m.inMemoryRepository.Create(ctx, *createdSecretEntry)
		if errCreateSecret != nil {
			return nil, errors.Wrapf(errCreateSecret, "Error creating secret with parent path ID of %d and name %s in memory", secret.PathID, secret.Name)
		}

		createdSecretEntry = newSecretEntry
	}

	return createdSecretEntry, nil
}

func (m DatabaseRepository) Count(ctx context.Context, params ListSecretParams) (uint, error) {
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
		errDelete := m.conn.Table(TableName).Unscoped().Delete(&Secret{}, id).Error
		if errDelete != nil {
			return errors.Wrapf(errDelete, "Error deleting secret with ID of %d in database", id)
		}
	} else {
		errUpdate := m.conn.Table(TableName).Where("id = ?", id).Update("deleted_at",
			sql.NullTime{Valid: true, Time: time.Now()}).Error
		if errUpdate != nil {
			return errors.Wrapf(errUpdate, "Error marking secret with ID of %d deleted in database", id)
		}
	}

	if m.inMemoryRepository != nil {
		errDelete := m.inMemoryRepository.Delete(ctx, id, forceDelete)
		if errDelete != nil {
			return errors.Wrapf(errDelete, "Error deleting secret with ID of %d in memory", id)
		}
	}

	return nil
}

func (m DatabaseRepository) GetQuery(tx *gorm.DB, selectedColumnNames []string, params ListSecretParams) (Query *gorm.DB) {
	conn := m.conn
	if tx != nil {
		conn = tx
	}
	Query = conn.Table(TableName).Select(selectedColumnNames)
	Query = params.DatabaseFilter(TableName, Query)
	return params.DatabaseOrder(TableName, Query, OrderMap)
}
