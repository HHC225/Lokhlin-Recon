// Package node -----------------------------
// @file      : node.go
// @time      : 2025/5/3 10:57
// -------------------------------------------
package node

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/handlers/node"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterNodeRoutes(api *gin.RouterGroup) {

	nodeRoutes := models.RouteGroup{
		Path: "/node",
		Routes: []models.Route{

			{
				Method:      "GET",
				Path:        "",
				Handler:     node.GetNode,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "GET",
				Path:        "/online",
				Handler:     node.GetNodeOnline,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/config/update",
				Handler:     node.ConfigUpdate,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/delete",
				Handler:     node.Delete,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/log",
				Handler:     node.GetLogs,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/plugin",
				Handler:     node.GetNodePlugin,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/restart",
				Handler:     node.RestartNode,
				Middlewares: common.WithAuth(),
			},
		},
	}

	common.RegisterRouteGroup(api, nodeRoutes)
}
