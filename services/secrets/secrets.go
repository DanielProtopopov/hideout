package secrets

import (
	"context"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"hideout/internal/secrets"
	"regexp"
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

func (m *SecretsService) UpdateSecret(ctx context.Context, Localizer *i18n.Localizer, secret secrets.Secret) (*secrets.Secret, error) {
	_, errCompile := regexp.Compile(`^[A-Za-z0-9_].+=.+$`)
	if errCompile != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CompileSecretValueRegexError"}})
		return nil, errors.Wrap(errCompile, msg)
	}

	return m.secretsRepository.Update(ctx, secret)
}

func (m *SecretsService) CreateSecret(ctx context.Context, Localizer *i18n.Localizer, secret secrets.Secret) (*secrets.Secret, error) {
	_, errCompile := regexp.Compile(`^[A-Za-z0-9_].+=.+$`)
	if errCompile != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CompileSecretValueRegexError"}})
		return nil, errors.Wrap(errCompile, msg)
	}

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
