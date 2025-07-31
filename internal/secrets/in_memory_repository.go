package secrets

import (
	"context"
	"database/sql"
	"hideout/internal/common/apperror"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/common/ordering"
	"hideout/internal/common/pagination"
	pathPkg "path"
	"slices"
	"time"
)

type InMemoryRepository struct {
	conn *[]Secret
}

func NewInMemoryRepository(conn *[]Secret) InMemoryRepository {
	return InMemoryRepository{conn: conn}
}

func (m InMemoryRepository) GetID(ctx context.Context) (uint, error) {
	id := uint(0)
	for _, secretEntry := range *m.conn {
		if secretEntry.ID > id {
			id = secretEntry.ID
		}
	}

	return id + 1, nil
}

func (m InMemoryRepository) GetMapByID(ctx context.Context, params ListSecretParams) (map[uint]*Secret, error) {
	secrets, errGetSecrets := m.Get(ctx, params)
	if errGetSecrets != nil {
		return nil, errGetSecrets
	}
	results := make(map[uint]*Secret)
	for _, secret := range secrets {
		results[secret.ID] = secret
	}

	return results, nil
}

func (m InMemoryRepository) GetMapByUID(ctx context.Context, params ListSecretParams) (map[string]*Secret, error) {
	secrets, errGetSecrets := m.Get(ctx, params)
	if errGetSecrets != nil {
		return nil, errGetSecrets
	}
	results := make(map[string]*Secret)
	for _, secret := range secrets {
		results[secret.UID] = secret
	}

	return results, nil
}

func (m InMemoryRepository) Get(ctx context.Context, params ListSecretParams) ([]*Secret, error) {
	var pathResults []*Secret
	for _, secretEntry := range *m.conn {
		if len(params.PathIDs) > 0 {
			if slices.Contains(params.PathIDs, secretEntry.ID) {
				pathResults = append(pathResults, &secretEntry)
			}
		} else {
			pathResults = append(pathResults, &secretEntry)
		}
	}

	var nameResults []*Secret
	for _, pathResult := range pathResults {
		if params.Name != "" {
			matched, errPathMatch := pathPkg.Match(params.Name, pathResult.Name)
			if errPathMatch != nil {
				return nil, errPathMatch
			}
			if matched {
				nameResults = append(nameResults, pathResult)
			}
		} else {
			nameResults = append(nameResults, pathResult)
		}
	}

	var typeResults []*Secret
	for _, nameEntry := range nameResults {
		if len(params.Types) > 0 {
			if slices.Index(params.Types, nameEntry.Type) != -1 {
				typeResults = append(typeResults, nameEntry)
			}
		} else {
			typeResults = append(typeResults, nameEntry)
		}
	}

	filteredResults := m.Filter(ctx, typeResults, params.ListParams)
	if params.Page == 0 && params.PerPage == 0 {
		return filteredResults, nil
	}

	offset, length := pagination.Paginate(len(filteredResults), int(params.Page*params.PerPage), int(params.PerPage))
	return filteredResults[offset:length], nil
}

func (m InMemoryRepository) GetMapByPath(ctx context.Context, params ListSecretParams) (map[uint][]*Secret, error) {
	secrets, errGetSecrets := m.Get(ctx, params)
	if errGetSecrets != nil {
		return nil, errGetSecrets
	}
	results := make(map[uint][]*Secret)
	for _, secret := range secrets {
		secretsInPath, secretExists := results[secret.PathID]
		if !secretExists {
			secretsInPath = []*Secret{}
		}
		secretsInPath = append(secretsInPath, secret)
		results[secret.PathID] = secretsInPath
	}

	return results, nil
}

func (m InMemoryRepository) GetByID(ctx context.Context, id uint) (*Secret, error) {
	for _, secretEntry := range *m.conn {
		if secretEntry.ID == id && !secretEntry.DeletedAt.Valid {
			return &secretEntry, nil
		}
	}

	return nil, apperror.ErrRecordNotFound
}

func (m InMemoryRepository) GetByUID(ctx context.Context, uid string) (*Secret, error) {
	for _, secretEntry := range *m.conn {
		if secretEntry.UID == uid && !secretEntry.DeletedAt.Valid {
			return &secretEntry, nil
		}
	}

	return nil, apperror.ErrRecordNotFound
}

func (m InMemoryRepository) Update(ctx context.Context, secret Secret) (*Secret, error) {
	for _, secretEntry := range *m.conn {
		if !secretEntry.DeletedAt.Valid && secretEntry.ID == secret.ID {
			secretEntry.PathID = secret.PathID
			secretEntry.Name = secret.Name
			secretEntry.Value = secret.Value
			secretEntry.Type = secret.Type
			secretEntry.UpdatedAt = time.Now()
			return &secretEntry, nil
		}
	}

	return nil, apperror.ErrRecordNotFound
}

