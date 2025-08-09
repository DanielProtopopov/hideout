package secrets

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/pkg/errors"
	"hideout/config"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/folders"
	"hideout/internal/secrets"
	"hideout/structs"
)

type Config struct {
}

type SecretsService struct {
	secretsConfig config.RepositoryConfig
	foldersConfig config.RepositoryConfig
	folders       *[]folders.Folder
	secrets       *[]secrets.Secret

	secretsRepository secrets.Repository
	foldersRepository folders.Repository
}

// NewService Creation of the service
func NewService(ctx context.Context, secretsConfig config.RepositoryConfig, foldersConfig config.RepositoryConfig, foldersList *[]folders.Folder, secretsList *[]secrets.Secret) (*SecretsService, error) {
	secretsService := &SecretsService{secretsConfig: secretsConfig, foldersConfig: foldersConfig}
	switch secretsConfig.Type {
	case RepositoryType_InMemory:
		{
			secretsService.secretsRepository = secrets.NewInMemoryRepository(&structs.Secrets)
			secretsService.secrets = secretsList
		}
	case RepositoryType_Redis:
		{
			var inMemorySecretsRep *secrets.InMemoryRepository = nil
			if secretsConfig.PreloadInMemory {
				inMemorySecretsRep = secrets.NewInMemoryRepository(&structs.Secrets)
			}
			redisSecretsRep := secrets.NewRedisRepository(structs.Redis, inMemorySecretsRep)
			if secretsConfig.PreloadInMemory {
				loadedSecrets, errLoadSecrets := redisSecretsRep.Load(ctx)
				if errLoadSecrets != nil {
					return nil, errors.Wrap(errLoadSecrets, "Error preloading secrets from Redis storage")
				}
				secretsService.secrets = &loadedSecrets
				secretsService.secretsRepository = redisSecretsRep
			}

			if secretsConfig.PreloadInMemory {
				errLoad := secretsService.LoadSecrets(ctx)
				if errLoad != nil {
					return nil, errors.Wrap(errLoad, "Error loading data into memory")
				}
			}
		}
	case RepositoryType_Database:
		{
			var inMemorySecretsRep *secrets.InMemoryRepository = nil
			if secretsConfig.PreloadInMemory {
				inMemorySecretsRep = secrets.NewInMemoryRepository(&structs.Secrets)
			}
			databaseSecretsRep := secrets.NewDatabaseRepository(structs.Gorm, inMemorySecretsRep)
			secretsService.secretsRepository = databaseSecretsRep

			if secretsConfig.PreloadInMemory {
				errLoad := secretsService.LoadSecrets(ctx)
				if errLoad != nil {
					return nil, errors.Wrap(errLoad, "Error loading data into memory")
				}
			}
		}
	case RepositoryType_File:
		{
			var inMemorySecretsRep *secrets.InMemoryRepository = nil
			if secretsConfig.PreloadInMemory {
				inMemorySecretsRep = secrets.NewInMemoryRepository(&structs.Secrets)
			}
			fileSecretsRep := secrets.NewFileRepository(secretsConfig.FileName, secretsConfig.FileEncoding, inMemorySecretsRep)
			secretsService.secretsRepository = fileSecretsRep

			if secretsConfig.PreloadInMemory {
				errLoad := secretsService.LoadSecrets(ctx)
				if errLoad != nil {
					return nil, errors.Wrap(errLoad, "Error loading data into memory")
				}
			}
		}
	}

	switch foldersConfig.Type {
	case RepositoryType_InMemory:
		{
			secretsService.foldersRepository = folders.NewInMemoryRepository(&structs.Folders)
			secretsService.folders = foldersList
		}
	case RepositoryType_Redis:
		{
			var inMemoryFoldersRep *folders.InMemoryRepository = nil
			if foldersConfig.PreloadInMemory {
				inMemoryFoldersRep = folders.NewInMemoryRepository(&structs.Folders)
			}
			redisFoldersRep := folders.NewRedisRepository(structs.Redis, inMemoryFoldersRep)
			if foldersConfig.PreloadInMemory {
				loadedFolders, errLoadFolders := redisFoldersRep.Load(ctx)
				if errLoadFolders != nil {
					return nil, errors.Wrap(errLoadFolders, "Error preloading folders into in-memory storage in Redis")
				}
				secretsService.folders = &loadedFolders
				secretsService.foldersRepository = redisFoldersRep
			}

			if foldersConfig.PreloadInMemory {
				errLoad := secretsService.LoadFolders(ctx)
				if errLoad != nil {
					return nil, errors.Wrap(errLoad, "Error loading data into memory")
				}
			}
		}
	case RepositoryType_Database:
		{
			var inMemoryFoldersRep *folders.InMemoryRepository = nil
			if foldersConfig.PreloadInMemory {
				inMemoryFoldersRep = folders.NewInMemoryRepository(&structs.Folders)
			}
			databaseFoldersRep := folders.NewDatabaseRepository(structs.Gorm, inMemoryFoldersRep)
			secretsService.foldersRepository = databaseFoldersRep

			if foldersConfig.PreloadInMemory {
				errLoad := secretsService.LoadFolders(ctx)
				if errLoad != nil {
					return nil, errors.Wrap(errLoad, "Error loading data into memory")
				}
			}
		}
	case RepositoryType_File:
		{
			var inMemoryFoldersRep *folders.InMemoryRepository = nil
			if foldersConfig.PreloadInMemory {
				inMemoryFoldersRep = folders.NewInMemoryRepository(&structs.Folders)
			}
			fileFoldersRep := folders.NewFileRepository(foldersConfig.FileName, foldersConfig.FileEncoding, inMemoryFoldersRep)
			secretsService.foldersRepository = fileFoldersRep

			if foldersConfig.PreloadInMemory {
				errLoad := secretsService.LoadFolders(ctx)
				if errLoad != nil {
					return nil, errors.Wrap(errLoad, "Error loading data into memory")
				}
			}
		}
	}

	return secretsService, nil
}

