// @title           Recon API
// @version         1.0
// @description     Recon  provides security scanning

// @host      localhost:8082
// @BasePath  /api

// @securityDefinitions.apikey  ApiKeyAuth
// @in                         header
// @name                       Authorization
// @description               Bearer token for authentication

package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"recon/internal/worker"

	"recon/internal/bootstrap"
	"recon/internal/config"
	"recon/internal/scheduler"
	"recon/internal/update"

	"recon/internal/logger"

	_ "recon/internal/database/mongodb"

	_ "recon/internal/database/redis"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"recon/internal/constants"

	_ "recon/internal/bootstrap"

	"recon/internal/api/routes"
	reconmcp "recon/internal/mcp"
	redisLogSubscriber "recon/internal/services/redis_log_subscriber"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"recon/docs"
)

//go:embed static/*
var embeddedFiles embed.FS

func main() {

	if config.GlobalConfig.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	err := os.Setenv("TZ", config.GlobalConfig.System.Timezone)
	if err != nil {
		return
	}
	if constants.Version == "" {
		constants.Version = "dev"
	}
	fmt.Printf("version: %v\n", constants.Version)
	PrintPlugin()
	err = update.Update()
	if err != nil {
		logger.Error("Update failed", zap.Error(err))
	}
	bootstrap.GetProjectList()

	scheduler.InitializeGlobalScheduler()
	scheduler.GetGlobalScheduler().Start()
	defer scheduler.GetGlobalScheduler().Stop()

	go redisLogSubscriber.SubscribeLogChannel()

	router := routes.SetupRouter()
	reconmcp.RegisterRoutes(router)
	gin.DisableConsoleColor()
	//router.Use(logger.GinLogger(), logger.GinRecovery())

	frontendFS, _ := fs.Sub(embeddedFiles, "static")

	assetsFS, _ := fs.Sub(embeddedFiles, "static/assets")

	cssFS, _ := fs.Sub(embeddedFiles, "static/css")
	dracoFS, _ := fs.Sub(embeddedFiles, "static/draco")
	jsFS, _ := fs.Sub(embeddedFiles, "static/js")
	libFS, _ := fs.Sub(embeddedFiles, "static/lib")
	pluginsFS, _ := fs.Sub(embeddedFiles, "static/plugins")
	pngFS, _ := fs.Sub(embeddedFiles, "static/png")

	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Title = constants.DisplayName() + " API"
	router.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.StaticFS("/assets", http.FS(assetsFS))

	router.StaticFS("/css", http.FS(cssFS))
	router.StaticFS("/draco", http.FS(dracoFS))
	router.StaticFS("/js", http.FS(jsFS))
	router.StaticFS("/lib", http.FS(libFS))
	router.StaticFS("/plugins", http.FS(pluginsFS))
	router.StaticFS("/png", http.FS(pngFS))

	router.GET("/map", func(c *gin.Context) {
		data, err := fs.ReadFile(frontendFS, "map.html")
		if err != nil {
			data, err = fs.ReadFile(frontendFS, "index.html")
		}
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read the collection page")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	router.GET("/favicon.ico", func(c *gin.Context) {
		data, err := fs.ReadFile(frontendFS, "favicon.ico")
		if err != nil {
			c.String(http.StatusNotFound, "Favicon not found")
			return
		}
		c.Data(http.StatusOK, "image/x-icon", data)
	})
	router.GET("/icon.svg", func(c *gin.Context) {
		data, err := fs.ReadFile(frontendFS, "icon.svg")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", data)
	})

	router.Static("/images", config.GlobalConfig.System.ImgDir)

	router.GET("/", func(c *gin.Context) {
		data, err := fs.ReadFile(frontendFS, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/mcp") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}

		if strings.HasPrefix(path, "/assets") || strings.HasPrefix(path, "/uploads") ||
			strings.HasPrefix(path, "/css") || strings.HasPrefix(path, "/draco") ||
			strings.HasPrefix(path, "/js") || strings.HasPrefix(path, "/lib") ||
			strings.HasPrefix(path, "/plugins") || strings.HasPrefix(path, "/png") ||
			strings.HasPrefix(path, "/images") {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		data, err := fs.ReadFile(frontendFS, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GlobalConfig.Server.Port),
		Handler: router,
	}

	worker.Run()

	go func() {
		logger.Info("Starting server", zap.Int("port", config.GlobalConfig.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}

func Banner() {
	fmt.Println(constants.DisplayName() + " collection service")
}

func PrintPlugin() {
	fmt.Println("Plugin installation key is stored in PLUGINKEY")
}
