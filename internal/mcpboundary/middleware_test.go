package mcpboundary

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	mcpSDK "github.com/modelcontextprotocol/go-sdk/mcp"
)

func internalMCPEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("MCP_INTERNAL_NOAUTH", "1")
	t.Setenv("MCP_INTERNAL_ALLOWED_IPS", "10.252.87.2/32")
	t.Setenv("MCP_INTERNAL_ALLOWED_HOSTS", "recon-mcp:8082")
}

func internalMCPRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	server := mcpSDK.NewServer(&mcpSDK.Implementation{Name: "boundary-fixture", Version: "1"}, nil)
	mcpSDK.AddTool(server, &mcpSDK.Tool{Name: "boundary_echo", Description: "Read-only fixture"},
		func(context.Context, *mcpSDK.CallToolRequest, struct{}) (*mcpSDK.CallToolResult, any, error) {
			return &mcpSDK.CallToolResult{Content: []mcpSDK.Content{&mcpSDK.TextContent{Text: "boundary-ok"}}}, nil, nil
		})
	handler := mcpSDK.NewStreamableHTTPHandler(func(*http.Request) *mcpSDK.Server { return server }, &mcpSDK.StreamableHTTPOptions{Stateless: true})
	group := router.Group("/mcp")
	group.Use(RouteMiddleware(func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) }))
	group.Any("", gin.WrapH(handler))
	return router
}

