// project-------------------------------------
// @file      : project.go
// @time      : 2025/5/13 21:22
// -------------------------------------------

package project

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/handlers/project"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterProjectRoutes(api *gin.RouterGroup) {

	projectRoutes := models.RouteGroup{
		Path: "/project",
		Routes: []models.Route{

			{
				Method:      "GET",
				Path:        "/all",
				Handler:     project.GetProjectsByTag,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/data",
				Handler:     project.GetProjectsData,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/content",
				Handler:     project.GetProjectContent,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/add",
				Handler:     project.AddProject,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/delete",
				Handler:     project.DeleteProject,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "/update",
				Handler:     project.UpdatePorject,
				Middlewares: common.WithAuth(),
			},
		},
	}

	common.RegisterRouteGroup(api, projectRoutes)
}
