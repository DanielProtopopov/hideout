package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"hideout/api/middleware"
	apiconfig "hideout/cmd/api/config"
	"log"
)

// @title Hideout API
// @version 1.0
// @description API for working with secrets
// @termsOfService https://swagger.io/terms/

// @contact.name API Support
// @contact.url https://www.swagger.io/support
// @contact.email support@hideout.com

// @license.name Private

// @host api.hideout.local
// @BasePath /api/v1
// @query.collection.format multi

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Description for what is this security definition being used

func Serve() {
	route := gin.Default()

	route.Use(middleware.NoCache).Use(LangWithConfig)

	/*
		api.SwaggerInfo.Schemes = []string{"http", "https"}
		api.SwaggerInfo.Host = apiconfig.Settings.SwaggerHost
	*/
	// use ginSwagger middleware to serve the API docs
	route.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	errRun := route.Run(fmt.Sprintf("%s:%d", apiconfig.Settings.Server.Host, apiconfig.Settings.Server.Port))
	log.Panic(errRun)
}

func LangWithConfig(c *gin.Context) {
	middleware.Language(c)
}
