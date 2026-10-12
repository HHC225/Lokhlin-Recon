package task

import (
	taskHandler "recon/internal/api/handlers/task"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerTemplateRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/template",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     taskHandler.TemplateList,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/detail",
				Handler:     taskHandler.TemplateDetail,
				Middlewares: common.WithAuth(),
			},
			{
				Method:  "POST",
				Path:    "/delete",
				Handler: taskHandler.Delete,
			},
			{
				Method:      "POST",
				Path:        "/save",
				Handler:     taskHandler.Save,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
