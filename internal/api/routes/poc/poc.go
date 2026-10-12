package poc

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/handlers/poc"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterPocRoutes(api *gin.RouterGroup) {

	pocRoutes := models.RouteGroup{
		Path: "/poc",
		Routes: []models.Route{

			{
				Method:      "POST",
				Path:        "",
				Handler:     poc.GetPocList,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/detail",
				Handler:     poc.GetPocDetail,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/update",
				Handler:     poc.UpdatePoc,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/add",
				Handler:     poc.AddPoc,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/delete",
				Handler:     poc.DeletePoc,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "GET",
				Path:        "/data/all",
				Handler:     poc.GetAllPocData,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/data/import",
				Handler:     poc.ImportPoc,
				Middlewares: common.WithAuth(),
			},
		},
	}

	common.RegisterRouteGroup(api, pocRoutes)
}
