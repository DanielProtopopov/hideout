package secrets

import (
	"context"
	"hideout/internal/paths"
)

func (m *SecretsService) GetPaths(ctx context.Context, params paths.ListPathParams) ([]*paths.Path, error) {
	return m.pathsRepository.Get(ctx, params)
}

func (m *SecretsService) GetPathsMapByID(ctx context.Context, params paths.ListPathParams) (map[uint]*paths.Path, error) {
	return m.pathsRepository.GetMapByID(ctx, params)
}

func (m *SecretsService) GetPathsMapByUID(ctx context.Context, params paths.ListPathParams) (map[string]*paths.Path, error) {
	return m.pathsRepository.GetMapByUID(ctx, params)
}

func (m *SecretsService) GetPathByUID(ctx context.Context, uid string) (*paths.Path, error) {
	return m.pathsRepository.GetByUID(ctx, uid)
}

func (m *SecretsService) GetPathByID(ctx context.Context, id uint) (*paths.Path, error) {
	return m.pathsRepository.GetByID(ctx, id)
}

func (m *SecretsService) UpdatePath(ctx context.Context, path paths.Path) (*paths.Path, error) {
	return m.pathsRepository.Update(ctx, path)
}

func (m *SecretsService) CreatePath(ctx context.Context, path paths.Path) (*paths.Path, error) {
	return m.pathsRepository.Create(ctx, path)
}

func (m *SecretsService) DeletePath(ctx context.Context, id uint, forceDelete bool) error {
	return m.pathsRepository.Delete(ctx, id, forceDelete)
}

func (m *SecretsService) CountPaths(ctx context.Context, params paths.ListPathParams) (uint, error) {
	return m.pathsRepository.Count(ctx, params)
}
