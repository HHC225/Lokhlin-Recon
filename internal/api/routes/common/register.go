// common-------------------------------------
// @file      : register.go
// @time      : 2025/5/8 21:10
// -------------------------------------------

package common

import (
	"github.com/gin-gonic/gin"
	"recon/internal/models"
)

func RegisterRouteGroup(group *gin.RouterGroup, routeGroup models.RouteGroup) {

	subGroup := group.Group(routeGroup.Path)

	for _, handlerFunc := range routeGroup.Middlewares {
		subGroup.Use(handlerFunc)
	}

	for _, route := range routeGroup.Routes {

		handlers := make([]gin.HandlerFunc, 0)

		handlers = append(handlers, route.Middlewares...)

		handlers = append(handlers, route.Handler)

		switch route.Method {
		case "GET":
			subGroup.GET(route.Path, handlers...)
		case "POST":
			subGroup.POST(route.Path, handlers...)
		case "PUT":
			subGroup.PUT(route.Path, handlers...)
		case "DELETE":
			subGroup.DELETE(route.Path, handlers...)
		case "PATCH":
			subGroup.PATCH(route.Path, handlers...)
		}
	}

	for _, childGroup := range routeGroup.Groups {
		RegisterRouteGroup(subGroup, childGroup)
	}
}
