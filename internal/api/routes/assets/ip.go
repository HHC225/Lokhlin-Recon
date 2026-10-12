package assets

import (
	iphandler "recon/internal/api/handlers/assets/ip"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerIPRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/ip",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     iphandler.GetIPAssets,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
