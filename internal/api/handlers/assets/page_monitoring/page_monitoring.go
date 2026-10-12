package page_monitoring

import (
	"errors"
	"fmt"

	"recon/internal/logger"
	"recon/internal/services/assets/page_monitoring"

	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/models"
)

var pageMonitoringService page_monitoring.Service

func init() {
	pageMonitoringService = page_monitoring.NewService()
}

// GetResult @Summary      List page monitoring results
// @Description  List page monitoring results endpoint
// @Tags         page_monitoring
// @Accept       json
// @Produce      json
// @Param        resultRequest  body      models.SearchRequest  true  "Results request"
// @Success      200  {object}  response.SuccessResponse{data=models.PageMonitoringResultResponse}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /api/assets/page-monitoring [post]
func GetResult(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := pageMonitoringService.GetResult(c, req)
	if err != nil {
		logger.Error(fmt.Sprintf("page-monitoring GetResult error: %s", err))
		response.InternalServerError(c, "api.page_monitoring.result.error", err)
		return
	}

	response.Success(c, result, "api.page_monitoring.result.success")
}

// GetDiff @Summary      Get page monitoring diff
// @Description  Get page monitoring diff endpoint
// @Tags         page_monitoring
// @Accept       json
// @Produce      json
// @Param        diffRequest  body      models.PageMonitoringDiffRequest  true  "Diff request"
// @Success      200  {object}  response.SuccessResponse{data=models.PageMonitoringBody}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      404  {object}  response.NotFoundResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /page-monitoring/diff [post]
func GetDiff(c *gin.Context) {
	var req models.PageMonitoringDiffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if req.ID == "" {
		response.BadRequest(c, "api.page_monitoring.id_missing", nil)
		return
	}

	result, err := pageMonitoringService.GetDiff(c.Request.Context(), req.ID)
	if err != nil {
		if errors.Is(err, page_monitoring.ErrPageNotFound) {
			response.NotFound(c, "api.page_monitoring.diff.not_found", err)
			return
		}
		logger.Error(fmt.Sprintf("page-monitoring GetDiff error: %s", err))
		response.InternalServerError(c, "api.page_monitoring.diff.error", err)
		return
	}

	response.Success(c, result, "api.page_monitoring.diff.success")
}

// GetHistory @Summary      Get page monitoring history
// @Description  Get page monitoring history endpoint
// @Tags         page_monitoring
// @Accept       json
// @Produce      json
// @Param        historyRequest  body      models.PageMonitoringHistoryRequest  true  "History request"
// @Success      200  {object}  response.SuccessResponse{data=[]models.PageMonitoring}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      404  {object}  response.NotFoundResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /page-monitoring/history [post]
func GetHistory(c *gin.Context) {
	var req models.PageMonitoringHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if req.ID == "" {
		response.BadRequest(c, "api.page_monitoring.id_missing", nil)
		return
	}

	result, err := pageMonitoringService.GetHistory(c.Request.Context(), req.ID)
	if err != nil {
		if errors.Is(err, page_monitoring.ErrPageNotFound) {
			response.NotFound(c, "api.page_monitoring.history.not_found", err)
			return
		}
		response.InternalServerError(c, "api.page_monitoring.history.error", err)
		return
	}

	response.Success(c, result, "api.page_monitoring.history.success")
}

// GetContent @Summary      Get page monitoring content
// @Description  Get page monitoring content endpoint
// @Tags         page_monitoring
// @Accept       json
// @Produce      json
// @Param        contentRequest  body      models.PageMonitoringContentRequest  true  "Content request"
// @Success      200  {object}  response.SuccessResponse{data=models.PageMonitoringBody}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      404  {object}  response.NotFoundResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /page-monitoring/content [post]
func GetContent(c *gin.Context) {
	var req models.PageMonitoringContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if req.ID == "" {
		response.BadRequest(c, "api.page_monitoring.id_missing", nil)
		return
	}

	if req.Flag != "current" && req.Flag != "previous" {
		response.BadRequest(c, "api.page_monitoring.invalid_flag", nil)
		return
	}

	result, err := pageMonitoringService.GetContent(c.Request.Context(), req.ID, req.Flag)
	if err != nil {
		if err == page_monitoring.ErrPageNotFound {
			response.NotFound(c, "api.page_monitoring.content.not_found", err)
			return
		}
		response.InternalServerError(c, "api.page_monitoring.content.error", err)
		return
	}

	response.Success(c, result, "api.page_monitoring.content.success")
}
