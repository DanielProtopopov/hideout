package middleware

import (
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

// NoCache добавление заголовка об отмене кэширования
func NoCache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache")

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetRequest(c.Request)
	})
}
