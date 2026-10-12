// task-------------------------------------
// @file      : task.go
// @time      : 2025/6/14 19:40
// -------------------------------------------

package task

import (
	taskHandler "recon/internal/api/handlers/task"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func registerTaskRoutes() models.RouteGroup {
	return models.RouteGroup{
		Path: "/",
		Routes: []models.Route{
			{
				Method:      "POST",
				Path:        "",
				Handler:     taskHandler.List,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "detail",
				Handler:     taskHandler.TaskDetail,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "add",
				Handler:     taskHandler.AddTask,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "delete",
				Handler:     taskHandler.DeleteTask,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "retest",
				Handler:     taskHandler.RetestTask,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "stop",
				Handler:     taskHandler.StopTask,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "start",
				Handler:     taskHandler.StartTask,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "POST",
				Path:        "progress/info",
				Handler:     taskHandler.ProgressInfo,
				Middlewares: common.WithAuth(),
			},
			{
				Method:      "GET",
				Path:        "names",
				Handler:     taskHandler.GetAllTaskNames,
				Middlewares: common.WithAuth(),
			},
		},
	}
}
