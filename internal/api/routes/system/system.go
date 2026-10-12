// Package routes -----------------------------
// @file      : system.go
// @time      : 2025/5/3 22:17
// -------------------------------------------
package system

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/handlers/system"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterSystemRoutes(api *gin.RouterGroup) {

	systemRoutes := models.RouteGroup{
		Path: "/system",
		Routes: []models.Route{

			{
				Method:      "GET",
				Path:        "version",
				Handler:     system.GetSystemVersion,
				Middlewares: common.WithAuth(),
			},
		},
	}

	common.RegisterRouteGroup(api, systemRoutes)
}
