// dirscan-------------------------------------
// @file      : dirscan.go
// @time      : 2025/5/17 15:06
// -------------------------------------------

package dirscan

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/models"
	"recon/internal/services/assets/dirscan"
)

var dirscanService dirscan.Service

func init() {
	dirscanService = dirscan.NewService()
}

// List @Summary List directory scan data
// @Tags DirScan
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body ListRequest true "Pagination and filters"
// @Success 200 {object} response.DataResponse{data=[]models.DirScanResult}
// @Router /api/assets/dirscan [post]
func List(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	results, err := dirscanService.List(c, req)
	if err != nil {
		response.InternalServerError(c, "api.dirscan.list.failed", err)
		return
	}

	response.Success(c, gin.H{
		"list": results,
	}, "api.dirscan.list.success")
}