func (m InMemoryRepository) Create(ctx context.Context, secret Secret) (*Secret, error) {
	for _, secretEntry := range *m.conn {
		if !secretEntry.DeletedAt.Valid && secretEntry.Name == secret.Name && secretEntry.PathID == secret.PathID {
			return nil, apperror.ErrAlreadyExists
		}
	}

	secret.CreatedAt = time.Now()
	*m.conn = append(*m.conn, secret)
	return &secret, nil
}

func (m InMemoryRepository) Count(ctx context.Context, params ListSecretParams) (uint, error) {
	// These are not needed when performing filtering and counting
	params.Pagination = pagination.Pagination{PerPage: 0, Page: 0}
	params.Order = []ordering.OrderRQ{}
	secretsList, errGetSecrets := m.Get(ctx, params)
	if errGetSecrets != nil {
		return 0, errGetSecrets
	}

	return uint(len(secretsList)), nil
}

func (m InMemoryRepository) Delete(ctx context.Context, id uint, forceDelete bool) error {
	for secretIndex, secretEntry := range *m.conn {
		if secretEntry.ID == id {
			if forceDelete {
				*m.conn = slices.Delete(*m.conn, secretIndex, secretIndex+1)
			} else {
				secretEntry.DeletedAt = sql.NullTime{Valid: true, Time: time.Now()}
			}
			return nil
		}
	}

	return apperror.ErrRecordNotFound
}

func (m InMemoryRepository) Filter(ctx context.Context, data []*Secret, params generics.ListParams) []*Secret {
	var idResults []*Secret
	for _, pathEntry := range data {
		if len(params.IDs) > 0 {
			if slices.Contains(params.IDs, pathEntry.ID) {
				idResults = append(idResults, pathEntry)
			}
		} else {
			idResults = append(idResults, pathEntry)
		}
	}

	var uidResults []*Secret
	for _, pathEntry := range idResults {
		if len(params.UIDs) > 0 {
			if slices.Contains(params.UIDs, pathEntry.UID) {
				uidResults = append(uidResults, pathEntry)
			}
		} else {
			uidResults = append(uidResults, pathEntry)
		}
	}

	var softDeletedResults []*Secret
	for _, pathEntry := range uidResults {
		if params.Deleted == model.Yes {
			if pathEntry.DeletedAt.Valid {
				softDeletedResults = append(softDeletedResults, pathEntry)
			}
		} else if params.Deleted == model.No {
			if !pathEntry.DeletedAt.Valid {
				softDeletedResults = append(softDeletedResults, pathEntry)
			}
		} else {
			softDeletedResults = append(softDeletedResults, pathEntry)
		}
	}

	return m.Sort(ctx, softDeletedResults, params.Order)
}

func (m InMemoryRepository) Sort(ctx context.Context, data []*Secret, ordering []ordering.OrderRQ) []*Secret {
	var orderParams []lessFunc
	for _, order := range ordering {
		columnMap, _ := OrderMap[order.OrderBy]
		switch columnMap {
		case "id":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Secret) bool {
						if order.Order {
							return p1.ID < p2.ID
						} else {
							return p1.ID > p2.ID
						}
					})
				}
			}
		case "path_id":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Secret) bool {
						if order.Order {
							return p1.PathID < p2.PathID
						} else {
							return p1.PathID > p2.PathID
						}
					})
				}
			}
		case "uid":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Secret) bool {
						if order.Order {
							return p1.UID < p2.UID
						} else {
							return p1.UID > p2.UID
						}
					})
				}
			}
		case "name":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Secret) bool {
						if order.Order {
							return p1.Name < p2.Name
						} else {
							return p1.Name > p2.Name
						}
					})
				}
			}
		case "type":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Secret) bool {
						if order.Order {
							return p1.Type < p2.Type
						} else {
							return p1.Type > p2.Type
						}
					})
				}
			}
		case "created_at":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Secret) bool {
						if order.Order {
							return p1.CreatedAt.Before(p2.CreatedAt)
						} else {
							return p1.CreatedAt.After(p2.CreatedAt)
						}
					})
				}
			}
		case "updated_at":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Secret) bool {
						if order.Order {
							return p1.UpdatedAt.Before(p2.UpdatedAt)
						} else {
							return p1.UpdatedAt.After(p2.UpdatedAt)
						}
					})
				}
			}
		case "deleted_at":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Secret) bool {
						if order.Order {
							return p1.DeletedAt.Valid && p2.DeletedAt.Valid && p1.DeletedAt.Time.Before(p2.DeletedAt.Time)
						} else {
							return p1.DeletedAt.Valid && p2.DeletedAt.Valid && p1.DeletedAt.Time.After(p2.DeletedAt.Time)
						}
					})
				}
			}
		}
	}

	OrderedBy(orderParams...).Sort(data)
	return data
}
