package sensitive

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/services/assets/sensitive"
)

var sensitiveService sensitive.Service

func init() {
	sensitiveService = sensitive.NewService()
}

// GetSensitiveInfo godoc
// @Summary List sensitive information
// @Description List sensitive information，Supports pagination, sorting and filtering
// @Tags Sensitive
// @Accept json
// @Produce json
// @Param query body models.SearchRequest true "Query parameters"
// @Success 200 {object} response.SuccessResponse{data=object{list=[]models.SensitiveInfo}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/sensitive [post]
func GetSensitiveInfo(c *gin.Context) {
	var query models.SearchRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	sensitiveInfo, err := sensitiveService.GetSensitiveInfo(c, query)
	if err != nil {
		logger.Error(fmt.Sprintf("GetSensitiveInfo err:%v", err))
		response.InternalServerError(c, "api.sensitive.list.failed", err)
		return
	}

	response.Success(c, gin.H{
		"list": sensitiveInfo,
	}, "api.sensitive.list.success")
}

// GetSensitiveInfoNumber godoc
// @Summary List sensitive information
// @Description List sensitive information，Supports pagination, sorting and filtering
// @Tags Sensitive
// @Accept json
// @Produce json
// @Param query body models.SearchRequest true "Query parameters"
// @Success 200 {object} response.SuccessResponse{data=object{list=[]models.SensitiveInfo}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/sensitive/number [post]
func GetSensitiveInfoNumber(c *gin.Context) {
	var query models.SearchRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	totalCount, urlCount, err := sensitiveService.GetSensitiveInfoNumber(c, query)
	if err != nil {
		logger.Error(fmt.Sprintf("GetSensitiveInfoNumber err:%v", err))
		response.InternalServerError(c, "api.sensitive.number.failed", err)
		return
	}

	response.Success(c, gin.H{
		"total": urlCount,
		"all":   totalCount,
	}, "")
}

// GetSensitiveInfoBody godoc
// @Summary List sensitive information
// @Description List sensitive information，Supports pagination, sorting and filtering
// @Tags Sensitive
// @Accept json
// @Produce json
// @Param query body models.SearchRequest true "Query parameters"
// @Success 200 {object} response.SuccessResponse{data=object{list=[]models.SensitiveInfo}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/sensitive/number [post]
func GetSensitiveInfoBody(c *gin.Context) {
	var query models.GetSensitiveBodyRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	body, err := sensitiveService.GetBodyByID(c, query.ID)
	if err != nil {
		logger.Error(fmt.Sprintf("GetSensitiveInfoBody err:%v", err))
		response.InternalServerError(c, "api.sensitive.body.failed", err)
		return
	}

	response.Success(c, gin.H{
		"body": body,
	}, "")
}

// GetSensitiveInfoName godoc
// @Summary List sensitive information
// @Description List sensitive information，Supports pagination, sorting and filtering
// @Tags Sensitive
// @Accept json
// @Produce json
// @Param query body models.SearchRequest true "Query parameters"
// @Success 200 {object} response.SuccessResponse{data=object{list=[]models.SensitiveInfo}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/sensitive/names [post]
func GetSensitiveInfoName(c *gin.Context) {
	var query models.SearchRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	sensitiveNameInfo, err := sensitiveService.GetSIDStatistics(c, query)
	if err != nil {
		logger.Error(fmt.Sprintf("GetSensitiveInfo err:%v", err))
		response.InternalServerError(c, "api.sensitive.list.failed", err)
		return
	}

	response.Success(c, gin.H{
		"list": sensitiveNameInfo,
	}, "api.sensitive.list.success")
}

// GetSensitiveMatchInfo godoc
// @Summary Get deduplicated matches for the specified sid
// @Tags SensitiveResult
// @Accept json
// @Produce json
// @Param data body request.GetSensitiveMatchInfoRequest true "Request body"
// @Success 200 {object} response.DataResponse[response.MatchListResponse]
// @Failure 400 {object} response.ErrorResponse
// @Router /api/assets/sensitive/info [post]
func GetSensitiveMatchInfo(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if req.Sid == "" {
		response.BadRequest(c, "api.bad_request", errors.New("sid is required"))
		return
	}

	data, err := sensitiveService.GetMatchInfo(c, req)
	if err != nil {
		logger.Error(fmt.Sprintf("GetSensitiveMatchInfo err:%v", err))
		response.InternalServerError(c, "api.sensitive.info.failed", err)
		return
	}

	response.Success(c, data, "api.success")
}
