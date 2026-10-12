// fingerprint-------------------------------------
// @file      : fingerprint.go
// @time      : 2025/10/28 20:38
// -------------------------------------------

package fingerprint

import (
	"github.com/gin-gonic/gin"
	handler "recon/internal/api/handlers/fingerprint"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterFingerprintRoutes(api *gin.RouterGroup) {

	fingerprintRoutes := models.RouteGroup{
		Path: "/fingerprint",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "/data",
				Handler:     handler.Data,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/update",
				Handler:     handler.Update,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/add",
				Handler:     handler.Add,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/delete",
				Handler:     handler.Delete,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "GET",
				Path:        "/version",
				Handler:     handler.GetVersion,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/batch-add",
				Handler:     handler.BatchAdd,
				Middlewares: common.WithAuth(),
			},
		},
	}

	common.RegisterRouteGroup(api, fingerprintRoutes)
}
