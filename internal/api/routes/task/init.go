// task-------------------------------------
// @file      : task.go
// @time      : 2025/5/10 15:33
// -------------------------------------------

package task

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterTaskRoutes(api *gin.RouterGroup) {

	taskRoutes := models.RouteGroup{
		Path: "/task",
		Groups: []models.RouteGroup{
			registerTaskRoutes(),
			registerTemplateRoutes(),
			registerSchedulerTaskRoutes(),
		},
	}

	common.RegisterRouteGroup(api, taskRoutes)
}
