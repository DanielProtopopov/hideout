package public

import (
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetSitemapHandler
// @Summary Получение sitemap
// @Description Получение sitemap
// @ID public-get-sitemap
// @Tags Общедоступные методы
// @Produce json
// @Success 200 {} string string
// @Failure 400 {} string string
// @Failure 404 {} string string
// @Failure 500 {} string string
// @Router /public/sitemap/ [get]
func GetSitemapHandler(c *gin.Context) {
	rqContext := c.Request.Context()
	SentryHub := sentry.GetHubFromContext(rqContext)
	if SentryHub == nil {
		SentryHub = sentry.CurrentHub().Clone()
		rqContext = sentry.SetHubOnContext(rqContext, SentryHub)
	}

	// Language, _ := c.Get("Language")
	// Localizer := i18n.NewLocalizer(apiconfig.Settings.Bundle, Language.(string))

	c.HTML(http.StatusOK, "", gin.H{})
}
