package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/services/assets/app"
)

var appService app.Service

func init() {
	appService = app.NewService()
}

// GetAppData godoc
// @Summary List application data
// @Description List application data
// @Tags Application management
// @Accept json
// @Produce json
// @Param request body models.SearchRequest true "Request parameters"
// @Success 200 {object} response.Response{data=models.AppResponse} "Success"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 500 {object} response.Response "Server error"
// @Router /api/assets/app [post]
func GetAppData(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := appService.GetAppData(c, req)
	if err != nil {
		logger.Error(fmt.Sprintf("Get app data error: %s", err.Error()))
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}
