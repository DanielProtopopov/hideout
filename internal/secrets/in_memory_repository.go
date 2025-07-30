package secrets

import (
	"context"
	"database/sql"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	error2 "hideout/internal/pkg/error"
	pathPkg "path"
	"slices"
	"strings"
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

	return m.Filter(ctx, typeResults, params.ListParams)
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

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) GetByUID(ctx context.Context, uid string) (*Secret, error) {
	for _, secretEntry := range *m.conn {
		if secretEntry.UID == uid && !secretEntry.DeletedAt.Valid {
			return &secretEntry, nil
		}
	}

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) Update(ctx context.Context, id uint, value string) (*Secret, error) {
	for _, secretEntry := range *m.conn {
		if secretEntry.ID == id && !secretEntry.DeletedAt.Valid {
			secretEntry.Value = value
			return &secretEntry, nil
		}
	}

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) Create(ctx context.Context, id uint, uid string, pathID uint, name string, value string, valueType string) (*Secret, error) {
	for _, secretEntry := range *m.conn {
		if secretEntry.Name == name && secretEntry.PathID == pathID && !secretEntry.DeletedAt.Valid {
			return nil, error2.ErrAlreadyExists
		}
	}

	newSecret := Secret{Model: model.Model{ID: id}, UID: uid, PathID: pathID, Name: name, Value: value, Type: valueType}
	*m.conn = append(*m.conn, newSecret)
	return &newSecret, nil
}

func (m InMemoryRepository) Count(ctx context.Context, pathID uint, name string) (uint, error) {
	totalCount := uint(0)
	for _, secretEntry := range *m.conn {
		if pathID == secretEntry.PathID && strings.Contains(secretEntry.Name, name) && !secretEntry.DeletedAt.Valid {
			totalCount++
		}
	}

	return totalCount, nil
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

	return error2.ErrRecordNotFound
}

func (m InMemoryRepository) Filter(ctx context.Context, results []*Secret, params generics.ListParams) ([]*Secret, error) {
	var idResults []*Secret
	for _, pathEntry := range results {
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

	return uidResults, nil
}
