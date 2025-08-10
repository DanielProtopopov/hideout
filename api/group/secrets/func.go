package secrets

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/risor-io/risor"
	"github.com/risor-io/risor/object"
	"hideout/internal/common/apperror"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	secrets2 "hideout/internal/secrets"
	"hideout/services/secrets"
	"slices"
)

func (s *Secret) Process(ctx context.Context, secretsSvc *secrets.SecretsService) (string, string, error) {
	if secretsSvc == nil {
		return "", "", errors.New("Secrets service is non-existent")
	}

	secretsList, errGetSecretUIDs := secretsSvc.GetSecrets(ctx, secrets2.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No},
	})
	if errGetSecretUIDs != nil {
		return "", "", errGetSecretUIDs
	}

	// Delete current one from the list to avoid self-referencing
	secretsList = slices.DeleteFunc(secretsList, func(dp *secrets2.Secret) bool {
		return dp.UID == s.UID
	})

	var globalValues = map[string]any{}
	// Reference secrets by {{id}} and {{uid}} constructs
	for _, secretEntry := range secretsList {
		globalValues[fmt.Sprintf("{{%s}}", secretEntry.UID)] = secretEntry.Value
		globalValues[fmt.Sprintf("{{%d}}", secretEntry.ID)] = secretEntry.Value
	}

	// Disable dangerous and unnecessary modules
	evaluatedResult, errEvaluate := risor.Eval(ctx, s.Value, risor.WithGlobals(globalValues),
		risor.WithoutGlobals("errors", "exec", "filepath", "http", "net", "os"))
	if errEvaluate != nil {
		return "", "", errEvaluate
	}

	valueType := evaluatedResult.Type()
	switch valueType {
	case object.BOOL:
		return fmt.Sprintf("%t", evaluatedResult.Interface().(bool)), string(object.BOOL), nil
	case object.STRING:
		return fmt.Sprintf("%s", evaluatedResult.Interface().(string)), string(object.STRING), nil
	case object.INT:
		return fmt.Sprintf("%d", evaluatedResult.Interface().(int)), string(object.STRING), nil
	case object.FLOAT:
		return fmt.Sprintf("%f", evaluatedResult.Interface().(float64)), string(object.STRING), nil
	}

	return "", string(valueType), fmt.Errorf("Cannot process result with type of %s", string(valueType))
}

func ExportToDotEnv(secrets []Secret) (string, error) {
	return "", apperror.ErrNotImplemented
}

func ArchiveExport(data []byte, archiveType uint, compressionType uint) (string, string, []byte, error) {
	return "", "", nil, apperror.ErrNotImplemented
}
