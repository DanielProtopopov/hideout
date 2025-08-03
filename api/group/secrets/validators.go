package secrets

import (
	"context"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/common/rqrs"
	secrets2 "hideout/internal/secrets"
	"hideout/services/secrets"
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

	_, errGetFolderByUID := secretsService.GetFolderByUID(ctx, rq.FolderUID)
	if errGetFolderByUID != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
		Errors = append(Errors, rqrs.Error{Message: msg, Description: errGetFolderByUID.Error(), Code: 0})
	}

	return Errors
}

func (rq CreateSecretsRQ) Validate(ctx context.Context, secretsService *secrets.SecretsService, Localizer *i18n.Localizer) (Errors []rqrs.Error) {
	if len(rq.Data) == 0 {
		Errors = append(Errors, rqrs.Error{Message: "No data for creation of folder(s)/secret(s) supplied", Description: "", Code: 0})
	}
	for _, createSecretEntry := range rq.Data {
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
	for _, updateSecretEntry := range rq.Data {
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
