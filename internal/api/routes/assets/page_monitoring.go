package assets

import (
	"recon/internal/api/handlers/assets/page_monitoring"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerPageMonitoringRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/page-monitoring",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     page_monitoring.GetResult,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/diff",
				Handler:     page_monitoring.GetDiff,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
