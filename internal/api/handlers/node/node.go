// Package node -----------------------------
// @file      : node.go
// @time      : 2025/4/30 15:36
// -------------------------------------------
package node

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/models"
	"recon/internal/services/node"
)

var nodeService node.Service

// GetNode @Summary      List node data
// @Description  List data for all online nodes
// @Tags         node
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  response.SuccessResponse{data=object{list=[]models.NodeData}}
// @Failure      401  {object}  response.UnauthorizedResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /node [get]
func GetNode(c *gin.Context) {
	result, err := nodeService.GetNodeData(c.Request.Context(), false)
	if err != nil {
		response.InternalServerError(c, "api.node.get_data.failed", err)
		return
	}

	response.Success(c, response.DataResponse{List: result}, "")
}

// GetNodeOnline @Summary      List online nodes
// @Description  List data for all online nodes
// @Tags         node
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  response.SuccessResponse{data=object{list=[]string}}
// @Failure      401  {object}  response.UnauthorizedResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /node/online [get]
func GetNodeOnline(c *gin.Context) {
	result, err := nodeService.GetNodeData(c.Request.Context(), true)
	if err != nil {
		response.InternalServerError(c, "api.node.get_data.failed", err)
		return
	}

	names := make([]string, 0, len(result))
	for _, item := range result {
		names = append(names, item.Name)
	}

	response.Success(c, response.DataResponse{List: names}, "")
}

type NodeConfigUpdateRequest struct {
	Name          string `json:"name" binding:"required"`
	OldName       string `json:"oldName" binding:"omitempty"`
	ModulesConfig string `json:"ModulesConfig" binding:"required"`
	State         bool   `json:"state" binding:"required"`
}

// @Summary Update node configuration
// @Tags node
// @Accept json
// @Produce json
// @Param data body NodeConfigUpdateRequest true "Node configuration"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/node/config/update [post]
func ConfigUpdate(c *gin.Context) {
	var req NodeConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	old := req.OldName
	if strings.TrimSpace(old) == "" {
		old = req.Name
	}
	msg := models.Message{
		Name:    old,
		Type:    "nodeConfig",
		Content: req.Name + "[*]" + fmt.Sprintf("%v", req.State) + "[*]" + req.ModulesConfig,
	}
	if err := nodeService.RefreshConfig(c, msg); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

type NodeDeleteRequest struct {
	Names []string `json:"names" binding:"required"`
}

// @Summary Delete nodes
// @Tags node
// @Accept json
// @Produce json
// @Param data body NodeDeleteRequest true "Node names"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/node/delete [post]
func Delete(c *gin.Context) {
	var req NodeDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if len(req.Names) == 0 {
		response.BadRequest(c, "api.bad_request", nil)
		return
	}
	ctx := context.Background()
	if err := nodeService.DeleteNodes(ctx, req.Names); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

type NodeLogRequest struct {
	Name string `json:"name" binding:"required"`
}

// @Summary Get node logs
// @Tags node
// @Accept json
// @Produce json
// @Param data body NodeLogRequest true "Node name"
// @Success 200 {object} response.Response{data=object{logs=string}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/node/log [post]
func GetLogs(c *gin.Context) {
	var req NodeLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	ctx := context.Background()
	logs, err := nodeService.GetNodeLogs(ctx, req.Name)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"logs": logs}, "")
}

// @Summary Get node plugin information
// @Tags node
// @Accept json
// @Produce json
// @Param data body NodeLogRequest true "Node name"
// @Success 200 {object} response.Response{data=object{list=[]models.NodePluginInfo}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/node/plugin [post]
func GetNodePlugin(c *gin.Context) {
	var req NodeLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	ctx := context.Background()
	plugins, err := nodeService.GetNodePlugin(ctx, req.Name)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, response.DataResponse{List: plugins}, "")
}

// @Summary Restart nodes
// @Tags node
// @Accept json
// @Produce json
// @Param data body NodeLogRequest true "Node name"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/node/restart [post]
func RestartNode(c *gin.Context) {
	var req NodeLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	ctx := context.Background()
	if err := nodeService.RestartNode(ctx, req.Name); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

func init() {
	nodeService = node.NewService()
}
