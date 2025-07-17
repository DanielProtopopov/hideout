package secrets

import (
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	apiconfig "hideout/cmd/api/config"
	"hideout/internal/common/rqrs"
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
		log.Printf("Ошибка валидации данных тела запроса: %s", errValidate)
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestValidationError"}})
		response.Errors = append(response.Errors, rqrs.Error{Message: msg, Description: errValidate.Error(), Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}
	validationSpan.Finish()

	// @TODO Implement retrieval of secrets

	c.JSON(http.StatusOK, response)
}

// GetSecretByIDHandler
// @Summary Getting secret by UID
// @Description Getting secret by UID
// @ID get-secret-by-id
// @Tags Secrets
// @Produce json
// @Param ID path integer true "Secret unique identifier"
// @Success 200 {object} GetSecretRS
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 400 {object} GetSecretRS
// @Failure 404 {object} GetSecretRS
// @Failure 500 {object} GetSecretRS
// @Router /secrets/{UID}/ [get]
func GetSecretByIDHandler(c *gin.Context) {
	rqContext := c.Request.Context()
	SentryHub := sentry.GetHubFromContext(rqContext)
	if SentryHub == nil {
		SentryHub = sentry.CurrentHub().Clone()
		rqContext = sentry.SetHubOnContext(rqContext, SentryHub)
	}

	Language, _ := c.Get("Language")
	Localizer := i18n.NewLocalizer(apiconfig.Settings.Bundle, Language.(string))

	validationSpan := sentry.StartSpan(rqContext, "get.secret.by.id")
	validationSpan.Description = "rq.validate"

	response := GetSecretRS{ResponseRS: rqrs.ResponseRS{Errors: []rqrs.Error{}}}
	var idParam string

	errBindURI := c.ShouldBindUri(&idParam)
	if errBindURI != nil {
		msg := Localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "RequestURIMappingError"}})
		response.ResponseRS.Errors = append(response.ResponseRS.Errors, rqrs.Error{Message: msg, Description: errBindURI.Error(), Code: 0})
		c.JSON(http.StatusBadRequest, response)
		return
	}
	validationSpan.Finish()

	prepareSpan := sentry.StartSpan(rqContext, "get.secret.by.id")
	prepareSpan.Description = "prepare"

	// @TODO Implement retrieving secret by UID

	c.JSON(http.StatusOK, response)
}
