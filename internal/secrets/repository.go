package secrets

import (
	"context"
	"hideout/internal/common/secrets"
	error2 "hideout/internal/pkg/error"
)

type Repository struct {
	conn map[string]secrets.Secret
}

func NewRepository(conn map[string]secrets.Secret) *Repository {
	return &Repository{conn: conn}
}

func (m Repository) GetPaths(ctx context.Context) (paths []string, err error) {
	return nil, error2.ErrNotImplemented
}

func (m Repository) Get(ctx context.Context, path string, name string) ([]secrets.Secret, error) {
	return nil, error2.ErrNotImplemented
}

func (m Repository) GetByName(ctx context.Context, name string) (secrets.Secret, error) {
	return secrets.Secret{}, error2.ErrNotImplemented
}

func (m Repository) Update(ctx context.Context, secret *secrets.Secret, value string) (secrets.Secret, error) {
	return secrets.Secret{}, nil
}

func (m Repository) Create(ctx context.Context, secret secrets.Secret, path string, ignoreExisting bool) (secrets.Secret, error) {
	if !ignoreExisting {
		_, exists := m.conn[path]
		if exists {
			return secrets.Secret{}, error2.ErrAlreadyExists
		}
	}

	m.conn[path] = secret
	return secret, nil
}

func (m Repository) Count(ctx context.Context, path string, name string) (uint, error) {
	return 0, error2.ErrNotImplemented
}

// Delete удаление записи
func (m Repository) Delete(ctx context.Context, path string, name string) error {
	return error2.ErrNotImplemented
}
