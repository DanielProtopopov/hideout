package secrets

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	error2 "hideout/internal/pkg/error"
	"maps"
	pathPkg "path"
	"slices"
	"sort"
	"strings"
)

type InMemoryRepository struct {
	conn map[string]map[string]Secret
}

func NewInMemoryRepository(conn map[string]map[string]Secret) *InMemoryRepository {
	return &InMemoryRepository{conn: conn}
}

func (m InMemoryRepository) GetPaths(ctx context.Context) (paths []string, err error) {
	paths = slices.Collect(maps.Keys(m.conn))
	sort.Slice(paths, func(i, j int) bool {
		s1 := strings.Replace(paths[i], "/", "\x00", -1)
		s2 := strings.Replace(paths[j], "/", "\x00", -1)
		return s1 < s2
	})
	return paths, nil
}

func (m InMemoryRepository) GetMapByID(ctx context.Context, params ListSecretParams) (map[uint]*Secret, error) {
	results, errGetResults := m.GetMapByPath(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	idSecrets := make(map[uint]*Secret)
	for _, secretsByPath := range results {
		for _, secretEntry := range secretsByPath {
			idSecrets[secretEntry.ID] = secretEntry
		}
	}

	return idSecrets, nil
}

func (m InMemoryRepository) GetMapByUID(ctx context.Context, params ListSecretParams) (map[string]*Secret, error) {
	results, errGetResults := m.GetMapByPath(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	uidSecrets := make(map[string]*Secret)
	for _, secretsByPath := range results {
		for _, secretEntry := range secretsByPath {
			uidSecrets[secretEntry.UID] = secretEntry
		}
	}

	return uidSecrets, nil
}

func (m InMemoryRepository) GetMapByPath(ctx context.Context, params ListSecretParams) (map[string][]*Secret, error) {
	pathResults := make(map[string][]*Secret)
	for pathVal, secretsInPath := range m.conn {
		for _, secret := range secretsInPath {
			matched, errPathMatch := pathPkg.Match(params.Path, pathVal)
			if errPathMatch != nil {
				return nil, errPathMatch
			}
			if matched {
				pathResults[pathVal] = append(pathResults[pathVal], &secret)
			}
		}
	}

	nameResults := make(map[string][]*Secret)
	for pathVal, secretsEntry := range pathResults {
		for _, secret := range secretsEntry {
			matched, errPathMatch := pathPkg.Match(params.Name, secret.Name)
			if errPathMatch != nil {
				return nil, errPathMatch
			}
			if matched {
				nameResults[pathVal] = append(nameResults[pathVal], secret)
			}
		}
	}

	typeResults := make(map[string][]*Secret)
	for pathVal, secretsEntry := range pathResults {
		for _, secret := range secretsEntry {
			if slices.Index(params.Types, secret.Type) != -1 {
				typeResults[pathVal] = append(typeResults[pathVal], secret)
			}
		}
	}

	listResults := params.Apply(typeResults)
	orderResults := params.ApplyOrder(listResults)
	return orderResults, nil
}

func (m InMemoryRepository) GetByUID(ctx context.Context, uid string) (Secret, error) {
	for _, uidSecrets := range m.conn {
		for secretUid, secret := range uidSecrets {
			if secretUid == uid {
				return secret, nil
			}
		}
	}

	return Secret{}, error2.ErrRecordNotFound
}

func (m InMemoryRepository) Update(ctx context.Context, uid string, value string) (Secret, error) {
	return Secret{}, nil
}

func (m InMemoryRepository) Create(ctx context.Context, path string, name string, value string) (Secret, error) {
	uidSecrets, pathExists := m.conn[path]
	if !pathExists {
		m.conn[path] = make(map[string]Secret)
	}

	for _, secretEntry := range uidSecrets {
		if secretEntry.Name == name {
			return Secret{}, error2.ErrAlreadyExists
		}
	}

	newSecret := Secret{Name: name, Value: value}
	m.conn[path][gofakeit.UUID()] = newSecret
	return newSecret, nil
}

func (m InMemoryRepository) Count(ctx context.Context, path string, name string) (uint, error) {
	pathSecretsCount := uint(0)
	pathSecrets := make(map[string]map[string]Secret)
	if path != "" {
		for pathVal, secret := range m.conn {
			matched, errPathMatch := pathPkg.Match(path, pathVal)
			if errPathMatch != nil {
				return 0, errPathMatch
			}
			if matched {
				pathSecrets[pathVal] = secret
				pathSecretsCount++
			}
		}
	} else {
		pathSecrets = m.conn
		for _, secretEntries := range m.conn {
			for _ = range secretEntries {
				pathSecretsCount++
			}
		}
	}

	namesCount := uint(0)
	if name != "" {
		for _, secretEntries := range pathSecrets {
			for _, secret := range secretEntries {
				if strings.Contains(secret.Name, name) {
					namesCount++
				}
			}
		}
	} else {
		namesCount = pathSecretsCount
	}

	return namesCount, nil
}

func (m InMemoryRepository) Delete(ctx context.Context, uid string) error {
	_, errGetResult := m.GetByUID(ctx, uid)
	if errGetResult != nil {
		return errGetResult
	}

	for secretPath, _ := range m.conn {
		delete(m.conn[secretPath], uid)
	}

	return nil
}
