package secrets

import (
	"context"
	"hideout/internal/paths"
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

func (s *SecretsService) GetSecretsByPath(ctx context.Context, pathUID string) ([]*secrets.Secret, error) {
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

func (s *SecretsService) GetPathsByPath(ctx context.Context, parentPathID uint) ([]*paths.Path, error) {
	var results []*paths.Path
	for _, path := range s.Paths {
		if path.ParentID == parentPathID {
			results = append(results, &path)
		}
	}

	return results, nil
}
