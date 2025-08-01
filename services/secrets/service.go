package secrets

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/pkg/errors"
	"hideout/config"
	"hideout/internal/common/model"
	"hideout/internal/folders"
	"hideout/internal/secrets"
	"hideout/structs"
)

type Config struct {
}

type SecretsService struct {
	config  config.RepositoryConfig
	folders *[]folders.Folder
	secrets *[]secrets.Secret

	secretsRepository secrets.Repository
	foldersRepository folders.Repository
}

// NewService Creation of the service
func NewService(ctx context.Context, config config.RepositoryConfig, foldersList *[]folders.Folder, secretsList *[]secrets.Secret) (*SecretsService, error) {
	switch config.Type {
	case RepositoryType_InMemory:
		{
			return &SecretsService{config: config, folders: foldersList, secrets: secretsList,
				secretsRepository: secrets.NewInMemoryRepository(&structs.Secrets), foldersRepository: folders.NewInMemoryRepository(&structs.Folders)}, nil
		}
	case RepositoryType_Redis:
		{
			var inMemorySecretsRep *secrets.InMemoryRepository = nil
			var inMemoryFoldersRep *folders.InMemoryRepository = nil
			if config.PreloadInMemory {
				inMemorySecretsRep = secrets.NewInMemoryRepository(&structs.Secrets)
				inMemoryFoldersRep = folders.NewInMemoryRepository(&structs.Folders)
			}
			redisSecretsRep := secrets.NewRedisRepository(structs.Redis, inMemorySecretsRep)
			redisFoldersRep := folders.NewRedisRepository(structs.Redis, inMemoryFoldersRep)
			if config.PreloadInMemory {
				loadedSecrets, errLoadSecrets := redisSecretsRep.Load(ctx)
				if errLoadSecrets != nil {
					return nil, errors.Wrap(errLoadSecrets, "Error preloading secrets into in-memory storage in Redis")
				}
				structs.Secrets = loadedSecrets
				loadedFolders, errLoadFolders := redisFoldersRep.Load(ctx)
				if errLoadFolders != nil {
					return nil, errors.Wrap(errLoadFolders, "Error preloading folders into in-memory storage in Redis")
				}
				structs.Folders = loadedFolders
			}
			secretsSvc := SecretsService{config: config, folders: foldersList, secrets: secretsList,
				secretsRepository: redisSecretsRep, foldersRepository: redisFoldersRep}

			if config.PreloadInMemory {
				errLoad := secretsSvc.Load(ctx)
				if errLoad != nil {
					return nil, errors.Wrap(errLoad, "Error loading data into memory")
				}
			}

			return &secretsSvc, nil
		}
	case RepositoryType_Database:
		{
			var inMemorySecretsRep *secrets.InMemoryRepository = nil
			var inMemoryFoldersRep *folders.InMemoryRepository = nil
			if config.PreloadInMemory {
				inMemorySecretsRep = secrets.NewInMemoryRepository(&structs.Secrets)
				inMemoryFoldersRep = folders.NewInMemoryRepository(&structs.Folders)
			}
			databaseSecretsRep := secrets.NewDatabaseRepository(structs.Gorm, inMemorySecretsRep)
			databaseFoldersRep := folders.NewDatabaseRepository(structs.Gorm, inMemoryFoldersRep)

			secretsSvc := SecretsService{config: config, folders: foldersList, secrets: secretsList,
				secretsRepository: databaseSecretsRep, foldersRepository: databaseFoldersRep}

			if config.PreloadInMemory {
				errLoad := secretsSvc.Load(ctx)
				if errLoad != nil {
					return nil, errors.Wrap(errLoad, "Error loading data into memory")
				}
			}

			return &secretsSvc, nil
		}
	}

	return nil, errors.New("invalid repository type")
}

func (s *SecretsService) Load(ctx context.Context) error {
	if s.config.PreloadInMemory {
		loadedSecrets, errLoadSecrets := s.secretsRepository.Load(ctx)
		if errLoadSecrets != nil {
			return errors.Wrap(errLoadSecrets, "Error preloading secrets into in-memory storage in Redis")
		}
		structs.Secrets = loadedSecrets
		loadedFolders, errLoadFolders := s.foldersRepository.Load(ctx)
		if errLoadFolders != nil {
			return errors.Wrap(errLoadFolders, "Error preloading folders into in-memory storage in Redis")
		}
		structs.Folders = loadedFolders
	}

	return nil
}

