// assets-------------------------------------
// @file      : mp.go
// @time      : 2025/5/14 22:56
// -------------------------------------------

package assets

import (
	"recon/internal/api/handlers/assets/mp"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerMpRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/mp",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     mp.GetMPData,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
