package configuration

import (
	"github.com/gin-gonic/gin"
	confHandler "recon/internal/api/handlers/configuration"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterConfigurationRoutes(api *gin.RouterGroup) {
	group := models.RouteGroup{
		Path: "/configuration",
		Groups: []models.RouteGroup{
			{
				Path: "/subfinder",
				Routes: []models.Route{
					{Method: "GET", Path: "/data", Handler: confHandler.GetSubfinderData, Middlewares: common.WithAuth()},
					{Method: "POST", Path: "/save", Handler: confHandler.SaveSubfinderData, Middlewares: common.WithAuth()},
				},
			},
			{
				Path: "/rad",
				Routes: []models.Route{
					{Method: "GET", Path: "/data", Handler: confHandler.GetRadData, Middlewares: common.WithAuth()},
					{Method: "POST", Path: "/save", Handler: confHandler.SaveRadData, Middlewares: common.WithAuth()},
				},
			},
			{
				Path: "/system",
				Routes: []models.Route{
					{Method: "GET", Path: "/data", Handler: confHandler.GetSystemData, Middlewares: common.WithAuth()},
					{Method: "POST", Path: "/save", Handler: confHandler.SaveSystemData, Middlewares: common.WithAuth()},
				},
			},
			{
				Path: "/deduplication",
				Routes: []models.Route{
					{Method: "GET", Path: "/config", Handler: confHandler.GetDeduplicationConfig, Middlewares: common.WithAuth()},
					{Method: "POST", Path: "/save", Handler: confHandler.SaveDeduplicationConfig, Middlewares: common.WithAuth()},
				},
			},
			{
				Path: "/notification",
				Routes: []models.Route{
					{Method: "GET", Path: "/data", Handler: confHandler.GetNotificationData, Middlewares: common.WithAuth()},
					{Method: "POST", Path: "/add", Handler: confHandler.AddNotificationData, Middlewares: common.WithAuth()},
					{Method: "POST", Path: "/update", Handler: confHandler.UpdateNotificationData, Middlewares: common.WithAuth()},
					{Method: "POST", Path: "/delete", Handler: confHandler.DeleteNotification, Middlewares: common.WithAuth()},
				},
				Groups: []models.RouteGroup{
					{
						Path: "/config",
						Routes: []models.Route{
							{Method: "GET", Path: "/data", Handler: confHandler.GetNotificationConfigData, Middlewares: common.WithAuth()},
							{Method: "POST", Path: "/update", Handler: confHandler.UpdateNotificationConfigData, Middlewares: common.WithAuth()},
						},
					},
				},
			},
		},
	}
	common.RegisterRouteGroup(api, group)
}
