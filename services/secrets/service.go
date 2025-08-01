package secrets

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/pkg/errors"
	"hideout/config"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/paths"
	"hideout/internal/secrets"
	"hideout/structs"
)

type Config struct {
}

type SecretsService struct {
	config  config.RepositoryConfig
	paths   *[]paths.Path
	secrets *[]secrets.Secret

	secretsRepository secrets.Repository
	pathsRepository   paths.Repository
}

// NewService Creation of the service
func NewService(ctx context.Context, config config.RepositoryConfig, pathsList *[]paths.Path, secretsList *[]secrets.Secret) (*SecretsService, error) {
	switch config.Type {
	case RepositoryType_InMemory:
		{
			return &SecretsService{config: config, paths: pathsList, secrets: secretsList,
				secretsRepository: secrets.NewInMemoryRepository(&structs.Secrets), pathsRepository: paths.NewInMemoryRepository(&structs.Paths)}, nil
		}
	case RepositoryType_Redis:
		{
			var inMemorySecretsRep *secrets.InMemoryRepository = nil
			var inMemoryPathsRep *paths.InMemoryRepository = nil
			if config.PreloadInMemory {
				inMemorySecretsRep = secrets.NewInMemoryRepository(&structs.Secrets)
				inMemoryPathsRep = paths.NewInMemoryRepository(&structs.Paths)
			}
			redisSecretsRep := secrets.NewRedisRepository(structs.Redis, inMemorySecretsRep)
			redisPathsRep := paths.NewRedisRepository(structs.Redis, inMemoryPathsRep)
			if config.PreloadInMemory {
				loadedSecrets, errLoadSecrets := redisSecretsRep.Load(ctx)
				if errLoadSecrets != nil {
					return nil, errors.Wrap(errLoadSecrets, "Error preloading secrets into in-memory storage in Redis")
				}
				structs.Secrets = loadedSecrets
				loadedPaths, errLoadPaths := redisPathsRep.Load(ctx)
				if errLoadPaths != nil {
					return nil, errors.Wrap(errLoadPaths, "Error preloading paths into in-memory storage in Redis")
				}
				structs.Paths = loadedPaths
			}
			secretsSvc := SecretsService{config: config, paths: pathsList, secrets: secretsList,
				secretsRepository: redisSecretsRep, pathsRepository: redisPathsRep}

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
			var inMemoryPathsRep *paths.InMemoryRepository = nil
			if config.PreloadInMemory {
				inMemorySecretsRep = secrets.NewInMemoryRepository(&structs.Secrets)
				inMemoryPathsRep = paths.NewInMemoryRepository(&structs.Paths)
			}
			databaseSecretsRep := secrets.NewDatabaseRepository(structs.Gorm, inMemorySecretsRep)
			databasePathsRep := paths.NewDatabaseRepository(structs.Gorm, inMemoryPathsRep)

			secretsSvc := SecretsService{config: config, paths: pathsList, secrets: secretsList,
				secretsRepository: databaseSecretsRep, pathsRepository: databasePathsRep}

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
		loadedPaths, errLoadPaths := s.pathsRepository.Load(ctx)
		if errLoadPaths != nil {
			return errors.Wrap(errLoadPaths, "Error preloading paths into in-memory storage in Redis")
		}
		structs.Paths = loadedPaths
	}

	return nil
}

func (s *SecretsService) GetPathByUID(ctx context.Context, pathUID string) (*paths.Path, error) {
	return s.pathsRepository.GetByUID(ctx, pathUID)
}

func (s *SecretsService) GetPaths(ctx context.Context, pathID uint) ([]*paths.Path, error) {
	return s.pathsRepository.Get(ctx, paths.ListPathParams{
		ListParams: generics.ListParams{Deleted: model.No}, ParentPathID: pathID,
	})
}

func (s *SecretsService) GetSecrets(ctx context.Context, pathID uint) ([]*secrets.Secret, error) {
	return s.secretsRepository.Get(ctx, secrets.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No}, PathIDs: []uint{pathID},
	})
}

func (s *SecretsService) DeleteSecret(ctx context.Context, secretID uint, forceDelete bool) error {
	return s.secretsRepository.Delete(ctx, secretID, forceDelete)
}

func (s *SecretsService) CreateSecret(ctx context.Context, pathID uint, name string, value string, valueType string) (*secrets.Secret, error) {
	id, errGetID := s.secretsRepository.GetID(ctx)
	if errGetID != nil {
		return nil, errGetID
	}
	return s.secretsRepository.Create(ctx, secrets.Secret{Model: model.Model{ID: id}, PathID: pathID, UID: gofakeit.UUID(),
		Name: name, Value: value, Type: valueType})
}

