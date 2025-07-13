package middleware

import (
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"hideout/internal/translations"
)

func Language(c *gin.Context) {
	requestLanguage := translations.DefaultLanguage

	lang, exists := c.GetQuery("lang")
	if !exists || lang == "" {
		requestLanguage = translations.DefaultLanguage
	} else {
		for _, availLang := range translations.Languages {
			if availLang == lang {
				requestLanguage = availLang
				break
			}
		}
	}

	c.Set("Language", requestLanguage)

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("Language", requestLanguage)
	})
}
