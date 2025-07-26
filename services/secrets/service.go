package secrets

import (
	"context"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/paths"
	error2 "hideout/internal/pkg/error"
	"hideout/internal/secrets"
)

type Config struct {
}

type SecretsService struct {
	Config  *Config
	Paths   []paths.Path
	Secrets []secrets.Secret

	SecretsRepository secrets.Repository
	PathsRepository   paths.Repository
}

// NewService Creation of the service
func NewService(config Config, pathsList []paths.Path, secretsList []secrets.Secret, secretsRep secrets.Repository, pathsRep paths.Repository) (*SecretsService, error) {
	return &SecretsService{Config: &config, Paths: pathsList, Secrets: secretsList,
		SecretsRepository: secretsRep, PathsRepository: pathsRep}, nil
}

func (s *SecretsService) Copy(ctx context.Context, pathIDs []uint, secretIDs []uint, pathIDFrom uint, pathIDTo uint) ([]*paths.Path, []*secrets.Secret, error) {
	return nil, nil, error2.ErrNotImplemented
}

func (s *SecretsService) copyPaths(ctx context.Context, pathIDs []uint, pathIDFrom uint, pathIDTo uint) ([]*secrets.Secret, error) {
	return nil, error2.ErrNotImplemented
}

func (s *SecretsService) copySecrets(ctx context.Context, secretIDs []uint, pathIDFrom uint, pathIDTo uint) ([]*secrets.Secret, error) {
	fromPath, errGetFromPath := s.PathsRepository.GetByID(ctx, pathIDFrom)
	if errGetFromPath != nil {
		return nil, errGetFromPath
	}
	toPath, errGetToPath := s.PathsRepository.GetByID(ctx, pathIDTo)
	if errGetToPath != nil {
		return nil, errGetToPath
	}
	secretsList, errGetSecrets := s.SecretsRepository.Get(ctx, secrets.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No, IDs: secretIDs}, PathIDs: []uint{fromPath.ID},
	})
	if errGetSecrets != nil {
		return nil, errGetSecrets
	}
	var newSecrets []*secrets.Secret
	for _, secret := range secretsList {
		newSecret, errCreateSecret := s.SecretsRepository.Create(ctx, toPath.ID, secret.Name, secret.Value, secret.Type)
		if errCreateSecret != nil {
			return nil, errCreateSecret
		}
		newSecrets = append(newSecrets, newSecret)
	}
	return newSecrets, nil
}

func (s *SecretsService) getSecretsByPath(ctx context.Context, pathUID string) ([]*secrets.Secret, error) {
	var results []*secrets.Secret
	pathByUID, errGetPath := s.PathsRepository.GetByUID(ctx, pathUID)
	if errGetPath != nil {
		return nil, errGetPath
	}

	for _, secret := range s.Secrets {
		if secret.PathID == pathByUID.ID {
			results = append(results, &secret)
		}
	}

	return results, nil
}

func (s *SecretsService) getPathsByPath(ctx context.Context, parentPathID uint) ([]*paths.Path, error) {
	var results []*paths.Path
	for _, path := range s.Paths {
		if path.ParentID == parentPathID {
			results = append(results, &path)
		}
	}

	return results, nil
}
