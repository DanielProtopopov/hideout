package secrets

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	"hideout/internal/common/secrets"
	error2 "hideout/internal/pkg/error"
	"maps"
	pathPkg "path"
	"slices"
	"sort"
	"strings"
)

type Repository struct {
	conn map[string]map[string]secrets.Secret
}

func NewRepository(conn map[string]map[string]secrets.Secret) *Repository {
	return &Repository{conn: conn}
}

func (m Repository) GetPaths(ctx context.Context) (paths []string, err error) {
	paths = slices.Collect(maps.Keys(m.conn))
	sort.Slice(paths, func(i, j int) bool {
		s1 := strings.Replace(paths[i], "/", "\x00", -1)
		s2 := strings.Replace(paths[j], "/", "\x00", -1)
		return s1 < s2
	})
	return paths, nil
}

func (m Repository) Get(ctx context.Context, path string, name string) (map[string]secrets.Secret, error) {
	pathSecrets := make(map[string]map[string]secrets.Secret)
	for pathVal, secret := range m.conn {
		matched, errPathMatch := pathPkg.Match(path, pathVal)
		if errPathMatch != nil {
			return nil, errPathMatch
		}
		if matched {
			pathSecrets[pathVal] = secret
		}
	}

	nameSecrets := make(map[string]secrets.Secret)
	for pathVal, secretsEntry := range pathSecrets {
		for _, secret := range secretsEntry {
			if strings.Contains(secret.Name, name) {
				nameSecrets[pathVal] = secret
			}
		}
	}

	return nameSecrets, nil
}

func (m Repository) GetByUID(ctx context.Context, uid string) (secrets.Secret, error) {
	for _, uidSecrets := range m.conn {
		for secretUid, secret := range uidSecrets {
			if secretUid == uid {
				return secret, nil
			}
		}
	}

	return secrets.Secret{}, error2.ErrRecordNotFound
}

func (m Repository) Update(ctx context.Context, uid string, value string) (secrets.Secret, error) {
	return secrets.Secret{}, nil
}

func (m Repository) Create(ctx context.Context, path string, name string, value string) (secrets.Secret, error) {
	uidSecrets, pathExists := m.conn[path]
	if !pathExists {
		m.conn[path] = make(map[string]secrets.Secret)
	}

	for _, secretEntry := range uidSecrets {
		if secretEntry.Name == name {
			return secrets.Secret{}, error2.ErrAlreadyExists
		}
	}

	newSecret := secrets.Secret{Name: name, Value: value}
	m.conn[path][gofakeit.UUID()] = newSecret
	return newSecret, nil
}

func (m Repository) Count(ctx context.Context, path string, name string) (uint, error) {
	pathSecretsCount := uint(0)
	pathSecrets := make(map[string]map[string]secrets.Secret)
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

func (m Repository) Delete(ctx context.Context, uid string) error {
	_, errGetResult := m.GetByUID(ctx, uid)
	if errGetResult != nil {
		return errGetResult
	}

	for secretPath, _ := range m.conn {
		delete(m.conn[secretPath], uid)
	}

	return nil
}
