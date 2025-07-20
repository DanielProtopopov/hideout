package secrets

import (
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	apiconfig "hideout/cmd/api/config"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/common/pagination"
	"hideout/internal/common/rqrs"
	"hideout/internal/pkg/extra"
	"hideout/internal/secrets"
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

	validationSpan := sentry.StartSpan(rqContext, "get.secrets.list")
	validationSpan.Description = "rq.validate"

	var request GetSecretsRQ
	response := GetSecretsRS{Data: []Secret{}, ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}

	errBindBody := c.ShouldBindBodyWith(&request, binding.JSON)
	if errBindBody != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestBodyMappingError"}})
		response.ResponseListRS.Errors = append(response.ResponseListRS.Errors, rqrs.Error{Message: msg, Description: errBindBody.Error(), Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}

	errValidate := request.Validate(rqContext, Localizer)
	if errValidate != nil {
		log.Printf("Error validating body of the request: %s", errValidate)
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestValidationError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errValidate.Error(), Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}
	validationSpan.Finish()

	inMemoryRepository := secrets.NewInMemoryRepository(structs.Secrets)
	secretResults, errGetSecrets := inMemoryRepository.GetMapByUID(rqContext, secrets.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No}, Path: request.Path, Name: request.Name, Types: request.Types,
	})
	if errGetSecrets != nil {
		log.Printf("Error fetching secrets: %s", errGetSecrets.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretsError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetSecrets.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	var parentIDs []uint
	for _, secret := range secretResults {
		if secret.ParentID != 0 {
			parentIDs = append(parentIDs, secret.ParentID)
		}
	}
	if len(parentIDs) > 0 {
		parentIDs = extra.UniqueUint(parentIDs)
	}
	parentSecretsMap, errGetParentSecrets := inMemoryRepository.GetMapByID(rqContext, secrets.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No, Order: request.Ordering, Pagination: pagination.Pagination{
			PerPage: request.Pagination.PerPage, Page: request.Pagination.Page,
		}}, Path: request.Path, Name: request.Name, Types: request.Types,
	})
	if errGetParentSecrets != nil {
		log.Printf("Error fetching parent secrets: %s", errGetParentSecrets.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretsError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetParentSecrets.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	for _, secret := range secretResults {
		secretEntry := Secret{
			UID: secret.UID, ParentUID: "", Path: secret.Path, Name: secret.Name, Value: secret.Value, Type: secret.Type,
		}
		if secret.ParentID != 0 {
			parentSecret, exists := parentSecretsMap[secret.ParentID]
			if exists {
				secretEntry.ParentUID = parentSecret.UID
			}
		}
		response.Data = append(response.Data, secretEntry)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSecretsHandler
// @Summary Update secrets
// @Description Update secrets
// @ID update-secrets
// @Tags Secrets
// @Produce json
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

// DeleteSecretsHandler
// @Summary Delete secrets
// @Description Delete secrets
// @ID delete-secrets
// @Tags Secrets
// @Produce json
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

	validationSpan := sentry.StartSpan(rqContext, "delete.secrets")
	validationSpan.Description = "rq.validate"

	response := DeleteSecretsRS{ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}
	var idParam string

	errBindURI := c.ShouldBindUri(&idParam)
	if errBindURI != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestURIMappingError"}})
		response.ResponseListRS.Errors = append(response.ResponseListRS.Errors, rqrs.Error{Message: msg, Description: errBindURI.Error(), Code: 0})
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
