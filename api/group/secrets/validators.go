package secrets

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/common/rqrs"
	secrets2 "hideout/internal/secrets"
	"hideout/services/secrets"
	"regexp"
	"strings"
)

func (rq GetSecretsRQ) Validate(ctx context.Context, secretsService *secrets.SecretsService, Localizer *i18n.Localizer) (Errors []rqrs.Error) {
	errSecretsPagination := rq.SecretsPagination.Validate(ctx)
	if errSecretsPagination != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "PaginationError"}})
		Errors = append(Errors, rqrs.Error{Message: msg, Description: errors.Wrap(errSecretsPagination, "Secrets pagination validation failed").Error(), Code: 0})
	}

	errFoldersPagination := rq.FoldersPagination.Validate(ctx)
	if errFoldersPagination != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "PaginationError"}})
		Errors = append(Errors, rqrs.Error{Message: msg, Description: errors.Wrap(errSecretsPagination, "Folders pagination validation failed").Error(), Code: 0})
	}

	for _, orderVal := range rq.SecretsOrder {
		errSecretOrdering := orderVal.Validate(ctx, Localizer)
		if errSecretOrdering != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "OrderError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errors.Wrap(errSecretOrdering, "Secret order validation failed").Error(), Code: 0})
		}
	}

	for _, orderVal := range rq.FoldersOrder {
		errFolderOrdering := orderVal.Validate(ctx, Localizer)
		if errFolderOrdering != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "OrderError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errors.Wrap(errFolderOrdering, "Folder order validation failed").Error(), Code: 0})
		}
	}

	if rq.FolderUID != "" {
		_, errGetFolderByUID := secretsService.GetFolderByUID(ctx, rq.FolderUID)
		if errGetFolderByUID != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetFolderByUID.Error(), Code: 0})
		}
	}

	return Errors
}

func (rq CreateSecretsRQ) Validate(ctx context.Context, secretsService *secrets.SecretsService, Localizer *i18n.Localizer) (Errors []rqrs.Error) {
	if len(rq.Data) == 0 {
		Errors = append(Errors, rqrs.Error{Message: "No data for creation of folder(s)/secret(s) supplied", Description: "", Code: 0})
		return Errors
	}
	regexValue, errCompile := regexp.Compile(`^[A-Za-z0-9_].+=.+$`)
	if errCompile != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CompileSecretValueRegexError"}})
		Errors = append(Errors, rqrs.Error{Message: msg, Description: errCompile.Error(), Code: 0})
		return Errors
	}
	for _, createSecretEntry := range rq.Data {
		if createSecretEntry.Script != "" && createSecretEntry.Value != "" {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "OnlySecretOrValueError"},
				TemplateData: map[string]interface{}{"UID": createSecretEntry.Name}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
		}
		isValidName := regexValue.MatchString(createSecretEntry.Value)
		if !isValidName {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "InvalidSecretNameError"},
				TemplateData: map[string]interface{}{"UID": createSecretEntry.Name}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
		}

		secretEntry := Secret{
			UID: gofakeit.UUID(), FolderUID: createSecretEntry.FolderUID, Name: createSecretEntry.Name,
			Value: createSecretEntry.Value, Script: createSecretEntry.Script,
		}
		_, _, errProcessSecret := secretEntry.Process(ctx, secretsService)
		if errProcessSecret != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "DynamicSecretError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errProcessSecret.Error(), Code: 0})
		}

		folderByUID, errGetFolderByUID := secretsService.GetFolderByUID(ctx, createSecretEntry.FolderUID)
		if errGetFolderByUID != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetFolderByUID.Error(), Code: 0})
		}
		if folderByUID != nil {
			secretsInFolder, errGetSecrets := secretsService.GetSecrets(ctx, secrets2.ListSecretParams{
				ListParams: generics.ListParams{Deleted: model.No}, FolderIDs: []uint{folderByUID.ID},
			})
			if errGetSecrets != nil {
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretsError"}})
				Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetSecrets.Error(), Code: 0})
			}

			// Duplicate secret check
			for _, secretInFolder := range secretsInFolder {
				if strings.EqualFold(createSecretEntry.Name, secretInFolder.Name) {
					msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "SecretAlreadyExistsError"},
						TemplateData: map[string]interface{}{"Name": createSecretEntry.Name, "FolderUID": secretInFolder.UID}})
					Errors = append(Errors, rqrs.Error{Message: msg, Description: "", Code: 0})
				}
			}
		}
	}

	return Errors
}

func (rq UpdateSecretsRQ) Validate(ctx context.Context, secretsService *secrets.SecretsService, Localizer *i18n.Localizer) (Errors []rqrs.Error) {
	regexValue, errCompile := regexp.Compile(`^[A-Za-z0-9_].+=.+$`)
	if errCompile != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CompileSecretValueRegexError"}})
		Errors = append(Errors, rqrs.Error{Message: msg, Description: errCompile.Error(), Code: 0})
		return Errors
	}

	for _, updateSecretEntry := range rq.Data {
		if updateSecretEntry.Script != "" && updateSecretEntry.Value != "" {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "OnlySecretOrValueError"},
				TemplateData: map[string]interface{}{"UID": updateSecretEntry.Name}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
		}
		isValidName := regexValue.MatchString(updateSecretEntry.Value)
		if !isValidName {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "InvalidSecretNameError"},
				TemplateData: map[string]interface{}{"UID": updateSecretEntry.Name}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
		}
		_, _, errProcessSecret := updateSecretEntry.Process(ctx, secretsService)
		if errProcessSecret != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "DynamicSecretError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errProcessSecret.Error(), Code: 0})
		}
		_, errGetFolderByUID := secretsService.GetFolderByUID(ctx, updateSecretEntry.FolderUID)
		if errGetFolderByUID != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetFolderByUID.Error(), Code: 0})
		}
		_, errGetSecretByUID := secretsService.GetSecretByUID(ctx, updateSecretEntry.UID)
		if errGetSecretByUID != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretByUIDError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetSecretByUID.Error(), Code: 0})
		}
	}

	return Errors
}

