package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"recon/internal/services/apikey"
)

var mcpApiKeyService apikey.Service

func init() {
	mcpApiKeyService = apikey.NewService()
}

func MCPAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authenticateGateway(c) {
			return
		}
		rawKey := extractAPIKey(c)
		if rawKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "API key is required. Use X-API-Key header or Authorization: Bearer <api_key>",
			})
			return
		}

		record, err := mcpApiKeyService.Validate(c.Request.Context(), rawKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid API key",
			})
			return
		}

		c.Set("apiKeyID", record.ID.Hex())
		c.Set("apiKeyName", record.Name)
		c.Next()
	}
}

func extractAPIKey(c *gin.Context) string {
	if key := strings.TrimSpace(c.GetHeader("X-API-Key")); key != "" {
		return key
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	token := strings.TrimSpace(parts[1])
	if apikey.IsAPIKeyFormat(token) {
		return token
	}
	return ""
}