func internalMCPRequest(router http.Handler, remote, host string, headers http.Header, method string, id int) *httptest.ResponseRecorder {
	params := map[string]any{}
	if method == "initialize" {
		params = map[string]any{"protocolVersion": "2025-06-18", "capabilities": map[string]any{}, "clientInfo": map[string]string{"name": "console-fixture", "version": "1"}}
	} else if method == "tools/call" {
		params = map[string]any{"name": "boundary_echo", "arguments": map[string]any{}}
	}
	message := map[string]any{"jsonrpc": "2.0", "method": method, "params": params}
	if id != 0 {
		message["id"] = id
	}
	body, _ := json.Marshal(message)
	r := httptest.NewRequest(http.MethodPost, "http://recon-mcp:8082/mcp", strings.NewReader(string(body)))
	r.RemoteAddr, r.Host = remote, host
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json, text/event-stream")
	r.Header.Set("MCP-Protocol-Version", "2025-06-18")
	for key, values := range headers {
		r.Header[key] = append([]string(nil), values...)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

func TestInternalMCPAllowsOnlyDirectConsolePeer(t *testing.T) {
	internalMCPEnvironment(t)
	router := internalMCPRouter()
	for _, tc := range []struct {
		name, remote, host string
		headers            http.Header
		want               int
	}{
		{"console", "10.252.87.2:43000", "recon-mcp:8082", nil, 200},
		{"worker", "10.252.87.4:43000", "recon-mcp:8082", nil, 403},
		{"worker forged forwarding", "10.252.87.4:43000", "recon-mcp:8082", http.Header{"X-Forwarded-For": {"10.252.87.2"}, "X-Real-Ip": {"10.252.87.2"}, "Forwarded": {"for=10.252.87.2"}}, 403},
		{"external", "192.0.2.10:43000", "recon-mcp:8082", nil, 403},
		{"loopback", "127.0.0.1:43000", "recon-mcp:8082", nil, 403},
		{"missing socket peer", "", "recon-mcp:8082", nil, 403},
		{"peer without port", "10.252.87.2", "recon-mcp:8082", nil, 403},
		{"wrong host", "10.252.87.2:43000", "recon:8082", nil, 403},
		{"wrong host port", "10.252.87.2:43000", "recon-mcp:8083", nil, 403},
		{"public host", "10.252.87.2:43000", "public.example", nil, 403},
		{"same origin", "10.252.87.2:43000", "recon-mcp:8082", http.Header{"Origin": {"http://recon-mcp:8082"}}, 200},
		{"foreign origin", "10.252.87.2:43000", "recon-mcp:8082", http.Header{"Origin": {"https://outside.example"}}, 403},
		{"origin userinfo", "10.252.87.2:43000", "recon-mcp:8082", http.Header{"Origin": {"http://user@recon-mcp:8082"}}, 403},
		{"origin path", "10.252.87.2:43000", "recon-mcp:8082", http.Header{"Origin": {"http://recon-mcp:8082/path"}}, 403},
		{"origin query", "10.252.87.2:43000", "recon-mcp:8082", http.Header{"Origin": {"http://recon-mcp:8082?"}}, 403},
		{"origin empty fragment", "10.252.87.2:43000", "recon-mcp:8082", http.Header{"Origin": {"http://recon-mcp:8082#"}}, 403},
		{"origin different scheme", "10.252.87.2:43000", "recon-mcp:8082", http.Header{"Origin": {"https://recon-mcp:8082"}}, 403},
		{"null origin", "10.252.87.2:43000", "recon-mcp:8082", http.Header{"Origin": {"null"}}, 403},
		{"multiple origins", "10.252.87.2:43000", "recon-mcp:8082", http.Header{"Origin": {"http://recon-mcp:8082", "https://outside.example"}}, 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := internalMCPRequest(router, tc.remote, tc.host, tc.headers, "initialize", 1)
			if w.Code != tc.want {
				t.Fatalf("MCP boundary status=%d want=%d body=%s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestInternalMCPFlagsFailClosedAndDefaultRequiresAuthentication(t *testing.T) {
	for _, tc := range []struct {
		name, flag, ips, hosts string
		want                   int
	}{
		{"default", "", "", "", 401},
		{"off", "0", "10.252.87.2/32", "recon-mcp:8082", 401},
		{"missing peers", "1", "", "recon-mcp:8082", 503},
		{"missing hosts", "1", "10.252.87.2/32", "", 503},
		{"invalid peers", "1", "invalid", "recon-mcp:8082", 503},
		{"unrestricted peers", "1", "0.0.0.0/0", "recon-mcp:8082", 503},
		{"network peers", "1", "10.252.87.0/29", "recon-mcp:8082", 503},
		{"wildcard host", "1", "10.252.87.2/32", "*", 503},
		{"URL instead of host", "1", "10.252.87.2/32", "http://recon-mcp:8082", 503},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MCP_INTERNAL_NOAUTH", tc.flag)
			t.Setenv("MCP_INTERNAL_ALLOWED_IPS", tc.ips)
			t.Setenv("MCP_INTERNAL_ALLOWED_HOSTS", tc.hosts)
			w := internalMCPRequest(internalMCPRouter(), "10.252.87.2:43000", "recon-mcp:8082", nil, "initialize", 1)
			if w.Code != tc.want {
				t.Fatalf("MCP configuration status=%d want=%d", w.Code, tc.want)
			}
		})
	}
}

func TestInternalMCPNoHeadersHandshakeListAndCall(t *testing.T) {
	internalMCPEnvironment(t)
	router := internalMCPRouter()
	for _, tc := range []struct {
		method string
		id     int
		want   int
	}{
		{"initialize", 1, 200}, {"notifications/initialized", 0, 202},
		{"tools/list", 2, 200}, {"tools/call", 3, 200},
	} {
		w := internalMCPRequest(router, "10.252.87.2:43000", "recon-mcp:8082", nil, tc.method, tc.id)
		if w.Code != tc.want {
			t.Fatalf("%s status=%d body=%s", tc.method, w.Code, w.Body.String())
		}
		if tc.method == "tools/list" && !strings.Contains(w.Body.String(), "boundary_echo") {
			t.Fatal("MCP tools were not listed")
		}
		if tc.method == "tools/call" && !strings.Contains(w.Body.String(), "boundary-ok") {
			t.Fatalf("tools/call did not reach the SDK: %s", w.Body.String())
		}
	}
}
