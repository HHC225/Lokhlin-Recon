// export-------------------------------------
// @file      : export.go
// @time      : 2025/5/25 15:12
// -------------------------------------------

package export

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/handlers/export"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterExportRoutes(api *gin.RouterGroup) {

	nodeRoutes := models.RouteGroup{
		Path: "/export",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     export.Export,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/fields",
				Handler:     export.GetExportFields,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "GET",
				Path:        "/records",
				Handler:     export.GetExportRecords,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/delete",
				Handler:     export.DeleteExport,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "GET",
				Path:        "/download",
				Handler:     export.DownloadExport,
				Middlewares: common.WithAuth(),
			},
		},
	}

	common.RegisterRouteGroup(api, nodeRoutes)
}
