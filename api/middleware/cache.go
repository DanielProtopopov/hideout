package middleware

import (
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

// NoCache Adding Cache-Control header to avoid caching
func NoCache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache")

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetRequest(c.Request)
	})
}
