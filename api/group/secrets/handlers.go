package secrets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	apiconfig "hideout/cmd/api/config"
	"hideout/internal/common/apperror"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/common/rqrs"
	"hideout/internal/folders"
	secrets2 "hideout/internal/secrets"
	"hideout/services/secrets"
	"hideout/structs"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GetSecretsHandler
// @Summary Getting secrets list
// @Description Getting secrets list
// @ID list-secrets
// @Tags Secrets
// @Produce json
// @Param params body GetSecretsRQ true "Secrets request"
// @Success 200 {object} GetSecretsRS
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 400 {object} GetSecretsRS
// @Failure 404 {object} GetSecretsRS
// @Failure 500 {object} GetSecretsRS
// @Router /secrets/ [post]
func GetSecretsHandler(c *gin.Context) {
	rqContext := c.Request.Context()
	SentryHub := sentry.GetHubFromContext(rqContext)
	if SentryHub == nil {
		SentryHub = sentry.CurrentHub().Clone()
		rqContext = sentry.SetHubOnContext(rqContext, SentryHub)
	}

	Language, _ := c.Get("Language")
	Localizer := i18n.NewLocalizer(apiconfig.Settings.Bundle, Language.(string))

	rqContext = context.WithValue(rqContext, "Sentry", SentryHub)
	rqContext = context.WithValue(rqContext, "Localizer", Localizer)
	rqContext = context.WithValue(rqContext, "Language", Language)

	validationSpan := sentry.StartSpan(rqContext, "validate.get.secrets")
	validationSpan.Description = "rq.validate"

	var request GetSecretsRQ
	response := GetSecretsRS{Secrets: []Secret{}, Folders: []Folder{}, ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}

	secretsSvc, errCreateService := secrets.NewService(rqContext, apiconfig.Settings.SecretsRepository,
		apiconfig.Settings.FoldersRepository, &structs.Folders, &structs.Secrets)
	if errCreateService != nil {
		log.Printf("Error creating secrets service: %s", errCreateService.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CreateSecretsServiceError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errCreateService.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	errBindBody := c.ShouldBindBodyWith(&request, binding.JSON)
	if errBindBody != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestBodyMappingError"}})
		response.ResponseListRS.Errors = append(response.ResponseListRS.Errors, rqrs.Error{Message: msg, Description: errBindBody.Error(), Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}

	errValidate := request.Validate(rqContext, secretsSvc, Localizer)
	if errValidate != nil {
		response.Errors = errValidate
		c.JSON(http.StatusBadRequest, response)
		return
	}
	validationSpan.Finish()

	runSpan := sentry.StartSpan(rqContext, "get.secrets")
	runSpan.Description = "run"

	var parentFolder *folders.Folder = nil
	if request.FolderUID != "" {
		folderByUID, errGetFolder := secretsSvc.GetFolderByUID(rqContext, request.FolderUID)
		if errGetFolder != nil {
			if errors.Is(errGetFolder, apperror.ErrRecordNotFound) {
				log.Printf("Folder with UID of %s was not found", request.FolderUID)
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "FolderNotFoundError"},
					TemplateData: map[string]interface{}{"UID": request.FolderUID}})
				response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
				c.JSON(http.StatusNotFound, response)
				return
			}
			log.Printf("Error fetching folder with UID of %s: %s", request.FolderUID, errGetFolder.Error())
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"},
				TemplateData: map[string]interface{}{"UID": request.FolderUID}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolder.Error(), Code: 0})
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		parentFolder = folderByUID
	}

	listSecretParams := secrets2.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No, Pagination: request.SecretsPagination, Order: request.SecretsOrder},
		IsDynamic:  model.YesOrNo,
	}
	if parentFolder != nil {
		listSecretParams.FolderIDs = append(listSecretParams.IDs, parentFolder.ID)
	}
	secretResults, errGetSecrets := secretsSvc.GetSecrets(rqContext, listSecretParams)
	if errGetSecrets != nil {
		log.Printf("Error fetching secrets: %s", errGetSecrets.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretsError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetSecrets.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	listFolderParams := folders.ListFolderParams{
		ListParams: generics.ListParams{Deleted: model.No, Pagination: request.FoldersPagination, Order: request.FoldersOrder},
	}
	if parentFolder != nil {
		listFolderParams.ParentFolderID = parentFolder.ID
	}

	folderResults, errGetFolders := secretsSvc.GetFolders(rqContext, listFolderParams)

	if errGetFolders != nil {
		log.Printf("Error fetching folders: %s", errGetFolders.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFoldersError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolders.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	for _, secret := range secretResults {
		secretEntry := Secret{ID: secret.ID, UID: secret.UID, Name: secret.Name, Value: secret.Value, Type: secret.Type,
			IsDynamic: secret.IsDynamic}
		if parentFolder != nil {
			secretEntry.FolderUID = parentFolder.UID
		}
		response.Secrets = append(response.Secrets, secretEntry)
	}

	for _, folder := range folderResults {
		folderEntry := Folder{UID: folder.UID, Name: folder.Name, ParentUID: request.FolderUID}
		response.Folders = append(response.Folders, folderEntry)
	}

	runSpan.Finish()

	processSpan := sentry.StartSpan(rqContext, "process.secrets")
	processSpan.Description = "run"
	for secretIndex, _ := range response.Secrets {
		if response.Secrets[secretIndex].IsDynamic {
			value, _, errProcessSecret := response.Secrets[secretIndex].Process(rqContext, secretsSvc)
			if errProcessSecret != nil {
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "DynamicSecretError"}})
				response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errProcessSecret.Error(), Code: 0})
			} else {
				response.Secrets[secretIndex].Value = value
			}
		}
	}

	processSpan.Finish()

	c.JSON(http.StatusOK, response)
}

