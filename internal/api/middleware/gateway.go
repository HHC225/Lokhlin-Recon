package middleware

import (
	"net/http"
	"os"
	"recon/internal/constants"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const gatewayAuthorizationHeader = "X-Recon-Authorization"

type gatewayClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	Role     string `json:"role"`
}

// authenticateGateway handles only dedicated, signed console assertions.
// A supplied assertion must succeed on its own; other credentials cannot rescue it.
func authenticateGateway(c *gin.Context) bool {
	headers, present := c.Request.Header[http.CanonicalHeaderKey(gatewayAuthorizationHeader)]
	if !present {
		return false
	}
	deny := func() bool {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code": http.StatusUnauthorized, "message": "Invalid gateway authorization",
		})
		return true
	}
	secret := os.Getenv("RECON_GATEWAY_SECRET")
	if secret == "" || len(headers) != 1 {
		return deny()
	}
	parts := strings.Split(headers[0], " ")
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return deny()
	}
	claims := &gatewayClaims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("console-gateway"),
		jwt.WithAudience("recon"), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid || strings.TrimSpace(claims.Subject) == "" ||
		strings.TrimSpace(claims.Username) == "" || claims.Role != "super_admin" ||
		len(claims.Audience) != 1 || claims.Audience[0] != "recon" ||
		claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return deny()
	}
	validity := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
	if validity <= 0 || validity > 60*time.Second {
		return deny()
	}
	c.Set("userID", claims.Subject)
	c.Set("username", claims.Username)
	c.Set("role", claims.Role)
	c.Set(constants.GatewayAuthenticatedContextKey, true)
	c.Next()
	return true
}
