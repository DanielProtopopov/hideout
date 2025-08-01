package secrets

import (
	"context"
	"hideout/internal/folders"
)

func (m *SecretsService) GetFolders(ctx context.Context, params folders.ListFolderParams) ([]*folders.Folder, error) {
	return m.foldersRepository.Get(ctx, params)
}

func (m *SecretsService) GetFoldersMapByID(ctx context.Context, params folders.ListFolderParams) (map[uint]*folders.Folder, error) {
	return m.foldersRepository.GetMapByID(ctx, params)
}

func (m *SecretsService) GetFoldersMapByUID(ctx context.Context, params folders.ListFolderParams) (map[string]*folders.Folder, error) {
	return m.foldersRepository.GetMapByUID(ctx, params)
}

func (m *SecretsService) GetFolderByUID(ctx context.Context, uid string) (*folders.Folder, error) {
	return m.foldersRepository.GetByUID(ctx, uid)
}

func (m *SecretsService) GetFolderByID(ctx context.Context, id uint) (*folders.Folder, error) {
	return m.foldersRepository.GetByID(ctx, id)
}

func (m *SecretsService) UpdateFolder(ctx context.Context, folder folders.Folder) (*folders.Folder, error) {
	return m.foldersRepository.Update(ctx, folder)
}

func (m *SecretsService) CreateFolder(ctx context.Context, folder folders.Folder) (*folders.Folder, error) {
	return m.foldersRepository.Create(ctx, folder)
}

func (m *SecretsService) DeleteFolder(ctx context.Context, id uint, forceDelete bool) error {
	return m.foldersRepository.Delete(ctx, id, forceDelete)
}

func (m *SecretsService) CountFolders(ctx context.Context, params folders.ListFolderParams) (uint, error) {
	return m.foldersRepository.Count(ctx, params)
}
