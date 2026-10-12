// configuration-------------------------------------
// @file      : rad.go
// @time      : 2025/10/29 20:58
// -------------------------------------------

package configuration

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/services/configuration"
)

var radConfigService *configuration.Service

// GetRadData @Summary Get  Rad  configuration
// @Tags        configuration
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} response.SuccessResponse{data=object{content=string}}
// @Router      /configuration/rad/data [get]
func GetRadData(c *gin.Context) {
	content, err := radConfigService.GetRadContent(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"content": content}, "api.success")
}

// SaveRadData @Summary Save  Rad  configuration
// @Tags        configuration
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body saveContentRequest true "Configuration content"
// @Success     200 {object} response.SuccessResponse
// @Router      /configuration/rad/save [post]
func SaveRadData(c *gin.Context) {
	var req saveContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if err := radConfigService.SaveRadContent(c, req.Content); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

func init() {
	radConfigService = configuration.NewService()
}
