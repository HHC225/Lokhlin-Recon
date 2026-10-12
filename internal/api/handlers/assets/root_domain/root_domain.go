package root_domain

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/services/assets/root_domain"
)

var rootDomainService root_domain.Service

func init() {
	rootDomainService = root_domain.NewService()
}

// GetRootDomainData godoc
// @Summary List root domain data
// @Description List root domains with pagination
// @Tags Root domain management
// @Accept json
// @Produce json
// @Param request body GetRootDomainDataRequest true "Request parameters"
// @Success 200 {object} response.Response{data=map[string]interface{}} "Success"
// @Failure 400 {object} response.Response "Invalid request parameters"
// @Failure 500 {object} response.Response "Internal server error"
// @Security ApiKeyAuth
// @Router /api/assets/root_domain [post]
func GetRootDomainData(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := rootDomainService.GetRootDomainData(c, req)
	if err != nil {
		logger.Error(fmt.Sprintf("Get root domain data error: %s", err.Error()))
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}
