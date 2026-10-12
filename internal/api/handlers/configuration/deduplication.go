// configuration-------------------------------------
// @file      : deduplication.go
// @time      : 2025/10/29 20:58
// -------------------------------------------

package configuration

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/models"
	"recon/internal/services/configuration"
)

var dedupService *configuration.Service

type saveDedupRequest struct {
	RunNow bool `json:"runNow"`
}

// GetDeduplicationConfig @Summary Get deduplication configuration
// @Tags        configuration
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} response.SuccessResponse{data=object}
// @Router      /configuration/deduplication/config [get]
func GetDeduplicationConfig(c *gin.Context) {
	data, err := dedupService.GetDeduplicationConfig(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, data, "api.success")
}

// SaveDeduplicationConfig @Summary Save deduplication configuration
// @Tags        configuration
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body map[string]interface{} true "Deduplication configuration"
// @Success     200 {object} response.SuccessResponse
// @Router      /configuration/deduplication/save [post]
func SaveDeduplicationConfig(c *gin.Context) {
	var body models.DepConfig
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if err := dedupService.SaveDeduplicationConfig(c, body); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

func init() {
	dedupService = configuration.NewService()
}
