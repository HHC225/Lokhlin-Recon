package middleware

import (
	"recon/internal/api/response"
	"recon/internal/logger"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"recon/internal/config"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authenticateGateway(c) {
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required", nil)
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization header format", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.GlobalConfig.JWT.Secret), nil
		})

		if err != nil {
			logger.Error("Failed to parse token", zap.Error(err))
			response.Unauthorized(c, "Failed to parse token", nil)
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

			c.Set("userID", claims["userID"])
			c.Set("username", claims["username"])
			c.Set("role", claims["role"])
			c.Next()
		} else {
			response.Unauthorized(c, "Invalid token", nil)
			c.Abort()
			return
		}
	}
}
