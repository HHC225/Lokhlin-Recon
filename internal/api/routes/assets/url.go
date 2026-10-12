// assets-------------------------------------
// @file      : url.go
// @time      : 2025/5/14 23:11
// -------------------------------------------

package assets

import (
	"recon/internal/api/handlers/assets/url"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerUrlRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/url",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     url.GetURLs,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
