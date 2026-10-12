package mcp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mcpSDK "github.com/modelcontextprotocol/go-sdk/mcp"
	"recon/internal/api/middleware"
	"recon/internal/mcpboundary"
)

func RegisterRoutes(router *gin.Engine) {
	handler := mcpSDK.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcpSDK.Server { return newServer() },
		&mcpSDK.StreamableHTTPOptions{Stateless: true},
	)

	mcpGroup := router.Group("/mcp")
	mcpGroup.Use(mcpboundary.RouteMiddleware(middleware.MCPAPIKeyMiddleware()))
	mcpGroup.Any("", gin.WrapH(handler))
	mcpGroup.Any("/*filepath", gin.WrapH(handler))
}
