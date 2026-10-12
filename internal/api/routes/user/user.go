package user

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/handlers/user"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterUserRoutes(api *gin.RouterGroup) {

	userRoutes := models.RouteGroup{
		Path: "/user",
		Routes: []models.Route{

			{
				Method:  "POST",
				Path:    "/login",
				Handler: user.Login,
			},

			{
				Method:      "POST",
				Path:        "/changePassword",
				Handler:     user.ChangePassword,
				Middlewares: common.WithAuth(),
			},
			//{
			//	Method:  "POST",
			//	Path:    "/register",
			//	Handler: user.Register,
			//},

			//{
			//	Method:      "GET",
			//	Path:        "",
			//	Handler:     user.List,
			//	Middlewares: common.WithAuthAndPermission(constants.PermissionUserRead),
			//},
			//{
			//	Method:      "GET",
			//	Path:        "/:id",
			//	Handler:     user.Get,
			//	Middlewares: common.WithAuthAndPermission(constants.PermissionUserRead),
			//},

			//{
			//	Method:      "PUT",
			//	Path:        "/:id",
			//	Handler:     user.Update,
			//	Middlewares: common.WithAuthAndPermission(constants.PermissionUserUpdate),
			//},

			//{
			//	Method:      "DELETE",
			//	Path:        "/:id",
			//	Handler:     user.Delete,
			//	Middlewares: common.WithAuthAndPermission(constants.PermissionUserDelete),
			//},
		},
	}

	common.RegisterRouteGroup(api, userRoutes)
}
