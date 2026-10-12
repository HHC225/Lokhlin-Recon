package plugin

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"recon/internal/models"
)

func TestPluginMutationsUseValidatedGatewayIdentity(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("PLUGINKEY", []byte("standalone-test-key"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		call func(*gin.Context, string) error
		want string
	}{
		{"save", func(c *gin.Context, key string) error {
			return (&service{}).Save(c, &models.PluginSaveRequest{ID: "invalid", Key: key})
		}, "invalid id format"},
		{"archive", func(c *gin.Context, key string) error { return (&service{}).Import(c, "missing.zip", key) }, "Failed to read ZIP file"},
		{"data", func(c *gin.Context, key string) error {
			return (&service{}).ImportByData(c, &models.PluginImportByDataRequest{JSON: "invalid", Key: key})
		}, "Failed to parse JSON string"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, gateway := range []bool{false, true} {
				ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
				ctx.Request = httptest.NewRequest("POST", "/plugin", nil)
				ctx.Request.Header.Set("X-Recon-Authorization", "Bearer unsigned-browser-value")
				ctx.Set("role", "super_admin")
				ctx.Set("userID", "verified-console-user")
				if gateway {
					ctx.Set("recon.gatewayAuthenticated", true)
				}
				err := tc.call(ctx, "")
				want := "invalid plugin key"
				if gateway {
					want = tc.want
				}
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("validated gateway=%v: got %v, want %s", gateway, err, want)
				}
				if !gateway {
					err = tc.call(ctx, "standalone-test-key")
					if err == nil || !strings.Contains(err.Error(), tc.want) {
						t.Fatalf("standalone key rejected: %v", err)
					}
				}
			}
		})
	}
}