// UpdateSecretsHandler
// @Summary Update secrets
// @Description Update secrets
// @ID update-secrets
// @Tags Secrets
// @Produce json
// @Param params body UpdateSecretsRQ true "Secrets update request"
// @Success 200 {object} UpdateSecretsRS
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 400 {object} UpdateSecretsRS
// @Failure 404 {object} UpdateSecretsRS
// @Failure 500 {object} UpdateSecretsRS
// @Router /secrets/ [patch]
func UpdateSecretsHandler(c *gin.Context) {
	rqContext := c.Request.Context()
	SentryHub := sentry.GetHubFromContext(rqContext)
	if SentryHub == nil {
		SentryHub = sentry.CurrentHub().Clone()
		rqContext = sentry.SetHubOnContext(rqContext, SentryHub)
	}

	Language, _ := c.Get("Language")
	Localizer := i18n.NewLocalizer(apiconfig.Settings.Bundle, Language.(string))

	rqContext = context.WithValue(rqContext, "Sentry", SentryHub)
	rqContext = context.WithValue(rqContext, "Localizer", Localizer)
	rqContext = context.WithValue(rqContext, "Language", Language)

	validationSpan := sentry.StartSpan(rqContext, "validate.update.secrets")
	validationSpan.Description = "rq.validate"

	var request UpdateSecretsRQ
	response := UpdateSecretsRS{Data: []Secret{}, ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}

	secretsSvc, errCreateService := secrets.NewService(rqContext, apiconfig.Settings.SecretsRepository,
		apiconfig.Settings.FoldersRepository, &structs.Folders, &structs.Secrets)
	if errCreateService != nil {
		log.Printf("Error creating secrets service: %s", errCreateService.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CreateSecretsServiceError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errCreateService.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	errBindBody := c.ShouldBindBodyWith(&request, binding.JSON)
	if errBindBody != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestBodyMappingError"}})
		response.ResponseListRS.Errors = append(response.ResponseListRS.Errors, rqrs.Error{Message: msg, Description: errBindBody.Error(), Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}

	errValidate := request.Validate(rqContext, secretsSvc, Localizer)
	if errValidate != nil {
		response.Errors = errValidate
		c.JSON(http.StatusBadRequest, response)
		return
	}
	validationSpan.Finish()

	runSpan := sentry.StartSpan(rqContext, "update.secrets")
	runSpan.Description = "run"

	for _, updateSecretEntry := range request.Data {
		folderByUID, errGetFolderByUID := secretsSvc.GetFolderByUID(rqContext, updateSecretEntry.FolderUID)
		if errGetFolderByUID != nil {
			log.Printf("Error retrieving folder with UID of %s: %s", updateSecretEntry.UID, errGetFolderByUID.Error())
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolderByUID.Error(), Code: 0})
			continue
		}
		updatedSecret, errUpdateSecret := secretsSvc.UpdateSecret(rqContext, secrets2.Secret{
			FolderID: folderByUID.ID, Name: updateSecretEntry.Name, Value: updateSecretEntry.Value,
			Type: updateSecretEntry.Type, IsDynamic: updateSecretEntry.IsDynamic,
		})
		if errUpdateSecret != nil {
			log.Printf("Error updating secret with UID of %s: %s", updateSecretEntry.UID, errUpdateSecret.Error())
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "UpdateSecretError"},
				TemplateData: map[string]interface{}{"UID": updateSecretEntry.UID}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errUpdateSecret.Error(), Code: 0})
			continue
		}
		response.Data = append(response.Data, Secret{
			ID: updatedSecret.ID, UID: updatedSecret.UID, FolderUID: folderByUID.UID, Name: updatedSecret.Name,
			Value: updatedSecret.Value, Type: updatedSecret.Type, IsDynamic: updatedSecret.IsDynamic,
		})
	}

	processSpan := sentry.StartSpan(rqContext, "process.secrets")
	processSpan.Description = "run"
	for secretIndex, _ := range response.Data {
		if response.Data[secretIndex].IsDynamic {
			value, _, errProcessSecret := response.Data[secretIndex].Process(rqContext, secretsSvc)
			if errProcessSecret != nil {
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "DynamicSecretError"}})
				response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errProcessSecret.Error(), Code: 0})
			} else {
				response.Data[secretIndex].Value = value
			}
		}
	}

	processSpan.Finish()
	runSpan.Finish()

	c.JSON(http.StatusOK, response)
}

