package routes

import (
	"strings"

	"github.com/gin-gonic/gin"
	"recon/internal/api/middleware"
	"recon/internal/api/routes/assets"
	confRoutes "recon/internal/api/routes/configuration"
	dictRoutes "recon/internal/api/routes/dictionary"
	"recon/internal/api/routes/export"
	"recon/internal/api/routes/fingerprint"
	"recon/internal/api/routes/plugin"
	"recon/internal/api/routes/poc"
	"recon/internal/api/routes/project"
	sroutes "recon/internal/api/routes/sensitive_rule"
	"recon/internal/api/routes/system"
	"recon/internal/api/routes/task"
	"recon/internal/api/routes/user"

	apikeyRoutes "recon/internal/api/routes/apikey"
	"recon/internal/api/routes/node"
)

func loggerWithSkipPaths(skipPrefixes []string) gin.HandlerFunc {

	defaultLogger := gin.Logger()

	return func(c *gin.Context) {

		path := c.Request.URL.Path
		for _, prefix := range skipPrefixes {
			if strings.HasPrefix(path, prefix) {

				c.Next()
				return
			}
		}

		defaultLogger(c)
	}
}

func SetupRouter() *gin.Engine {
	router := gin.New()

	router.Use(loggerWithSkipPaths([]string{"/assets", "/uploads", "/images"}))

	router.Use(gin.Recovery())

	router.Use(middleware.I18nMiddleware())

	api := router.Group("/api")
	{

		user.RegisterUserRoutes(api)

		system.RegisterSystemRoutes(api)

		confRoutes.RegisterConfigurationRoutes(api)

		assets.RegisterAssetsRoutes(api)

		fingerprint.RegisterFingerprintRoutes(api)

		sroutes.RegisterSensitiveRuleRoutes(api)

		node.RegisterNodeRoutes(api)

		task.RegisterTaskRoutes(api)

		project.RegisterProjectRoutes(api)

		poc.RegisterPocRoutes(api)

		export.RegisterExportRoutes(api)

		plugin.RegisterPluginRoutes(api)

		dictRoutes.RegisterDictionaryRoutes(api)

		apikeyRoutes.RegisterApiKeyRoutes(api)
	}

	return router
}
