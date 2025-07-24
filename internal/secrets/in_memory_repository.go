package secrets

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	error2 "hideout/internal/pkg/error"
	pathPkg "path"
	"slices"
	"strings"
)

type InMemoryRepository struct {
	conn []Secret
}

func NewRepository(conn []Secret) Repository {
	return InMemoryRepository{conn: conn}
}

func (m InMemoryRepository) getID() uint {
	id := uint(0)
	for _, secretEntry := range m.conn {
		if secretEntry.ID > id {
			id = secretEntry.ID
		}
	}

	return id + 1
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
	for _, secretEntry := range m.conn {
		if slices.Contains(params.PathIDs, secretEntry.ID) {
			pathResults = append(pathResults, &secretEntry)
		}
	}

	var nameResults []*Secret
	for _, pathResult := range pathResults {
		matched, errPathMatch := pathPkg.Match(params.Name, pathResult.Name)
		if errPathMatch != nil {
			return nil, errPathMatch
		}
		if matched {
			nameResults = append(nameResults, pathResult)
		}
	}

	var typeResults []*Secret
	for _, nameEntry := range nameResults {
		if slices.Index(params.Types, nameEntry.Type) != -1 {
			typeResults = append(typeResults, nameEntry)
		}
	}

	return typeResults, nil
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
	for _, uidSecret := range m.conn {
		if uidSecret.ID == id {
			return &uidSecret, nil
		}
	}

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) GetByUID(ctx context.Context, uid string) (*Secret, error) {
	for _, uidSecret := range m.conn {
		if uidSecret.UID == uid {
			return &uidSecret, nil
		}
	}

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) Update(ctx context.Context, id uint, value string) (*Secret, error) {
	for _, secretEntry := range m.conn {
		if secretEntry.ID == id {
			secretEntry.Value = value
			return &secretEntry, nil
		}
	}

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) Create(ctx context.Context, pathID uint, name string, value string, valueType string) (*Secret, error) {
	for _, secretEntry := range m.conn {
		if secretEntry.Name == name && secretEntry.PathID == pathID {
			return nil, error2.ErrAlreadyExists
		}
	}

	newSecret := Secret{ID: m.getID(), PathID: pathID, UID: gofakeit.UUID(), Name: name, Value: value, Type: valueType}
	m.conn = append(m.conn, newSecret)
	return &newSecret, nil
}

func (m InMemoryRepository) Count(ctx context.Context, pathID uint, name string) (uint, error) {
	totalCount := uint(0)
	for _, secretEntry := range m.conn {
		if pathID == secretEntry.PathID && strings.Contains(secretEntry.Name, name) {
			totalCount++
		}
	}

	return totalCount, nil
}

func (m InMemoryRepository) Delete(ctx context.Context, id uint) error {
	for secretIndex, secretEntry := range m.conn {
		if secretEntry.ID == id {
			m.conn = append(m.conn[:secretIndex], m.conn[secretIndex+1:]...)
			return nil
		}
	}

	return error2.ErrRecordNotFound
}
