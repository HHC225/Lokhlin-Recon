package middleware

import (
	"net/http"
	"net/http/httptest"
	"recon/internal/config"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const assertionTestSecret = "test-console-gateway-secret-with-enough-entropy"

func signedAssertion(t *testing.T, method jwt.SigningMethod, claims jwt.MapClaims, secret string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + token
}

func validAssertionClaims() jwt.MapClaims {
	now := time.Now().Unix()
	return jwt.MapClaims{
		"iss": "console-gateway", "aud": "recon", "sub": "console-user-42",
		"username": "admin@example.test", "role": "super_admin", "iat": now, "exp": now + 60,
	}
}

func authRequest(handler gin.HandlerFunc, assertion, authorization string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/private", handler, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"userID": c.GetString("userID"), "username": c.GetString("username"), "role": c.GetString("role")})
	})
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	if assertion != "" {
		req.Header.Set("X-Recon-Authorization", assertion)
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestRESTGatewayAssertionAccepted(t *testing.T) {
	t.Setenv("RECON_GATEWAY_SECRET", assertionTestSecret)
	w := authRequest(AuthMiddleware(), signedAssertion(t, jwt.SigningMethodHS256, validAssertionClaims(), assertionTestSecret), "")
	if w.Code != http.StatusOK {
		t.Fatalf("valid assertion rejected: status=%d body=%s", w.Code, w.Body.String())
	}
	if w.Body.String() != `{"role":"super_admin","userID":"console-user-42","username":"admin@example.test"}` {
		t.Fatalf("console identity was not propagated: %s", w.Body.String())
	}
}

func TestOnlySignedGatewayMarksPluginAuthority(t *testing.T) {
	t.Setenv("RECON_GATEWAY_SECRET", assertionTestSecret)
	gateway := signedAssertion(t, jwt.SigningMethodHS256, validAssertionClaims(), assertionTestSecret)
	standalone := signedAssertion(t, jwt.SigningMethodHS256, jwt.MapClaims{"userID": "local", "username": "admin", "role": "super_admin", "exp": time.Now().Add(time.Minute).Unix()}, config.GlobalConfig.JWT.Secret)
	for _, tc := range []struct {
		name, assertion, authorization string
		want                           bool
	}{
		{"gateway", gateway, "", true},
		{"standalone", "", standalone, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/private", AuthMiddleware(), func(c *gin.Context) { c.JSON(200, c.GetBool("recon.gatewayAuthenticated")) })
			r := httptest.NewRequest("GET", "/private", nil)
			r.Header.Set("X-Recon-Gateway-Authenticated", "true")
			if tc.assertion != "" {
				r.Header.Set("X-Recon-Authorization", tc.assertion)
			}
			if tc.authorization != "" {
				r.Header.Set("Authorization", tc.authorization)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			want := "false"
			if tc.want {
				want = "true"
			}
			if w.Code != 200 || w.Body.String() != want {
				t.Fatalf("got %d %s, want %s", w.Code, w.Body.String(), want)
			}
		})
	}
}

func TestRESTGatewayAssertionRejected(t *testing.T) {
	tests := []struct {
		name             string
		change           func(jwt.MapClaims)
		method           jwt.SigningMethod
		secret           string
		configuredSecret string
	}{
		{name: "disabled", configuredSecret: ""},
		{name: "wrong signature", secret: "another-secret"},
		{name: "wrong algorithm", method: jwt.SigningMethodHS384},
		{name: "wrong issuer", change: func(c jwt.MapClaims) { c["iss"] = "other" }},
		{name: "missing issuer", change: func(c jwt.MapClaims) { delete(c, "iss") }},
		{name: "wrong audience", change: func(c jwt.MapClaims) { c["aud"] = "other" }},
		{name: "missing audience", change: func(c jwt.MapClaims) { delete(c, "aud") }},
		{name: "multiple audiences", change: func(c jwt.MapClaims) { c["aud"] = []string{"recon", "other"} }},
		{name: "missing subject", change: func(c jwt.MapClaims) { delete(c, "sub") }},
		{name: "blank subject", change: func(c jwt.MapClaims) { c["sub"] = " " }},
		{name: "expired", change: func(c jwt.MapClaims) { c["iat"] = time.Now().Unix() - 120; c["exp"] = time.Now().Unix() - 60 }},
		{name: "missing expiry", change: func(c jwt.MapClaims) { delete(c, "exp") }},
		{name: "wrong role", change: func(c jwt.MapClaims) { c["role"] = "admin" }},
		{name: "missing role", change: func(c jwt.MapClaims) { delete(c, "role") }},
		{name: "missing username", change: func(c jwt.MapClaims) { delete(c, "username") }},
		{name: "future issued at", change: func(c jwt.MapClaims) { c["iat"] = time.Now().Unix() + 60; c["exp"] = time.Now().Unix() + 120 }},
		{name: "missing issued at", change: func(c jwt.MapClaims) { delete(c, "iat") }},
		{name: "overlong validity", change: func(c jwt.MapClaims) { c["exp"] = time.Now().Unix() + 3600 }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configured := assertionTestSecret
			if tc.name == "disabled" {
				configured = tc.configuredSecret
			}
			t.Setenv("RECON_GATEWAY_SECRET", configured)
			claims := validAssertionClaims()
			if tc.change != nil {
				tc.change(claims)
			}
			method := tc.method
			if method == nil {
				method = jwt.SigningMethodHS256
			}
			secret := tc.secret
			if secret == "" {
				secret = assertionTestSecret
			}
			w := authRequest(AuthMiddleware(), signedAssertion(t, method, claims, secret), "")
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("invalid assertion accepted: status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func legacyJWT(t *testing.T) string {
	t.Helper()
	return signedAssertion(t, jwt.SigningMethodHS512, jwt.MapClaims{
		"userID": "standalone-user", "username": "standalone", "role": "admin",
		"exp": time.Now().Unix() + 3600,
	}, config.GlobalConfig.JWT.Secret)
}

func TestRESTLegacyJWTStillAccepted(t *testing.T) {
	t.Setenv("RECON_GATEWAY_SECRET", assertionTestSecret)
	w := authRequest(AuthMiddleware(), "", legacyJWT(t))
	if w.Code != http.StatusOK {
		t.Fatalf("standalone JWT rejected: status=%d body=%s", w.Code, w.Body.String())
	}
	if w.Body.String() != `{"role":"admin","userID":"standalone-user","username":"standalone"}` {
		t.Fatalf("standalone identity changed: %s", w.Body.String())
	}
}

func TestRESTInvalidGatewayCannotFallbackToLegacyJWT(t *testing.T) {
	t.Setenv("RECON_GATEWAY_SECRET", assertionTestSecret)
	w := authRequest(AuthMiddleware(), "Bearer invalid", legacyJWT(t))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("invalid gateway assertion fell back to JWT: %d", w.Code)
	}
}

func TestRESTEmptyOrDuplicateGatewayHeadersRejected(t *testing.T) {
	t.Setenv("RECON_GATEWAY_SECRET", assertionTestSecret)
	for _, values := range [][]string{{""}, {signedAssertion(t, jwt.SigningMethodHS256, validAssertionClaims(), assertionTestSecret), signedAssertion(t, jwt.SigningMethodHS256, validAssertionClaims(), assertionTestSecret)}} {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/private", AuthMiddleware(), func(c *gin.Context) { c.Status(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/private", nil)
		req.Header.Set("Authorization", legacyJWT(t))
		for _, value := range values {
			req.Header.Add("X-Recon-Authorization", value)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("ambiguous gateway headers accepted: status=%d", w.Code)
		}
	}
}

func TestRESTUntrustedIdentityHeadersRejected(t *testing.T) {
	t.Setenv("RECON_GATEWAY_SECRET", assertionTestSecret)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/private", AuthMiddleware(), func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("X-User-ID", "console-user-42")
	req.Header.Set("X-User-Role", "super_admin")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unsigned identity headers accepted: %d", w.Code)
	}
}
