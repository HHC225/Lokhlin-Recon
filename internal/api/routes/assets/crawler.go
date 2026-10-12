// assets-------------------------------------
// @file      : crawler.go
// @time      : 2025/5/14 23:33
// -------------------------------------------

package assets

import (
	"recon/internal/api/handlers/assets/crawler"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerCrawlerRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/crawler",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     crawler.GetCrawlers,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
