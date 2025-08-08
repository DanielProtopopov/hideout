package secrets

import (
	"context"
	"hideout/internal/secrets"
)

func (m *SecretsService) GetSecretID(ctx context.Context) (uint, error) {
	return m.secretsRepository.GetID(ctx)
}

func (m *SecretsService) GetSecrets(ctx context.Context, params secrets.ListSecretParams) ([]*secrets.Secret, error) {
	return m.secretsRepository.Get(ctx, params)
}

func (m *SecretsService) GetSecretsMapByID(ctx context.Context, params secrets.ListSecretParams) (map[uint]*secrets.Secret, error) {
	return m.secretsRepository.GetMapByID(ctx, params)
}

func (m *SecretsService) GetSecretsMapByUID(ctx context.Context, params secrets.ListSecretParams) (map[string]*secrets.Secret, error) {
	return m.secretsRepository.GetMapByUID(ctx, params)
}

func (m *SecretsService) GetSecretByUID(ctx context.Context, uid string) (*secrets.Secret, error) {
	return m.secretsRepository.GetByUID(ctx, uid)
}

func (m *SecretsService) GetSecretByID(ctx context.Context, id uint) (*secrets.Secret, error) {
	return m.secretsRepository.GetByID(ctx, id)
}

func (m *SecretsService) UpdateSecret(ctx context.Context, secret secrets.Secret) (*secrets.Secret, error) {
	return m.secretsRepository.Update(ctx, secret)
}

func (m *SecretsService) CreateSecret(ctx context.Context, secret secrets.Secret) (*secrets.Secret, error) {
	secretID, errGetID := m.GetSecretID(ctx)
	if errGetID != nil {
		return nil, errGetID
	}
	secret.ID = secretID
	return m.secretsRepository.Create(ctx, secret)
}

func (m *SecretsService) DeleteSecret(ctx context.Context, id uint, forceDelete bool) error {
	return m.secretsRepository.Delete(ctx, id, forceDelete)
}

func (m *SecretsService) CountSecrets(ctx context.Context, params secrets.ListSecretParams) (uint, error) {
	return m.secretsRepository.Count(ctx, params)
}
