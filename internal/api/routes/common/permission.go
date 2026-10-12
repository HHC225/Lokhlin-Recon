// common-------------------------------------
// @file      : permission.go
// @time      : 2025/5/8 21:11
// -------------------------------------------

package common

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/middleware"
)

func WithAuth() []gin.HandlerFunc {
	return []gin.HandlerFunc{middleware.AuthMiddleware()}
}

func WithPermission(permission string) []gin.HandlerFunc {
	return []gin.HandlerFunc{middleware.PermissionMiddleware(permission)}
}

func WithAuthAndPermission(permission string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthMiddleware(),
		middleware.PermissionMiddleware(permission),
	}
}
