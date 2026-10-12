// ip-------------------------------------
// @file      : ip.go
// @time      : 2025/11/15 21:17
// -------------------------------------------

package ip

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/models"
	ipservice "recon/internal/services/assets/ip"
)

var ipService ipservice.Service

func init() {
	ipService = ipservice.NewService()
}

// GetIPAssets @Summary      List IP assets
// @Description  Query IP asset information
// @Tags         Assets
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      models.SearchRequest  true  "Query conditions"
// @Success      200      {object}  response.SuccessResponse{data=models.IPAssetResponse}
// @Failure      400      {object}  response.BadRequestResponse
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /api/assets/ip [post]
func GetIPAssets(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := ipService.GetIPAssets(c, req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}
