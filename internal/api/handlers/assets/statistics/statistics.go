package statistics

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/services/assets/statistics"
)

var statisticsService statistics.Service

func init() {
	statisticsService = statistics.NewService()
}

// GetAssetStatisticsData godoc
// @Summary Get asset statistics
// @Description Count assets by type
// @Tags Asset statistics
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=statistics.StatisticsData}
// @Router /api/assets/statistics [get]
func GetAssetStatisticsData(c *gin.Context) {
	data, err := statisticsService.GetStatisticsData(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, data, "")
}

// GetAssetPortStatistics godoc
// @Summary Get port statistics
// @Description Get asset port distribution
// @Tags Asset statistics
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body statistics.PortStatisticsRequest true "Query parameters"
// @Success 200 {object} response.Response{data=statistics.PortStatisticsResponse}
// @Router /api/assets/statistics/port [post]
func GetAssetPortStatistics(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	data, err := statisticsService.GetPortStatistics(c, &req)
	if err != nil {
		logger.Error(fmt.Sprintf("GetPortStatistics err:%v", err))
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, data, "")
}

// GetAssetTitleStatistics godoc
// @Summary Get title statistics
// @Description Get asset title distribution
// @Tags Asset statistics
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body statistics.TitleStatisticsRequest true "Query parameters"
// @Success 200 {object} response.Response{data=statistics.TitleStatisticsResponse}
// @Router /api/v1/assets/statistics/title [post]
func GetAssetTitleStatistics(c *gin.Context) {
	var req models.TitleStatisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	data, err := statisticsService.GetTitleStatistics(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, data, "")
}

// GetAssetServiceStatistics godoc
// @Summary Get service type statistics
// @Description Get asset service type distribution
// @Tags Asset statistics
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body statistics.TypeStatisticsRequest true "Query parameters"
// @Success 200 {object} response.Response{data=statistics.TypeStatisticsResponse}
// @Router /api/assets/statistics/service [post]
func GetAssetServiceStatistics(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	data, err := statisticsService.GetTypeStatistics(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, data, "")
}

// GetAssetIconStatistics godoc
// @Summary Get icon statistics
// @Description Get asset icon distribution
// @Tags Asset statistics
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body statistics.IconStatisticsRequest true "Query parameters"
// @Success 200 {object} response.Response{data=statistics.IconStatisticsResponse}
// @Router /api/assets/statistics/icon [post]
func GetAssetIconStatistics(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	data, err := statisticsService.GetIconStatistics(c, &req)
	if err != nil {
		logger.Error(fmt.Sprintf("GetIconStatistics err:%v", err))
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, data, "")
}

// GetAssetAppStatistics godoc
// @Summary Get application statistics
// @Description Get asset application distribution
// @Tags Asset statistics
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body statistics.AppStatisticsRequest true "Query parameters"
// @Success 200 {object} response.Response{data=statistics.AppStatisticsResponse}
// @Router /api/assets/statistics/app [post]
func GetAssetAppStatistics(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	data, err := statisticsService.GetAppStatistics(c, &req)
	if err != nil {
		logger.Error(fmt.Sprintf("GetAppStatistics err:%v", err))
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, data, "")
}