func (rq DeleteSecretsRQ) Validate(ctx context.Context, secretsService *secrets.SecretsService, Localizer *i18n.Localizer) (Errors []rqrs.Error) {
	for _, deleteFolderEntry := range rq.FolderUIDs {
		_, errGetFolderByUID := secretsService.GetFolderByUID(ctx, deleteFolderEntry)
		if errGetFolderByUID != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetFolderByUID.Error(), Code: 0})
		}
	}

	for _, deleteSecretEntry := range rq.SecretUIDs {
		_, errGetSecretByUID := secretsService.GetSecretByUID(ctx, deleteSecretEntry)
		if errGetSecretByUID != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretByUIDError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetSecretByUID.Error(), Code: 0})
		}
	}

	return Errors
}

func (rq CopyPasteSecretsRQ) Validate(ctx context.Context, secretsService *secrets.SecretsService, Localizer *i18n.Localizer) (Errors []rqrs.Error) {
	for _, copyFolderUID := range rq.FolderUIDs {
		_, errGetFolderByUID := secretsService.GetFolderByUID(ctx, copyFolderUID)
		if errGetFolderByUID != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetFolderByUID.Error(), Code: 0})
		}
	}

	for _, copySecretUID := range rq.SecretUIDs {
		_, errGetSecretByUID := secretsService.GetSecretByUID(ctx, copySecretUID)
		if errGetSecretByUID != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretByUIDError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetSecretByUID.Error(), Code: 0})
		}
	}

	_, errGetFolderFromUID := secretsService.GetFolderByUID(ctx, rq.FromFolderUID)
	if errGetFolderFromUID != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
		Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetFolderFromUID.Error(), Code: 0})
	}

	_, errGetFolderToUID := secretsService.GetFolderByUID(ctx, rq.ToFolderUID)
	if errGetFolderToUID != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
		Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetFolderToUID.Error(), Code: 0})
	}

	return Errors
}

func (rq ExportSecretsRQ) Validate(ctx context.Context, secretsService *secrets.SecretsService, Localizer *i18n.Localizer) (Errors []rqrs.Error) {
	if rq.FolderUID != "" {
		_, errGetFolderByUID := secretsService.GetFolderByUID(ctx, rq.FolderUID)
		if errGetFolderByUID != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetFolderByUID.Error(), Code: 0})
		}
	}

	errSecretsPagination := rq.Pagination.Validate(ctx)
	if errSecretsPagination != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "PaginationError"}})
		Errors = append(Errors, rqrs.Error{Message: msg, Description: errors.Wrap(errSecretsPagination,
			"Secrets pagination validation failed").Error(), Code: 0})
	}

	for _, orderVal := range rq.Order {
		errSecretOrdering := orderVal.Validate(ctx, Localizer)
		if errSecretOrdering != nil {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "OrderError"}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: errors.Wrap(errSecretOrdering,
				"Secret order validation failed").Error(), Code: 0})
		}
	}

	if rq.Format == "" {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "BodyParamMissingError"},
			TemplateData: map[string]interface{}{"Name": "Format"}})
		Errors = append(Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
	} else {
		_, exportFormatExists := ExportFormatsMapInv[rq.Format]
		if !exportFormatExists {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "BodyParamInvalidError"},
				TemplateData: map[string]interface{}{"Name": "Format", "Values": strings.Join([]string{ExportFormatsMap[ExportFormat_DotEnv]}, ",")}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
		}
	}

	if rq.ArchiveType != "" {
		_, exportArchiveTypeExists := ArchiveTypesMapInv[rq.ArchiveType]
		if !exportArchiveTypeExists {
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "BodyParamInvalidError"},
				TemplateData: map[string]interface{}{"Name": "ArchiveType", "Values": strings.Join([]string{ArchiveTypesMap[ArchiveType_Zip],
					ArchiveTypesMap[ArchiveType_Tar]}, ",")}})
			Errors = append(Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
			return Errors
		} else {
			if rq.CompressionType == "" {
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "BodyParamMissingError"},
					TemplateData: map[string]interface{}{"Name": "CompressionType"}})
				Errors = append(Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
			} else {
				_, compressionTypeExists := CompressionTypesMapInv[rq.CompressionType]
				if !compressionTypeExists {
					msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "BodyParamInvalidError"},
						TemplateData: map[string]interface{}{"Name": "CompressionType", "Values": strings.Join([]string{
							strings.Join(CompressionTypesMap[CompressionType_Brotli], ","), strings.Join(CompressionTypesMap[CompressionType_Bzip2], ","),
							strings.Join(CompressionTypesMap[CompressionType_Flate], ","), strings.Join(CompressionTypesMap[CompressionType_Gzip], ","),
							strings.Join(CompressionTypesMap[CompressionType_Lz4], ","), strings.Join(CompressionTypesMap[CompressionType_Lzip], ","),
							strings.Join(CompressionTypesMap[CompressionType_Minlz], ","), strings.Join(CompressionTypesMap[CompressionType_Snappy], ","),
							strings.Join(CompressionTypesMap[CompressionType_XZ], ","), strings.Join(CompressionTypesMap[CompressionType_Zlib], ","),
							strings.Join(CompressionTypesMap[CompressionType_Zstandard], ","),
						}, ", ")}})
					Errors = append(Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
				}
			}
		}
	}

	return Errors
}