func (m *SecretsService) Load(ctx context.Context) error {
	errLoadSecrets := m.LoadSecrets(ctx)
	if errLoadSecrets != nil {
		return errLoadSecrets
	}

	errLoadFolders := m.LoadFolders(ctx)
	if errLoadFolders != nil {
		return errLoadFolders
	}

	return nil
}

func (m *SecretsService) LoadSecrets(ctx context.Context) error {
	if m.secretsConfig.PreloadInMemory {
		loadedSecrets, errLoadSecrets := m.secretsRepository.Load(ctx)
		if errLoadSecrets != nil {
			return errors.Wrap(errLoadSecrets, "Error preloading secrets into in-memory storage")
		}
		structs.Secrets = loadedSecrets
	}

	return nil
}

func (m *SecretsService) LoadFolders(ctx context.Context) error {
	if m.foldersConfig.PreloadInMemory {
		loadedFolders, errLoadFolders := m.foldersRepository.Load(ctx)
		if errLoadFolders != nil {
			return errors.Wrap(errLoadFolders, "Error preloading folders into in-memory storage in Redis")
		}
		structs.Folders = loadedFolders
	}

	return nil
}

func (m *SecretsService) Tree(ctx context.Context, folderID uint) (TreeNode, error) {
	result := TreeNode{Name: "", Type: "Folder", Children: nil}
	existingFolder, errGetFolder := m.foldersRepository.GetByID(ctx, folderID)
	if errGetFolder != nil {
		return result, errGetFolder
	}
	result.Name = existingFolder.Name

	existingFolderFolders, errGetExistingFolderFolders := m.getFoldersByFolder(ctx, existingFolder.ID)
	if errGetExistingFolderFolders != nil {
		return result, errGetExistingFolderFolders
	}
	existingFolderSecrets, errGetExistingFolderSecrets := m.getSecretsByFolder(ctx, existingFolder.ID)
	if errGetExistingFolderSecrets != nil {
		return result, errGetExistingFolderSecrets
	}

	for _, existingFolderSecret := range existingFolderSecrets {
		result.Children = append(result.Children, TreeNode{Name: existingFolderSecret.Name, Type: "Secret"})
	}

	for _, existingFolderFolder := range existingFolderFolders {
		folderNode, errGetFolderNode := m.Tree(ctx, existingFolderFolder.ID)
		if errGetFolderNode != nil {
			return result, errGetFolderNode
		}
		result.Children = append(result.Children, folderNode)
	}

	return result, nil
}