func (s *SecretsService) Tree(ctx context.Context, folderID uint) (TreeNode, error) {
	result := TreeNode{Name: "", Type: "Folder", Children: nil}
	existingFolder, errGetFolder := s.foldersRepository.GetByID(ctx, folderID)
	if errGetFolder != nil {
		return result, errGetFolder
	}
	result.Name = existingFolder.Name

	existingFolderFolders, errGetExistingFolderFolders := s.getFoldersByFolder(ctx, existingFolder.ID)
	if errGetExistingFolderFolders != nil {
		return result, errGetExistingFolderFolders
	}
	existingFolderSecrets, errGetExistingFolderSecrets := s.getSecretsByFolder(ctx, existingFolder.ID)
	if errGetExistingFolderSecrets != nil {
		return result, errGetExistingFolderSecrets
	}

	for _, existingFolderSecret := range existingFolderSecrets {
		result.Children = append(result.Children, TreeNode{Name: existingFolderSecret.Name, Type: "Secret"})
	}

	for _, existingFolderFolder := range existingFolderFolders {
		folderNode, errGetFolderNode := s.Tree(ctx, existingFolderFolder.ID)
		if errGetFolderNode != nil {
			return result, errGetFolderNode
		}
		result.Children = append(result.Children, folderNode)
	}

	return result, nil
}

func (s *SecretsService) Delete(ctx context.Context, existingFolders []*folders.Folder, existingSecrets []*secrets.Secret, folderIDFrom uint, forceDelete bool) ([]*folders.Folder, []*secrets.Secret, error) {
	var deletedFolders []*folders.Folder
	var deletedSecrets []*secrets.Secret

	_, errGetFolderFrom := s.foldersRepository.GetByID(ctx, folderIDFrom)
	if errGetFolderFrom != nil {
		return nil, nil, errGetFolderFrom
	}

	// This deletes secrets From designed folder
	for _, existingSecret := range existingSecrets {
		errDeleteSecret := s.secretsRepository.Delete(ctx, existingSecret.ID, forceDelete)
		if errDeleteSecret != nil {
			return nil, nil, errDeleteSecret
		}
		deletedSecrets = append(deletedSecrets, existingSecret)
	}

	// This recursively deletes folders and their secrets from sub-folders
	for _, existingFolder := range existingFolders {
		existingFolderFolders, errGetExistingFolderFolders := s.getFoldersByFolder(ctx, existingFolder.ID)
		if errGetExistingFolderFolders != nil {
			return nil, nil, errGetExistingFolderFolders
		}
		existingFolderSecrets, errGetExistingFolderSecrets := s.getSecretsByFolder(ctx, existingFolder.ID)
		if errGetExistingFolderSecrets != nil {
			return nil, nil, errGetExistingFolderSecrets
		}

		// Delete folders & secrets from an existing folder
		deletedFolderFolders, deletedFolderSecrets, errDelete := s.Delete(ctx, existingFolderFolders, existingFolderSecrets, existingFolder.ID, forceDelete)
		if errDelete != nil {
			return nil, nil, errDelete
		}

		errDeleteFolder := s.foldersRepository.Delete(ctx, existingFolder.ID, forceDelete)
		if errDeleteFolder != nil {
			return nil, nil, errDeleteFolder
		}
		deletedFolders = append(deletedFolders, existingFolder)

		deletedFolders = append(deletedFolders, deletedFolderFolders...)
		deletedSecrets = append(deletedSecrets, deletedFolderSecrets...)
	}

	return deletedFolders, deletedSecrets, nil
}

