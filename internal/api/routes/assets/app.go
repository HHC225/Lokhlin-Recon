// assets-------------------------------------
// @file      : app.go
// @time      : 2025/5/14 22:41
// -------------------------------------------

package assets

import (
	"recon/internal/api/handlers/assets/app"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerAppRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/app",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     app.GetAppData,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
