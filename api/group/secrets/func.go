package secrets

import (
	"context"
	"github.com/pkg/errors"
	"github.com/risor-io/risor"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	secrets2 "hideout/internal/secrets"
	"hideout/services/secrets"
	"slices"
)

func (s *Secret) Process(ctx context.Context, secretsSvc *secrets.SecretsService) (string, error) {
	if secretsSvc == nil {
		return "", errors.New("Secrets service is non-existent")
	}

	secretsList, errGetSecretUIDs := secretsSvc.GetSecrets(ctx, secrets2.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No},
	})
	if errGetSecretUIDs != nil {
		return "", errGetSecretUIDs
	}

	// Delete current one from the list to avoid self-referencing
	secretsList = slices.DeleteFunc(secretsList, func(dp *secrets2.Secret) bool {
		return dp.UID == s.UID
	})

	var globalValues = map[string]any{}
	for _, secretEntry := range secretsList {
		globalValues[secretEntry.UID] = secretEntry.Value
	}

	evaluatedResult, errEvaluate := risor.Eval(ctx, s.Value, risor.WithGlobals(globalValues))
	if errEvaluate != nil {
		return "", errEvaluate
	}

	valueType := evaluatedResult.Type()
	if valueType == "string" {
		return evaluatedResult.Interface().(string), nil
	}

	return "", errors.New("Evaluated to a non-string value")
}