func (s *SecretsService) Copy(ctx context.Context, existingFolders []*folders.Folder, existingSecrets []*secrets.Secret, folderIDFrom uint, folderIDTo uint) ([]*folders.Folder, []*secrets.Secret, error) {
	var newSecrets []*secrets.Secret
	var newFolders []*folders.Folder

	_, errGetFolderFrom := s.foldersRepository.GetByID(ctx, folderIDFrom)
	if errGetFolderFrom != nil {
		return nil, nil, errGetFolderFrom
	}
	_, errGetFolderTo := s.foldersRepository.GetByID(ctx, folderIDTo)
	if errGetFolderTo != nil {
		return nil, nil, errGetFolderTo
	}

	// This copies secrets From designed folder To target folder
	copiedSecretsMap, errCopySecrets := s.copySecrets(ctx, existingSecrets, folderIDFrom, folderIDTo)
	if errCopySecrets != nil {
		return nil, nil, errCopySecrets
	}
	for _, copiedSecret := range copiedSecretsMap {
		newSecrets = append(newSecrets, copiedSecret)
	}

	// This copies folders From designed folder To target folder
	copiedFoldersMap, errCopyFolders := s.copyFolders(ctx, existingFolders, folderIDFrom, folderIDTo)
	if errCopyFolders != nil {
		return nil, nil, errCopyFolders
	}
	for _, copiedFolder := range copiedFoldersMap {
		newFolders = append(newFolders, copiedFolder)
	}

	// This recursively copies folders and their secrets from folders in From folder
	for _, existingFolder := range existingFolders {
		copiedFolder, _ := copiedFoldersMap[existingFolder.ID]
		existingFolderFolders, errGetExistingFolderFolders := s.getFoldersByFolder(ctx, existingFolder.ID)
		if errGetExistingFolderFolders != nil {
			return nil, nil, errGetExistingFolderFolders
		}
		existingFolderSecrets, errGetExistingFolderSecrets := s.getSecretsByFolder(ctx, existingFolder.ID)
		if errGetExistingFolderSecrets != nil {
			return nil, nil, errGetExistingFolderSecrets
		}

		// Copy folders & secrets from an existing folder to a copied folder
		createdFolders, createdSecrets, errCopy := s.Copy(ctx, existingFolderFolders, existingFolderSecrets, existingFolder.ID, copiedFolder.ID)
		if errCopy != nil {
			return nil, nil, errCopy
		}
		newFolders = append(newFolders, createdFolders...)
		newSecrets = append(newSecrets, createdSecrets...)
	}

	return newFolders, newSecrets, nil
}

func (s *SecretsService) copyFolders(ctx context.Context, foldersList []*folders.Folder, folderIDFrom uint, folderIDTo uint) (map[uint]*folders.Folder, error) {
	results := make(map[uint]*folders.Folder)
	_, errGetFromFolder := s.foldersRepository.GetByID(ctx, folderIDFrom)
	if errGetFromFolder != nil {
		return nil, errGetFromFolder
	}
	toFolder, errGetToFolder := s.foldersRepository.GetByID(ctx, folderIDTo)
	if errGetToFolder != nil {
		return nil, errGetToFolder
	}
	for _, folder := range foldersList {
		id, errGetID := s.foldersRepository.GetID(ctx)
		if errGetID != nil {
			return nil, errGetID
		}
		newFolder, errCreateFolder := s.foldersRepository.Create(ctx, folders.Folder{
			Model: model.Model{ID: id}, ParentID: toFolder.ID, UID: gofakeit.UUID(), Name: folder.Name,
		})
		if errCreateFolder != nil {
			return nil, errCreateFolder
		}
		results[folder.ID] = newFolder
	}
	return results, nil
}

func (s *SecretsService) copySecrets(ctx context.Context, secretsList []*secrets.Secret, folderIDFrom uint, folderIDTo uint) (map[uint]*secrets.Secret, error) {
	results := make(map[uint]*secrets.Secret)
	_, errGetFromFolder := s.foldersRepository.GetByID(ctx, folderIDFrom)
	if errGetFromFolder != nil {
		return nil, errGetFromFolder
	}
	toFolder, errGetToFolder := s.foldersRepository.GetByID(ctx, folderIDTo)
	if errGetToFolder != nil {
		return nil, errGetToFolder
	}
	for _, secret := range secretsList {
		id, errGetID := s.secretsRepository.GetID(ctx)
		if errGetID != nil {
			return nil, errGetID
		}
		newSecret, errCreateSecret := s.secretsRepository.Create(ctx, secrets.Secret{
			Model: model.Model{ID: id}, FolderID: toFolder.ID, UID: gofakeit.UUID(),
			Name: secret.Name, Value: secret.Value, Type: secret.Type,
		})
		if errCreateSecret != nil {
			return nil, errCreateSecret
		}
		results[secret.ID] = newSecret
	}
	return results, nil
}

func (s *SecretsService) getSecretsByFolder(ctx context.Context, parentFolderID uint) ([]*secrets.Secret, error) {
	var results []*secrets.Secret
	parentFolder, errGetFolder := s.foldersRepository.GetByID(ctx, parentFolderID)
	if errGetFolder != nil {
		return nil, errGetFolder
	}

	for _, secret := range *s.secrets {
		if secret.FolderID == parentFolder.ID {
			results = append(results, &secret)
		}
	}

	return results, nil
}

func (s *SecretsService) getFoldersByFolder(ctx context.Context, parentFolderID uint) ([]*folders.Folder, error) {
	var results []*folders.Folder
	for _, folder := range *s.folders {
		if folder.ParentID == parentFolderID {
			results = append(results, &folder)
		}
	}

	return results, nil
}
