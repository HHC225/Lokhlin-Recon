package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSupportedResponseLanguageFromAcceptLanguage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct{ header, want string }{
		{"", "ko-KR"},
		{"en-US,en;q=0.8", "en-US"},
		{"ko-KR,ko;q=0.9,en;q=0.8", "ko-KR"},
		{"zh-CN", "ko-KR"},
		{"invalid;header", "ko-KR"},
	} {
		router := gin.New()
		router.GET("/", I18nMiddleware(), func(c *gin.Context) { c.String(http.StatusOK, c.GetString("language")) })
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if tc.header != "" {
			req.Header.Set("Accept-Language", tc.header)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Header().Get("Content-Language") != tc.want || w.Body.String() != tc.want {
			t.Fatalf("Accept-Language %q: got header=%q body=%q, want %q", tc.header, w.Header().Get("Content-Language"), w.Body.String(), tc.want)
		}
	}
}