// DeleteSecretsHandler
// @Summary Delete secrets
// @Description Delete secrets
// @ID delete-secrets
// @Tags Secrets
// @Produce json
// @Param params body DeleteSecretsRQ true "Secrets delete request"
// @Success 200 {object} DeleteSecretsRS
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 400 {object} DeleteSecretsRS
// @Failure 404 {object} DeleteSecretsRS
// @Failure 500 {object} DeleteSecretsRS
// @Router /secrets/ [delete]
func DeleteSecretsHandler(c *gin.Context) {
	rqContext := c.Request.Context()
	SentryHub := sentry.GetHubFromContext(rqContext)
	if SentryHub == nil {
		SentryHub = sentry.CurrentHub().Clone()
		rqContext = sentry.SetHubOnContext(rqContext, SentryHub)
	}

	Language, _ := c.Get("Language")
	Localizer := i18n.NewLocalizer(apiconfig.Settings.Bundle, Language.(string))

	rqContext = context.WithValue(rqContext, "Sentry", SentryHub)
	rqContext = context.WithValue(rqContext, "Localizer", Localizer)
	rqContext = context.WithValue(rqContext, "Language", Language)

	validationSpan := sentry.StartSpan(rqContext, "validate.delete.secrets")
	validationSpan.Description = "rq.validate"

	var request DeleteSecretsRQ
	response := DeleteSecretsRS{ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}

	secretsSvc, errCreateService := secrets.NewService(rqContext, apiconfig.Settings.SecretsRepository,
		apiconfig.Settings.FoldersRepository, &structs.Folders, &structs.Secrets)
	if errCreateService != nil {
		log.Printf("Error creating secrets service: %s", errCreateService.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CreateSecretsServiceError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errCreateService.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	errBindBody := c.ShouldBindBodyWith(&request, binding.JSON)
	if errBindBody != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestBodyMappingError"}})
		response.ResponseListRS.Errors = append(response.ResponseListRS.Errors, rqrs.Error{Message: msg, Description: errBindBody.Error(), Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}

	errValidate := request.Validate(rqContext, secretsSvc, Localizer)
	if errValidate != nil {
		response.Errors = errValidate
		c.JSON(http.StatusBadRequest, response)
		return
	}
	validationSpan.Finish()

	runSpan := sentry.StartSpan(rqContext, "delete.secrets")
	runSpan.Description = "run"

	for _, deleteSecretUID := range request.SecretUIDs {
		secretByUID, errGetSecretByUID := secretsSvc.GetSecretByUID(rqContext, deleteSecretUID)
		if errGetSecretByUID != nil {
			log.Printf("Error retrieving secret with UID of %s: %s", deleteSecretUID, errGetSecretByUID.Error())
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretByUIDError"},
				TemplateData: map[string]interface{}{"UID": deleteSecretUID}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetSecretByUID.Error(), Code: 0})
			continue
		}
		errDeleteSecret := secretsSvc.DeleteSecret(rqContext, secretByUID.ID, false)
		if errDeleteSecret != nil {
			log.Printf("Error deleting secret with UID of %s: %s", deleteSecretUID, errDeleteSecret.Error())
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "DeleteSecretError"},
				TemplateData: map[string]interface{}{"UID": deleteSecretUID}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errDeleteSecret.Error(), Code: 0})
		}
	}

	for _, deleteFolderUID := range request.FolderUIDs {
		folderByUID, errGetFolderByUID := secretsSvc.GetFolderByUID(rqContext, deleteFolderUID)
		if errGetFolderByUID != nil {
			log.Printf("Error retrieving folder with UID of %s: %s", deleteFolderUID, errGetFolderByUID.Error())
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolderByUID.Error(), Code: 0})
			continue
		}
		errDeleteFolder := secretsSvc.DeleteFolder(rqContext, folderByUID.ID, false)
		if errDeleteFolder != nil {
			log.Printf("Error retrieving folder with UID of %s: %s", deleteFolderUID, errDeleteFolder.Error())
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "DeleteFolderError"},
				TemplateData: map[string]interface{}{"UID": deleteFolderUID}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errDeleteFolder.Error(), Code: 0})
		}
	}

	runSpan.Finish()
	c.JSON(http.StatusOK, response)
}

