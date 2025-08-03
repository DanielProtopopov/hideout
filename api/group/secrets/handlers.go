package secrets

import (
	"context"
	"errors"
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
	"log"
	"net/http"
)

// GetSecretsHandler
// @Summary Getting secrets list
// @Description Getting secrets list
// @ID list-secrets
// @Tags Брокеры
// @Produce json
// @Param params body GetSecretsRQ true "Secrets data"
// @Success 200 {object} GetSecretsRS
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 400 {object} GetSecretsRS
// @Failure 404 {object} GetSecretsRS
// @Failure 500 {object} GetSecretsRS
// @Router /secrets/list/ [post]
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

	validationSpan := sentry.StartSpan(rqContext, "get.secrets.list")
	validationSpan.Description = "rq.validate"

	var request GetSecretsRQ
	response := GetSecretsRS{Secrets: []Secret{}, Folders: []Folder{}, ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}

	secretsSvc, errCreateService := secrets.NewService(rqContext, apiconfig.Settings.Repository, &structs.Folders, &structs.Secrets)
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

	folderByUID, errGetFolder := secretsSvc.GetFolderByUID(rqContext, request.FolderUID)
	if errGetFolder != nil {
		if errors.Is(errGetFolder, apperror.ErrRecordNotFound) {
			log.Printf("Folder with UID of %s was not found", request.FolderUID)
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "FolderNotFoundError"}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
			c.JSON(http.StatusNotFound, response)
			return
		}
		log.Printf("Error fetching folder with UID of %s: %s", request.FolderUID, errGetFolder.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFolderError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolder.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	secretResults, errGetFolderSecrets := secretsSvc.GetSecrets(rqContext, secrets2.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No, Pagination: request.Pagination, Order: request.Order},
		FolderIDs:  []uint{folderByUID.ID},
	})
	if errGetFolderSecrets != nil {
		log.Printf("Error fetching folder secrets: %s", errGetFolderSecrets.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretsError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolderSecrets.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	folderResults, errGetFolderFolders := secretsSvc.GetFolders(rqContext, folders.ListFolderParams{
		ListParams:     generics.ListParams{Deleted: model.No, Pagination: request.Pagination, Order: request.Order},
		ParentFolderID: folderByUID.ID,
	})

	if errGetFolderFolders != nil {
		log.Printf("Error fetching folder' folders: %s", errGetFolderFolders.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetFoldersError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetFolderFolders.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	for _, secret := range secretResults {
		secretEntry := Secret{UID: secret.UID, FolderUID: folderByUID.UID, Name: secret.Name, Value: secret.Value, Type: secret.Type}
		response.Secrets = append(response.Secrets, secretEntry)
	}

	for _, folder := range folderResults {
		folderEntry := Folder{UID: folder.UID, Name: folder.Name, ParentUID: request.FolderUID}
		response.Folders = append(response.Folders, folderEntry)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSecretsHandler
// @Summary Update secrets
// @Description Update secrets
// @ID update-secrets
// @Tags Secrets
// @Produce json
// @Param params body UpdateSecretsRQ true "Secrets to update"
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

	validationSpan := sentry.StartSpan(rqContext, "update.secrets")
	validationSpan.Description = "rq.validate"

	var request UpdateSecretsRQ
	response := UpdateSecretsRS{Data: []Secret{}, ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}

	secretsSvc, errCreateService := secrets.NewService(rqContext, apiconfig.Settings.Repository, &structs.Folders, &structs.Secrets)
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

	prepareSpan := sentry.StartSpan(rqContext, "update.secrets")
	prepareSpan.Description = "prepare"

	// @TODO Implement updating a list of secrets

	c.JSON(http.StatusOK, response)
}

// DeleteSecretsHandler
// @Summary Delete secrets
// @Description Delete secrets
// @ID delete-secrets
// @Tags Secrets
// @Produce json
// @Param params body DeleteSecretsRQ true "Secrets to delete"
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

	validationSpan := sentry.StartSpan(rqContext, "delete.secrets")
	validationSpan.Description = "rq.validate"

	var request DeleteSecretsRQ
	response := DeleteSecretsRS{ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}

	secretsSvc, errCreateService := secrets.NewService(rqContext, apiconfig.Settings.Repository, &structs.Folders, &structs.Secrets)
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

	prepareSpan := sentry.StartSpan(rqContext, "delete.secrets")
	prepareSpan.Description = "prepare"

	// @TODO Implement deleting a list of secrets

	c.JSON(http.StatusOK, response)
}

// CreateSecretsHandler
// @Summary Create secrets
// @Description Create secrets
// @ID create-secrets
// @Tags Secrets
// @Produce json
// @Param params body CreateSecretsRQ true "Secrets to create"
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

	validationSpan := sentry.StartSpan(rqContext, "create.secrets")
	validationSpan.Description = "rq.validate"

	response := CreateSecretsRS{ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}
	var idParam string

	errBindURI := c.ShouldBindUri(&idParam)
	if errBindURI != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestURIMappingError"}})
		response.ResponseListRS.Errors = append(response.ResponseListRS.Errors, rqrs.Error{Message: msg, Description: errBindURI.Error(), Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}
	validationSpan.Finish()

	prepareSpan := sentry.StartSpan(rqContext, "create.secrets")
	prepareSpan.Description = "prepare"

	// @TODO Implement creating a list of secrets

	c.JSON(http.StatusOK, response)
}

// CopyPasteSecretsHandler
// @Summary Copy-paste secrets & folders
// @Description Copy-paste secrets & folders
// @ID copy-paste-secrets
// @Tags Secrets
// @Produce json
// @Param params body CopyPasteSecretsRQ true "Secrets to copy-and-paste"
// @Success 200 {object} CopyPasteSecretsRS
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 400 {object} CopyPasteSecretsRS
// @Failure 404 {object} CopyPasteSecretsRS
// @Failure 500 {object} CopyPasteSecretsRS
// @Router /secrets/copy-paste [put]
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

	validationSpan := sentry.StartSpan(rqContext, "update.secrets")
	validationSpan.Description = "rq.validate"

	response := UpdateSecretsRS{ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}
	var idParam string

	errBindURI := c.ShouldBindUri(&idParam)
	if errBindURI != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestURIMappingError"}})
		response.ResponseListRS.Errors = append(response.ResponseListRS.Errors, rqrs.Error{Message: msg, Description: errBindURI.Error(), Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}
	validationSpan.Finish()

	prepareSpan := sentry.StartSpan(rqContext, "update.secrets")
	prepareSpan.Description = "prepare"

	// @TODO Implement updating a list of secrets

	c.JSON(http.StatusOK, response)
}
