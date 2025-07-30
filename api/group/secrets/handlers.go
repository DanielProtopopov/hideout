package secrets

import (
	"errors"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	apiconfig "hideout/cmd/api/config"
	"hideout/internal/common/apperror"
	"hideout/internal/common/rqrs"
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

	validationSpan := sentry.StartSpan(rqContext, "get.secrets.list")
	validationSpan.Description = "rq.validate"

	var request GetSecretsRQ
	response := GetSecretsRS{Secrets: []Secret{}, Paths: []Path{}, ResponseListRS: rqrs.ResponseListRS{Errors: []rqrs.Error{}}}

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

	secretsSvc, errCreateService := secrets.NewService(rqContext, secrets.Config{}, &structs.Paths, &structs.Secrets, secrets.RepositoryType_Redis, false)
	if errCreateService != nil {
		log.Printf("Error creating secrets service: %s", errCreateService.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "CreateSecretsServiceError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errCreateService.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	pathByUID, errGetPath := secretsSvc.GetPathByUID(rqContext, request.PathUID)
	if errGetPath != nil {
		if errors.Is(errGetPath, apperror.ErrRecordNotFound) {
			log.Printf("Path with UID of %s was not found", request.PathUID)
			msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "PathNotFoundError"}})
			response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: msg, Code: 0})
			c.JSON(http.StatusNotFound, response)
			return
		}
		log.Printf("Error fetching path with UID of %s: %s", request.PathUID, errGetPath.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetPathError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetPath.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	secretResults, errGetPathSecrets := secretsSvc.GetSecrets(rqContext, pathByUID.ID)
	if errGetPathSecrets != nil {
		log.Printf("Error fetching path secrets: %s", errGetPathSecrets.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetSecretsError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetPathSecrets.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	pathResults, errGetPathPaths := secretsSvc.GetPaths(rqContext, pathByUID.ID)
	if errGetPathPaths != nil {
		log.Printf("Error fetching path' paths: %s", errGetPathPaths.Error())
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "GetPathsError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errGetPathPaths.Error(), Code: 0})
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	for _, secret := range secretResults {
		secretEntry := Secret{UID: secret.UID, PathUID: pathByUID.UID, Name: secret.Name, Value: secret.Value, Type: secret.Type}
		response.Secrets = append(response.Secrets, secretEntry)
	}

	for _, path := range pathResults {
		pathEntry := Path{UID: path.UID, Name: path.Name, ParentUID: request.PathUID}
		response.Paths = append(response.Paths, pathEntry)
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
