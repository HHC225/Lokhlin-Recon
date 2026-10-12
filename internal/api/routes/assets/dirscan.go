// assets-------------------------------------
// @file      : dirscan.go
// @time      : 2025/5/17 15:20
// -------------------------------------------

package assets

import (
	"recon/internal/api/handlers/assets/dirscan"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerDirscanRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/dirscan",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     dirscan.List,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
