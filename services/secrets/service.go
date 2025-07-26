package secrets

import (
	"context"
	"hideout/internal/paths"
	"hideout/internal/secrets"
	"hideout/structs"
)

type Config struct {
}

type SecretsService struct {
	config  *Config
	paths   *[]paths.Path
	secrets *[]secrets.Secret

	secretsRepository secrets.Repository
	pathsRepository   paths.Repository
}

// NewService Creation of the service
func NewService(config Config, pathsList *[]paths.Path, secretsList *[]secrets.Secret) (*SecretsService, error) {
	return &SecretsService{config: &config, paths: pathsList, secrets: secretsList,
		secretsRepository: secrets.NewRepository(structs.Secrets), pathsRepository: paths.NewRepository(structs.Paths)}, nil
}

func (s *SecretsService) CreateSecret(ctx context.Context, pathID uint, name string, value string, valueType string) (*secrets.Secret, error) {
	return s.secretsRepository.Create(ctx, pathID, name, value, valueType)
}

func (s *SecretsService) CreatePath(ctx context.Context, pathID uint, name string) (*paths.Path, error) {
	return s.pathsRepository.Create(ctx, pathID, name)
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
		result.Children = append(result.Children, TreeNode{
			Name: existingPathSecret.Name,
			Type: "Secret",
		})
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
		newPath, errCreatePath := s.pathsRepository.Create(ctx, toPath.ID, path.Name)
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
		newSecret, errCreateSecret := s.secretsRepository.Create(ctx, toPath.ID, secret.Name, secret.Value, secret.Type)
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
