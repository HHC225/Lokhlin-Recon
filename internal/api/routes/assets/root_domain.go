package assets

import (
	"recon/internal/api/handlers/assets/root_domain"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerRootDomainRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/root_domain",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     root_domain.GetRootDomainData,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