func (m *SecretsService) Delete(ctx context.Context, existingFolders []*folders.Folder, existingSecrets []*secrets.Secret, folderIDFrom uint, forceDelete bool) ([]*folders.Folder, []*secrets.Secret, error) {
	var deletedFolders []*folders.Folder
	var deletedSecrets []*secrets.Secret

	_, errGetFolderFrom := m.foldersRepository.GetByID(ctx, folderIDFrom)
	if errGetFolderFrom != nil {
		return nil, nil, errGetFolderFrom
	}

	// This deletes secrets From designed folder
	for _, existingSecret := range existingSecrets {
		errDeleteSecret := m.secretsRepository.Delete(ctx, existingSecret.ID, forceDelete)
		if errDeleteSecret != nil {
			return nil, nil, errDeleteSecret
		}
		deletedSecrets = append(deletedSecrets, existingSecret)
	}

	// This recursively deletes folders and their secrets from sub-folders
	for _, existingFolder := range existingFolders {
		existingFolderFolders, errGetExistingFolderFolders := m.getFoldersByFolder(ctx, existingFolder.ID)
		if errGetExistingFolderFolders != nil {
			return nil, nil, errGetExistingFolderFolders
		}
		existingFolderSecrets, errGetExistingFolderSecrets := m.getSecretsByFolder(ctx, existingFolder.ID)
		if errGetExistingFolderSecrets != nil {
			return nil, nil, errGetExistingFolderSecrets
		}

		// Delete folders & secrets from an existing folder
		deletedFolderFolders, deletedFolderSecrets, errDelete := m.Delete(ctx, existingFolderFolders, existingFolderSecrets, existingFolder.ID, forceDelete)
		if errDelete != nil {
			return nil, nil, errDelete
		}

		errDeleteFolder := m.foldersRepository.Delete(ctx, existingFolder.ID, forceDelete)
		if errDeleteFolder != nil {
			return nil, nil, errDeleteFolder
		}
		deletedFolders = append(deletedFolders, existingFolder)

		deletedFolders = append(deletedFolders, deletedFolderFolders...)
		deletedSecrets = append(deletedSecrets, deletedFolderSecrets...)
	}

	return deletedFolders, deletedSecrets, nil
}

func (m *SecretsService) Copy(ctx context.Context, existingFolders []*folders.Folder, existingSecrets []*secrets.Secret, folderIDFrom uint, folderIDTo uint) ([]*folders.Folder, []*secrets.Secret, error) {
	var newSecrets []*secrets.Secret
	var newFolders []*folders.Folder

	_, errGetFolderFrom := m.foldersRepository.GetByID(ctx, folderIDFrom)
	if errGetFolderFrom != nil {
		return nil, nil, errGetFolderFrom
	}
	_, errGetFolderTo := m.foldersRepository.GetByID(ctx, folderIDTo)
	if errGetFolderTo != nil {
		return nil, nil, errGetFolderTo
	}

	// This copies secrets From designed folder To target folder
	copiedSecretsMap, errCopySecrets := m.copySecrets(ctx, existingSecrets, folderIDFrom, folderIDTo)
	if errCopySecrets != nil {
		return nil, nil, errCopySecrets
	}
	for _, copiedSecret := range copiedSecretsMap {
		newSecrets = append(newSecrets, copiedSecret)
	}

	// This copies folders From designed folder To target folder
	copiedFoldersMap, errCopyFolders := m.copyFolders(ctx, existingFolders, folderIDFrom, folderIDTo)
	if errCopyFolders != nil {
		return nil, nil, errCopyFolders
	}
	for _, copiedFolder := range copiedFoldersMap {
		newFolders = append(newFolders, copiedFolder)
	}

	// This recursively copies folders and their secrets from folders in From folder
	for _, existingFolder := range existingFolders {
		copiedFolder, _ := copiedFoldersMap[existingFolder.ID]
		existingFolderFolders, errGetExistingFolderFolders := m.getFoldersByFolder(ctx, existingFolder.ID)
		if errGetExistingFolderFolders != nil {
			return nil, nil, errGetExistingFolderFolders
		}
		existingFolderSecrets, errGetExistingFolderSecrets := m.getSecretsByFolder(ctx, existingFolder.ID)
		if errGetExistingFolderSecrets != nil {
			return nil, nil, errGetExistingFolderSecrets
		}

		// Copy folders & secrets from an existing folder to a copied folder
		createdFolders, createdSecrets, errCopy := m.Copy(ctx, existingFolderFolders, existingFolderSecrets, existingFolder.ID, copiedFolder.ID)
		if errCopy != nil {
			return nil, nil, errCopy
		}
		newFolders = append(newFolders, createdFolders...)
		newSecrets = append(newSecrets, createdSecrets...)
	}

	return newFolders, newSecrets, nil
}