func (s *SecretsService) CreatePath(ctx context.Context, pathID uint, name string) (*paths.Path, error) {
	id, errGetID := s.pathsRepository.GetID(ctx)
	if errGetID != nil {
		return nil, errGetID
	}
	return s.pathsRepository.Create(ctx, paths.Path{Model: model.Model{ID: id}, ParentID: pathID,
		UID: gofakeit.UUID(), Name: name})
}

func (s *SecretsService) Tree(ctx context.Context, pathID uint) (TreeNode, error) {
	result := TreeNode{Name: "", Type: "Path", Children: nil}
	existingPath, errGetPath := s.pathsRepository.GetByID(ctx, pathID)
	if errGetPath != nil {
		return result, errGetPath
	}
	result.Name = existingPath.Name

	existingPathPaths, errGetExistingPathPaths := s.getPathsByPath(ctx, existingPath.ID)
	if errGetExistingPathPaths != nil {
		return result, errGetExistingPathPaths
	}
	existingPathSecrets, errGetExistingPathSecrets := s.getSecretsByPath(ctx, existingPath.ID)
	if errGetExistingPathSecrets != nil {
		return result, errGetExistingPathSecrets
	}

	for _, existingPathSecret := range existingPathSecrets {
		result.Children = append(result.Children, TreeNode{Name: existingPathSecret.Name, Type: "Secret"})
	}

	for _, existingPathPath := range existingPathPaths {
		pathNode, errGetPathNode := s.Tree(ctx, existingPathPath.ID)
		if errGetPathNode != nil {
			return result, errGetPathNode
		}
		result.Children = append(result.Children, pathNode)
	}

	return result, nil
}

func (s *SecretsService) Delete(ctx context.Context, existingPaths []*paths.Path, existingSecrets []*secrets.Secret, pathIDFrom uint, forceDelete bool) ([]*paths.Path, []*secrets.Secret, error) {
	var deletedPaths []*paths.Path
	var deletedSecrets []*secrets.Secret

	_, errGetPathFrom := s.pathsRepository.GetByID(ctx, pathIDFrom)
	if errGetPathFrom != nil {
		return nil, nil, errGetPathFrom
	}

	// This deletes secrets From designed path
	for _, existingSecret := range existingSecrets {
		errDeleteSecret := s.secretsRepository.Delete(ctx, existingSecret.ID, forceDelete)
		if errDeleteSecret != nil {
			return nil, nil, errDeleteSecret
		}
		deletedSecrets = append(deletedSecrets, existingSecret)
	}

	// This recursively deletes paths and their secrets from sub-paths
	for _, existingPath := range existingPaths {
		existingPathPaths, errGetExistingPathPaths := s.getPathsByPath(ctx, existingPath.ID)
		if errGetExistingPathPaths != nil {
			return nil, nil, errGetExistingPathPaths
		}
		existingPathSecrets, errGetExistingPathSecrets := s.getSecretsByPath(ctx, existingPath.ID)
		if errGetExistingPathSecrets != nil {
			return nil, nil, errGetExistingPathSecrets
		}

		// Delete paths & secrets from an existing path
		deletedPathPaths, deletedPathSecrets, errDelete := s.Delete(ctx, existingPathPaths, existingPathSecrets, existingPath.ID, forceDelete)
		if errDelete != nil {
			return nil, nil, errDelete
		}

		errDeletePath := s.pathsRepository.Delete(ctx, existingPath.ID, forceDelete)
		if errDeletePath != nil {
			return nil, nil, errDeletePath
		}
		deletedPaths = append(deletedPaths, existingPath)

		deletedPaths = append(deletedPaths, deletedPathPaths...)
		deletedSecrets = append(deletedSecrets, deletedPathSecrets...)
	}

	return deletedPaths, deletedSecrets, nil
}