// CreateSecretsHandler
// @Summary Create secrets
// @Description Create secrets
// @ID create-secrets
// @Tags Secrets
// @Produce json
// @Param params body CreateSecretsRQ true "Secrets create request"
// @Success 200 {object} CreateSecretsRS
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 400 {object} CreateSecretsRS
// @Failure 404 {object} CreateSecretsRS
// @Failure 500 {object} CreateSecretsRS
// @Router /secrets/ [put]
func CreateSecretsHandler(c *gin.Context) {
	rqContext := c.Request.Context()
	SentryHub := sentry.GetHubFromContext(rqContext)
	if SentryHub == nil {
		SentryHub = sentry.CurrentHub().Clone()
		rqContext = sentry.SetHubOnContext(rqContext, SentryHub)
	}

	Language, _ := c.Get("Language")
	Localizer := i18n.NewLocalizer(apiconfig.Settings.Bundle, Language.(string))

	rqContext = context.WithValue(rqContext, "Sentry", SentryHub)
	rqContext = context.WithValue(rqContext, "Localizer", Localizer)
	rqContext = context.WithValue(rqContext, "Language", Language)

	validationSpan := sentry.StartSpan(rqContext, "validate.create.secrets")
	validationSpan.Description = "rq.validate"

	var request CreateSecretsRQ
	response := CreateSecretsRS{ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}

	secretsSvc, errCreateService := secrets.NewService(rqContext, apiconfig.Settings.SecretsRepository,
		apiconfig.Settings.FoldersRepository, &structs.Folders, &structs.Secrets)
	if errCreateService != nil {
		log.Printf("Error creating secrets service: %s", errCreateService.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CreateSecretsServiceError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errCreateService.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	errBindBody := c.ShouldBindBodyWith(&request, binding.JSON)
	if errBindBody != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestBodyMappingError"}})
		response.ResponseListRS.Errors = append(response.ResponseListRS.Errors, rqrs.Error{Message: msg, Description: errBindBody.Error(), Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}

	errValidate := request.Validate(rqContext, secretsSvc, Localizer)
	if errValidate != nil {
		response.Errors = errValidate
		c.JSON(http.StatusBadRequest, response)
		return
	}
	validationSpan.Finish()

	runSpan := sentry.StartSpan(rqContext, "create.secrets")
	runSpan.Description = "run"

	for _, secretToCreate := range request.Data {
		folderByUID, errGetFolder := secretsSvc.GetFolderByUID(rqContext, secretToCreate.FolderUID)
		if errGetFolder != nil {
			if errors.Is(errGetFolder, apperror.ErrRecordNotFound) {
				log.Printf("Folder with UID of %s was not found", secretToCreate.FolderUID)
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "FolderNotFoundError"},
					TemplateData: map[string]interface{}{"UID": secretToCreate.FolderUID}})
				response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
			} else {
				log.Printf("Error fetching folder with UID of %s: %s", secretToCreate.FolderUID, errGetFolder.Error())
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
				response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolder.Error(), Code: 0})
			}
			continue
		}
		newSecret, errCreateSecret := secretsSvc.CreateSecret(rqContext, secrets2.Secret{
			FolderID: folderByUID.ID, UID: gofakeit.UUID(), Name: secretToCreate.Name,
			Value: secretToCreate.Value, Type: secretToCreate.Type, IsDynamic: secretToCreate.IsDynamic,
		})
		if errCreateSecret != nil {
			log.Printf("Error creating secret with name of %s: %s", secretToCreate.Name, errCreateSecret.Error())
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CreateSecretError"}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errCreateSecret.Error(), Code: 0})
			continue
		}
		response.Data = append(response.Data, Secret{
			ID: newSecret.ID, UID: newSecret.UID, FolderUID: folderByUID.UID, Name: newSecret.Name,
			Value: newSecret.Value, Type: newSecret.Type, IsDynamic: newSecret.IsDynamic,
		})
	}

	processSpan := sentry.StartSpan(rqContext, "process.secrets")
	processSpan.Description = "run"
	for secretIndex, _ := range response.Data {
		if response.Data[secretIndex].IsDynamic {
			value, _, errProcessSecret := response.Data[secretIndex].Process(rqContext, secretsSvc)
			if errProcessSecret != nil {
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "DynamicSecretError"}})
				response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errProcessSecret.Error(), Code: 0})
			} else {
				response.Data[secretIndex].Value = value
			}
		}
	}

	processSpan.Finish()

	runSpan.Finish()
	c.JSON(http.StatusOK, response)
}

