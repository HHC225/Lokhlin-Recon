// assets-------------------------------------
// @file      : subdomain.go
// @time      : 2025/5/13 23:32
// -------------------------------------------

package assets

import (
	"recon/internal/api/handlers/assets/subdomain"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerSubDomainRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/subdomain",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     subdomain.GetSubdomains,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/taker",
				Handler:     subdomain.GetSubdomainTakerData,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
