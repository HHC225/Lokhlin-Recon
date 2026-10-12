// Package system -----------------------------
// @file      : system.go
// @time      : 2025/5/2 23:01
// -------------------------------------------
package system

import (
	"encoding/json"
	"net/http"
	"os"
	"recon/internal/constants"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valyala/fasthttp"
	"recon/internal/api/response"
	"recon/internal/database/redis"
)

type VersionResponse struct {
	Code int `json:"code"`
	Data struct {
		List []VersionInfo `json:"list"`
	} `json:"data"`
}

type VersionInfo struct {
	Name     string `json:"name"`
	CVersion string `json:"cversion"`
	LVersion string `json:"lversion"`
	Msg      string `json:"msg"`
}

type RemoteVersion struct {
	Server    string `json:"server"`
	ServerMsg string `json:"server_msg"`
	Scan      string `json:"scan"`
	ScanMsg   string `json:"scan_msg"`
}

// GetSystemVersion @Summary      Get server and node versions
// @Description  Get current versions of the Recon server and all scanner nodes
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse{data=VersionListResponse}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      401  {object}  response.UnauthorizedResponse
// @Failure      404  {object}  response.NotFoundResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /system/version [get]
func GetSystemVersion(c *gin.Context) {
	var (
		serverLVersion string
		serverMsg      string
		scanLVersion   string
		scanMsg        string
	)

	serverLVersion = constants.Version
	if versionURL := strings.TrimSpace(os.Getenv("RECON_VERSION_URL")); versionURL != "" {
		client := &fasthttp.Client{ReadTimeout: 10 * time.Second}
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)
		req.SetRequestURI(versionURL)
		req.Header.SetMethod("GET")
		if err := client.Do(req, resp); err == nil && resp.StatusCode() == http.StatusOK {
			var version RemoteVersion
			if err := json.Unmarshal(resp.Body(), &version); err == nil {
				serverLVersion, serverMsg = version.Server, version.ServerMsg
				scanLVersion, scanMsg = version.Scan, version.ScanMsg
			}
		}
	}

	resultList := []VersionInfo{
		{
			Name:     constants.DisplayName() + "-Server",
			CVersion: constants.Version,
			LVersion: serverLVersion,
			Msg:      serverMsg,
		},
	}

	ctx := c.Request.Context()
	keys, err := redis.Client.Keys(ctx, "node:*").Result()
	if err != nil {
		response.InternalServerError(c, "api.system.version.redis_error", err)
		return
	}
	for _, key := range keys {
		name := key[5:]
		hashData, err := redis.Client.HGetAll(ctx, key).Result()
		if err != nil {
			response.InternalServerError(c, "api.system.version.redis_hash_error", err)
			return
		}
		latestVersion := scanLVersion
		if latestVersion == "" {
			latestVersion = hashData["version"]
		}
		resultList = append(resultList, VersionInfo{
			Name:     name,
			CVersion: hashData["version"],
			LVersion: latestVersion,
			Msg:      scanMsg,
		})
	}

	response.Success(c, VersionListResponse{
		List: resultList,
	}, "api.system.version.success")
}

// @Description Version information list
// @Date 2025-05-02
type VersionListResponse struct {
	List []VersionInfo `json:"list"`
}