// CopyPasteSecretsHandler
// @Summary Copy-paste secrets & folders
// @Description Copy-paste secrets & folders
// @ID copy-paste-secrets
// @Tags Secrets
// @Produce json
// @Param params body CopyPasteSecretsRQ true "Secrets copy-and-paste request"
// @Success 200 {object} CopyPasteSecretsRS
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 400 {object} CopyPasteSecretsRS
// @Failure 404 {object} CopyPasteSecretsRS
// @Failure 500 {object} CopyPasteSecretsRS
// @Router /secrets/copy-paste/ [put]
func CopyPasteSecretsHandler(c *gin.Context) {
	rqContext := c.Request.Context()
	SentryHub := sentry.GetHubFromContext(rqContext)
	if SentryHub == nil {
		SentryHub = sentry.CurrentHub().Clone()
		rqContext = sentry.SetHubOnContext(rqContext, SentryHub)
	}

	Language, _ := c.Get("Language")
	Localizer := i18n.NewLocalizer(apiconfig.Settings.Bundle, Language.(string))

	rqContext = context.WithValue(rqContext, "Sentry", SentryHub)
	rqContext = context.WithValue(rqContext, "Localizer", Localizer)
	rqContext = context.WithValue(rqContext, "Language", Language)

	validationSpan := sentry.StartSpan(rqContext, "validate.update.secrets")
	validationSpan.Description = "rq.validate"

	var request CopyPasteSecretsRQ
	response := CopyPasteSecretsRS{Folders: []Folder{}, Secrets: []Secret{}, ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}

	secretsSvc, errCreateService := secrets.NewService(rqContext, apiconfig.Settings.SecretsRepository,
		apiconfig.Settings.FoldersRepository, &structs.Folders, &structs.Secrets)
	if errCreateService != nil {
		log.Printf("Error creating secrets service: %s", errCreateService.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CreateSecretsServiceError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errCreateService.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	errBindBody := c.ShouldBindBodyWith(&request, binding.JSON)
	if errBindBody != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestBodyMappingError"}})
		response.ResponseListRS.Errors = append(response.ResponseListRS.Errors, rqrs.Error{Message: msg, Description: errBindBody.Error(), Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}

	errValidate := request.Validate(rqContext, secretsSvc, Localizer)
	if errValidate != nil {
		response.Errors = errValidate
		c.JSON(http.StatusBadRequest, response)
		return
	}
	validationSpan.Finish()

	runSpan := sentry.StartSpan(rqContext, "copy.paste.secrets")
	runSpan.Description = "run"

	secretsList, errGetSecrets := secretsSvc.GetSecrets(rqContext, secrets2.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No, UIDs: request.SecretUIDs},
	})
	if errGetSecrets != nil {
		log.Printf("Error retrieving secrets: %s", errGetSecrets.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretsError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetSecrets.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	foldersList, errGetFolders := secretsSvc.GetFolders(rqContext, folders.ListFolderParams{
		ListParams: generics.ListParams{Deleted: model.No, UIDs: request.FolderUIDs},
	})
	if errGetFolders != nil {
		log.Printf("Error retrieving folders: %s", errGetFolders.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFoldersError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolders.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	folderFrom, errGetFolderFromUID := secretsSvc.GetFolderByUID(rqContext, request.FromFolderUID)
	if errGetFolderFromUID != nil {
		log.Printf("Error retrieving folder (From) with UID of %s: %s", request.FromFolderUID, errGetFolderFromUID.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolderFromUID.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	folderTo, errGetFolderToUID := secretsSvc.GetFolderByUID(rqContext, request.ToFolderUID)
	if errGetFolderToUID != nil {
		log.Printf("Error retrieving folder (To) with UID of %s: %s", request.ToFolderUID, errGetFolderToUID.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolderToUID.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	copiedFolders, copiedSecrets, errCopy := secretsSvc.Copy(rqContext, foldersList, secretsList, folderFrom.ID, folderTo.ID)
	if errCopy != nil {
		log.Printf("Error copying secret(s) and/or folder(s) from folder with UID %s to folder with UID of %s: %s",
			request.FromFolderUID, request.ToFolderUID, errCopy.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CopySecretsFoldersError"},
			TemplateData: map[string]interface{}{"UID": folderTo.UID}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errCopy.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	for _, copiedSecret := range copiedSecrets {
		copiedSecretFolder, errGetFolderByID := secretsSvc.GetFolderByID(rqContext, copiedSecret.FolderID)
		if errGetFolderByID != nil {
			log.Printf("Error retrieving folder of copied secret: %s", errGetFolderByID.Error())
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretFolderError"},
				TemplateData: map[string]interface{}{"UID": copiedSecret.UID}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolderByID.Error(), Code: 0})
			continue
		}
		response.Secrets = append(response.Secrets, Secret{
			ID: copiedSecret.ID, UID: copiedSecret.UID, FolderUID: copiedSecretFolder.UID,
			Name: copiedSecret.Name, Value: copiedSecret.Value, Type: copiedSecret.Type, IsDynamic: copiedSecret.IsDynamic,
		})
	}

	for _, copiedFolder := range copiedFolders {
		copiedFolderParent, errGetFolderByID := secretsSvc.GetFolderByID(rqContext, copiedFolder.ParentID)
		if errGetFolderByID != nil {
			log.Printf("Error retrieving folder with ID of %d: %s", copiedFolder.ParentID, errGetFolderByID.Error())
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByIDError"}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolderByID.Error(), Code: 0})
			continue
		}
		response.Folders = append(response.Folders, Folder{
			ID: copiedFolder.ID, UID: copiedFolder.UID, ParentUID: copiedFolderParent.UID, Name: copiedFolder.Name,
		})
	}

	processSpan := sentry.StartSpan(rqContext, "process.secrets")
	processSpan.Description = "run"
	for secretIndex, _ := range response.Secrets {
		if response.Secrets[secretIndex].IsDynamic {
			value, _, errProcessSecret := response.Secrets[secretIndex].Process(rqContext, secretsSvc)
			if errProcessSecret != nil {
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "DynamicSecretError"}})
				response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errProcessSecret.Error(), Code: 0})
			} else {
				response.Secrets[secretIndex].Value = value
			}
		}
	}

	processSpan.Finish()

	runSpan.Finish()
	c.JSON(http.StatusOK, response)
}

// ExportSecretsHandler
// @Summary Export secrets into various formats
// @Description Export secrets into various formats
// @ID export-secrets
// @Tags Secrets
// @Param params body ExportSecretsRQ true "Secrets export request"
// @Produce application/octet-stream
// @Success 200 {string} string ""
// @Failure 401 {string} string ""
// @Failure 404 {string} string ""
// @Failure 500 {string} string ""
// @Router /secrets/export/ [post]
func ExportSecretsHandler(c *gin.Context) {
	rqContext := c.Request.Context()
	SentryHub := sentry.GetHubFromContext(rqContext)
	if SentryHub == nil {
		SentryHub = sentry.CurrentHub().Clone()
		rqContext = sentry.SetHubOnContext(rqContext, SentryHub)
	}

	Language, _ := c.Get("Language")
	Localizer := i18n.NewLocalizer(apiconfig.Settings.Bundle, Language.(string))

	rqContext = context.WithValue(rqContext, "Sentry", SentryHub)
	rqContext = context.WithValue(rqContext, "Localizer", Localizer)
	rqContext = context.WithValue(rqContext, "Language", Language)

	validationSpan := sentry.StartSpan(rqContext, "validate.get.secrets")
	validationSpan.Description = "rq.validate"

	var request ExportSecretsRQ
	response := ExportSecretsRS{Secrets: []Secret{}, ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}

	secretsSvc, errCreateService := secrets.NewService(rqContext, apiconfig.Settings.SecretsRepository,
		apiconfig.Settings.FoldersRepository, &structs.Folders, &structs.Secrets)
	if errCreateService != nil {
		log.Printf("Error creating secrets service: %s", errCreateService.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CreateSecretsServiceError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	errBindBody := c.ShouldBindBodyWith(&request, binding.JSON)
	if errBindBody != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestBodyMappingError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}

	errValidate := request.Validate(rqContext, secretsSvc, Localizer)
	if errValidate != nil {
		response.Errors = errValidate
		c.JSON(http.StatusBadRequest, response)
		return
	}
	validationSpan.Finish()

	runSpan := sentry.StartSpan(rqContext, "get.secrets")
	runSpan.Description = "run"

	var parentFolder *folders.Folder = nil
	if request.FolderUID != "" {
		folderByUID, errGetFolder := secretsSvc.GetFolderByUID(rqContext, request.FolderUID)
		if errGetFolder != nil {
			if errors.Is(errGetFolder, apperror.ErrRecordNotFound) {
				log.Printf("Folder with UID of %s was not found", request.FolderUID)
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "FolderNotFoundError"},
					TemplateData: map[string]interface{}{"UID": request.FolderUID}})
				response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
				c.JSON(http.StatusNotFound, response)
				return
			}
			log.Printf("Error fetching folder with UID of %s: %s", request.FolderUID, errGetFolder.Error())
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderByUIDError"},
				TemplateData: map[string]interface{}{"UID": request.FolderUID}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		parentFolder = folderByUID
	}

	listSecretParams := secrets2.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No, Pagination: request.Pagination, Order: request.Order},
		IsDynamic:  model.YesOrNo,
	}
	if parentFolder != nil {
		listSecretParams.FolderIDs = append(listSecretParams.IDs, parentFolder.ID)
	}
	secretResults, errGetSecrets := secretsSvc.GetSecrets(rqContext, listSecretParams)
	if errGetSecrets != nil {
		log.Printf("Error fetching secrets: %s", errGetSecrets.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretsError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	for _, secret := range secretResults {
		secretEntry := Secret{ID: secret.ID, UID: secret.UID, Name: secret.Name, Value: secret.Value, Type: secret.Type,
			IsDynamic: secret.IsDynamic}
		if parentFolder != nil {
			secretEntry.FolderUID = parentFolder.UID
		}
		response.Secrets = append(response.Secrets, secretEntry)
	}

	runSpan.Finish()

	processSpan := sentry.StartSpan(rqContext, "process.secrets")
	processSpan.Description = "run"
	for secretIndex, _ := range response.Secrets {
		if response.Secrets[secretIndex].IsDynamic {
			value, _, errProcessSecret := response.Secrets[secretIndex].Process(rqContext, secretsSvc)
			if errProcessSecret != nil {
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "DynamicSecretError"}})
				response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errProcessSecret.Error(), Code: 0})
			} else {
				response.Secrets[secretIndex].Value = value
			}
		}
	}

	processSpan.Finish()

	exportSpan := sentry.StartSpan(rqContext, "export.secrets")
	exportSpan.Description = "run"
	var exportedData []byte
	switch request.Format {
	case ExportFormatsMap[ExportFormat_DotEnv]:
		{
			exportedDataVal, errExportToDotEnv := ExportToDotEnv(rqContext, response.Secrets)
			if errExportToDotEnv != nil {
				msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "ExportSecretsError"}})
				response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errExportToDotEnv.Error(), Code: 0})
				c.JSON(http.StatusInternalServerError, response)
				return
			}
			exportedData = []byte(exportedDataVal)
		}
	}

	archiveType, _ := ArchiveTypesMapInv[request.ArchiveType]
	compressionType, _ := CompressionTypesMapInv[request.CompressionType]
	exportType, _ := ExportFormatsMapInv[request.Format]
	if archiveType == ArchiveType_None {
		var uuid = gofakeit.UUID()
		exportTypeVal, _ := ExportExtensionsMap[exportType]
		secretsFile := fmt.Sprintf("secrets-%s%s", strings.ReplaceAll(uuid, "-", ""), exportTypeVal)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", secretsFile))
		c.Data(http.StatusOK, "application/octet-stream", exportedData)
		return
	}

	archiveFilename, errArchiveData := ArchiveExport(rqContext, exportedData, archiveType, compressionType, exportType)
	if errArchiveData != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "ArchiveSecretsError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errArchiveData.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	archiveFile, errOpenSecretsArchive := os.Open(archiveFilename)
	if errOpenSecretsArchive != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "OpenSecretsArchiveError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errOpenSecretsArchive.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	defer func() {
		archiveFile.Close()
		os.Remove(archiveFilename)
	}()

	var buffer bytes.Buffer
	_, errReadArchiveSecretsFile := io.Copy(&buffer, archiveFile)
	if errReadArchiveSecretsFile != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "ReadSecretsArchiveError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errReadArchiveSecretsFile.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(archiveFilename)))
	c.Data(http.StatusOK, "application/octet-stream", buffer.Bytes())

	exportSpan.Finish()
}