func (s *SecretsService) Copy(ctx context.Context, existingPaths []*paths.Path, existingSecrets []*secrets.Secret, pathIDFrom uint, pathIDTo uint) ([]*paths.Path, []*secrets.Secret, error) {
	var newSecrets []*secrets.Secret
	var newPaths []*paths.Path

	_, errGetPathFrom := s.pathsRepository.GetByID(ctx, pathIDFrom)
	if errGetPathFrom != nil {
		return nil, nil, errGetPathFrom
	}
	_, errGetPathTo := s.pathsRepository.GetByID(ctx, pathIDTo)
	if errGetPathTo != nil {
		return nil, nil, errGetPathTo
	}

	// This copies secrets From designed path To target path
	copiedSecretsMap, errCopySecrets := s.copySecrets(ctx, existingSecrets, pathIDFrom, pathIDTo)
	if errCopySecrets != nil {
		return nil, nil, errCopySecrets
	}
	for _, copiedSecret := range copiedSecretsMap {
		newSecrets = append(newSecrets, copiedSecret)
	}

	// This copies paths From designed path To target path
	copiedPathsMap, errCopyPaths := s.copyPaths(ctx, existingPaths, pathIDFrom, pathIDTo)
	if errCopyPaths != nil {
		return nil, nil, errCopyPaths
	}
	for _, copiedPath := range copiedPathsMap {
		newPaths = append(newPaths, copiedPath)
	}

	// This recursively copies paths and their secrets from paths in From path
	for _, existingPath := range existingPaths {
		copiedPath, _ := copiedPathsMap[existingPath.ID]
		existingPathPaths, errGetExistingPathPaths := s.getPathsByPath(ctx, existingPath.ID)
		if errGetExistingPathPaths != nil {
			return nil, nil, errGetExistingPathPaths
		}
		existingPathSecrets, errGetExistingPathSecrets := s.getSecretsByPath(ctx, existingPath.ID)
		if errGetExistingPathSecrets != nil {
			return nil, nil, errGetExistingPathSecrets
		}

		// Copy paths & secrets from an existing path to a copied path
		createdPaths, createdSecrets, errCopy := s.Copy(ctx, existingPathPaths, existingPathSecrets, existingPath.ID, copiedPath.ID)
		if errCopy != nil {
			return nil, nil, errCopy
		}
		newPaths = append(newPaths, createdPaths...)
		newSecrets = append(newSecrets, createdSecrets...)
	}

	return newPaths, newSecrets, nil
}

func (s *SecretsService) copyPaths(ctx context.Context, pathsList []*paths.Path, pathIDFrom uint, pathIDTo uint) (map[uint]*paths.Path, error) {
	results := make(map[uint]*paths.Path)
	_, errGetFromPath := s.pathsRepository.GetByID(ctx, pathIDFrom)
	if errGetFromPath != nil {
		return nil, errGetFromPath
	}
	toPath, errGetToPath := s.pathsRepository.GetByID(ctx, pathIDTo)
	if errGetToPath != nil {
		return nil, errGetToPath
	}
	for _, path := range pathsList {
		id, errGetID := s.pathsRepository.GetID(ctx)
		if errGetID != nil {
			return nil, errGetID
		}
		newPath, errCreatePath := s.pathsRepository.Create(ctx, paths.Path{
			Model: model.Model{ID: id}, ParentID: toPath.ID, UID: gofakeit.UUID(), Name: path.Name,
		})
		if errCreatePath != nil {
			return nil, errCreatePath
		}
		results[path.ID] = newPath
	}
	return results, nil
}

func (s *SecretsService) copySecrets(ctx context.Context, secretsList []*secrets.Secret, pathIDFrom uint, pathIDTo uint) (map[uint]*secrets.Secret, error) {
	results := make(map[uint]*secrets.Secret)
	_, errGetFromPath := s.pathsRepository.GetByID(ctx, pathIDFrom)
	if errGetFromPath != nil {
		return nil, errGetFromPath
	}
	toPath, errGetToPath := s.pathsRepository.GetByID(ctx, pathIDTo)
	if errGetToPath != nil {
		return nil, errGetToPath
	}
	for _, secret := range secretsList {
		id, errGetID := s.secretsRepository.GetID(ctx)
		if errGetID != nil {
			return nil, errGetID
		}
		newSecret, errCreateSecret := s.secretsRepository.Create(ctx, secrets.Secret{
			Model: model.Model{ID: id}, PathID: toPath.ID, UID: gofakeit.UUID(),
			Name: secret.Name, Value: secret.Value, Type: secret.Type,
		})
		if errCreateSecret != nil {
			return nil, errCreateSecret
		}
		results[secret.ID] = newSecret
	}
	return results, nil
}

func (s *SecretsService) getSecretsByPath(ctx context.Context, parentPathID uint) ([]*secrets.Secret, error) {
	var results []*secrets.Secret
	parentPath, errGetPath := s.pathsRepository.GetByID(ctx, parentPathID)
	if errGetPath != nil {
		return nil, errGetPath
	}

	for _, secret := range *s.secrets {
		if secret.PathID == parentPath.ID {
			results = append(results, &secret)
		}
	}

	return results, nil
}

func (s *SecretsService) getPathsByPath(ctx context.Context, parentPathID uint) ([]*paths.Path, error) {
	var results []*paths.Path
	for _, path := range *s.paths {
		if path.ParentID == parentPathID {
			results = append(results, &path)
		}
	}

	return results, nil
}
