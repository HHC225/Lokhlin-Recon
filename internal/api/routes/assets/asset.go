package assets

import (
	"recon/internal/api/handlers/assets/asset"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerAssetRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/asset",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     asset.GetAssets,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/detail",
				Handler:     asset.GetAssetByID,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/changelog",
				Handler:     asset.GetChangeLog,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/card",
				Handler:     asset.GetAssetCardData,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/hosts",
				Handler:     asset.GetHostsByRootDomain,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/by-host",
				Handler:     asset.GetAssetsByHost,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
