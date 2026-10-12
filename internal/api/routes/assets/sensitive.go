// assets-------------------------------------
// @file      : sensitive.go
// @time      : 2025/5/17 11:33
// -------------------------------------------

package assets

import (
	"recon/internal/api/handlers/assets/sensitive"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerSensitivveRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/sensitive",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     sensitive.GetSensitiveInfo,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/number",
				Handler:     sensitive.GetSensitiveInfoNumber,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/body",
				Handler:     sensitive.GetSensitiveInfoBody,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/names",
				Handler:     sensitive.GetSensitiveInfoName,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/info",
				Handler:     sensitive.GetSensitiveMatchInfo,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