func (m *SecretsService) copyFolders(ctx context.Context, foldersList []*folders.Folder, folderIDFrom uint, folderIDTo uint) (map[uint]*folders.Folder, error) {
	results := make(map[uint]*folders.Folder)
	_, errGetFromFolder := m.foldersRepository.GetByID(ctx, folderIDFrom)
	if errGetFromFolder != nil {
		return nil, errGetFromFolder
	}
	toFolder, errGetToFolder := m.foldersRepository.GetByID(ctx, folderIDTo)
	if errGetToFolder != nil {
		return nil, errGetToFolder
	}
	for _, folder := range foldersList {
		id, errGetID := m.foldersRepository.GetID(ctx)
		if errGetID != nil {
			return nil, errGetID
		}
		newFolder, errCreateFolder := m.foldersRepository.Create(ctx, folders.Folder{
			Model: model.Model{ID: id}, ParentID: toFolder.ID, UID: gofakeit.UUID(), Name: folder.Name,
		})
		if errCreateFolder != nil {
			return nil, errCreateFolder
		}
		results[folder.ID] = newFolder
	}
	return results, nil
}

func (m *SecretsService) copySecrets(ctx context.Context, secretsList []*secrets.Secret, folderIDFrom uint, folderIDTo uint) (map[uint]*secrets.Secret, error) {
	results := make(map[uint]*secrets.Secret)
	_, errGetFromFolder := m.foldersRepository.GetByID(ctx, folderIDFrom)
	if errGetFromFolder != nil {
		return nil, errGetFromFolder
	}
	toFolder, errGetToFolder := m.foldersRepository.GetByID(ctx, folderIDTo)
	if errGetToFolder != nil {
		return nil, errGetToFolder
	}
	for _, secret := range secretsList {
		id, errGetID := m.secretsRepository.GetID(ctx)
		if errGetID != nil {
			return nil, errGetID
		}
		newSecret, errCreateSecret := m.secretsRepository.Create(ctx, secrets.Secret{
			Model: model.Model{ID: id}, FolderID: toFolder.ID, UID: gofakeit.UUID(),
			Name: secret.Name, Value: secret.Value, Type: secret.Type, IsDynamic: secret.IsDynamic,
		})
		if errCreateSecret != nil {
			return nil, errCreateSecret
		}
		results[secret.ID] = newSecret
	}
	return results, nil
}

func (m *SecretsService) getSecretsByFolder(ctx context.Context, parentFolderID uint) ([]*secrets.Secret, error) {
	return m.secretsRepository.Get(ctx, secrets.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No},
		FolderIDs:  []uint{parentFolderID},
	})
}

func (m *SecretsService) getFoldersByFolder(ctx context.Context, parentFolderID uint) ([]*folders.Folder, error) {
	return m.foldersRepository.Get(ctx, folders.ListFolderParams{
		ListParams:     generics.ListParams{Deleted: model.No},
		ParentFolderID: parentFolderID,
	})
}
