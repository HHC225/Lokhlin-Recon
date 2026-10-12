package mp

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/services/assets/mp"
)

var mpService mp.Service

func init() {
	mpService = mp.NewService()
}

// GetMPData godoc
// @Summary List mini program data
// @Description List mini program data
// @Tags Mini program management
// @Accept json
// @Produce json
// @Param request body models.SearchRequest true "Request parameters"
// @Success 200 {object} response.Response{data=models.MPResponse} "Success"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 500 {object} response.Response "Server error"
// @Router /api/assets/mp [post]
func GetMPData(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := mpService.GetMPData(c, req)
	if err != nil {
		logger.Error(fmt.Sprintf("Get mp data error: %s", err.Error()))
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}
