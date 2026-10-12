// subdomain-------------------------------------
// @file      : taker.go
// @time      : 2025/5/14 21:53
// -------------------------------------------

package subdomain

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/models"
	"recon/internal/services/assets/subdomain"
)

var takerService subdomain.TakerService

func init() {
	takerService = subdomain.NewTakerService()
}

// GetSubdomainTakerData godoc
// @Summary List subdomain takeover data
// @Description List subdomain takeover data
// @Tags Subdomain takeover
// @Accept json
// @Produce json
// @Param request body subdomain.SubdomainTakerRequest true "Request parameters"
// @Success 200 {object} response.Response{data=subdomain.SubdomainTakerResponse} "Success"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 500 {object} response.Response "Server error"
// @Router /api/assets/subdomain/taker [post]
func GetSubdomainTakerData(c *gin.Context) {
	var request models.SearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := takerService.GetSubdomainTakerData(c, request)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}
