package apikey

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/handlers/apikey"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterApiKeyRoutes(api *gin.RouterGroup) {
	routes := models.RouteGroup{
		Path: "/apikey",
		Routes: []models.Route{
			{
				Method:      "GET",
				Path:        "/list",
				Handler:     apikey.List,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/create",
				Handler:     apikey.Create,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/delete",
				Handler:     apikey.Delete,
				Middlewares: common.WithAuth(),
			},
		},
	}
	common.RegisterRouteGroup(api, routes)
}
