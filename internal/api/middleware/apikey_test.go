package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestMCPGatewayAssertionAccepted(t *testing.T) {
	t.Setenv("RECON_GATEWAY_SECRET", assertionTestSecret)
	w := authRequest(MCPAPIKeyMiddleware(), signedAssertion(t, jwt.SigningMethodHS256, validAssertionClaims(), assertionTestSecret), "")
	if w.Code != http.StatusOK {
		t.Fatalf("MCP gateway assertion rejected: status=%d body=%s", w.Code, w.Body.String())
	}
	if w.Body.String() != `{"role":"super_admin","userID":"console-user-42","username":"admin@example.test"}` {
		t.Fatalf("MCP identity was not propagated: %s", w.Body.String())
	}
}

func TestMCPGatewayInvalidAndDisabledAssertionsRejected(t *testing.T) {
	for _, secret := range []string{"", assertionTestSecret} {
		t.Setenv("RECON_GATEWAY_SECRET", secret)
		w := authRequest(MCPAPIKeyMiddleware(), signedAssertion(t, jwt.SigningMethodHS256, validAssertionClaims(), "wrong-key"), "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("MCP accepted invalid/disabled assertion: %d", w.Code)
		}
	}
}

func TestMCPLegacyAPIKeyAndGatewayFallbackBoundary(t *testing.T) {
	t.Setenv("RECON_GATEWAY_SECRET", assertionTestSecret)
	key, err := mcpApiKeyService.Create(context.Background(), "gateway-auth-verification", "test-fixture")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mcpApiKeyService.Delete(context.Background(), key.ID); err != nil {
			t.Error(err)
		}
	})
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/private", MCPAPIKeyMiddleware(), func(c *gin.Context) {
		if c.GetString("apiKeyID") != key.ID {
			c.Status(http.StatusForbidden)
			return
		}
		c.Status(http.StatusOK)
	})
	for _, tc := range []struct {
		name, keyHeader, bearerHeader, gateway string
		want                                   int
	}{
		{name: "X-API-Key", keyHeader: key.Key, want: http.StatusOK},
		{name: "Authorization API key", bearerHeader: "Bearer " + key.Key, want: http.StatusOK},
		{name: "invalid key", keyHeader: "ssk_missing-fixture", want: http.StatusUnauthorized},
		{name: "invalid gateway cannot rescue valid API key", keyHeader: key.Key, gateway: "Bearer invalid", want: http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/private", nil)
			if tc.keyHeader != "" {
				req.Header.Set("X-API-Key", tc.keyHeader)
			}
			if tc.bearerHeader != "" {
				req.Header.Set("Authorization", tc.bearerHeader)
			}
			if tc.gateway != "" {
				req.Header.Set("X-Recon-Authorization", tc.gateway)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("unexpected authorization status: got %d want %d", w.Code, tc.want)
			}
		})
	}
}
