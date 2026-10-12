package sensitive_rule

import (
	"github.com/gin-gonic/gin"
	handler "recon/internal/api/handlers/sensitiveRule"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterSensitiveRuleRoutes(api *gin.RouterGroup) {
	routes := models.RouteGroup{
		Path: "/sensitive",
		Routes: []models.Route{
			{Method: "POST", Path: "/data", Handler: handler.Data, Middlewares: common.WithAuth()},
			{Method: "POST", Path: "/update", Handler: handler.Update, Middlewares: common.WithAuth()},
			{Method: "POST", Path: "/add", Handler: handler.Add, Middlewares: common.WithAuth()},
			{Method: "POST", Path: "/update/state", Handler: handler.UpdateState, Middlewares: common.WithAuth()},
			{Method: "POST", Path: "/delete", Handler: handler.Delete, Middlewares: common.WithAuth()},
		},
	}
	common.RegisterRouteGroup(api, routes)
}
