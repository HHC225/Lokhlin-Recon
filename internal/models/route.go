// models-------------------------------------
// @file      : route.go
// @time      : 2025/5/8 21:09
// -------------------------------------------

package models

import "github.com/gin-gonic/gin"

type RouteGroup struct {
	Path        string
	Middlewares []gin.HandlerFunc
	Routes      []Route
	Groups      []RouteGroup
}

type Route struct {
	Method      string
	Path        string
	Handler     gin.HandlerFunc
	Middlewares []gin.HandlerFunc
}
