package task

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/services/task/template"
)

var templateService template.Service

func init() {
	templateService = template.NewService()
}

type ListRequest struct {
	PageIndex int    `json:"pageIndex" binding:"required" example:"1"`
	PageSize  int    `json:"pageSize" binding:"required" example:"10"`
	Query     string `json:"query" example:"test"`
}

// @Summary List templates
// @Description List scan templates
// @Tags Template management
// @Accept json
// @Produce json
// @Param request body ListRequest true "Request parameters"
// @Success 200 {object} response.Response{data=task.TemplateList} "Success"
// @Router /api/task/template [post]
func TemplateList(c *gin.Context) {
	var req ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := templateService.List(c.Request.Context(), req.PageIndex, req.PageSize, req.Query)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

type DetailRequest struct {
	ID string `json:"id" binding:"required" example:"507f1f77bcf86cd799439011"`
}

// @Summary Get template details
// @Description Get scan template details
// @Tags Template management
// @Accept json
// @Produce json
// @Param request body DetailRequest true "Request parameters"
// @Success 200 {object} response.Response{data=task.Template} "Success"
// @Router /api/task/template/detail [post]
func TemplateDetail(c *gin.Context) {
	var req DetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := templateService.Detail(c, req.ID)
	if err != nil {
		logger.Error(fmt.Sprintf("TemplateDetail error: %v", err))
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

type SaveRequest struct {
	ID     string              `json:"id" example:"507f1f77bcf86cd799439011"`
	Result models.ScanTemplate `json:"result" binding:"required"`
}

// @Summary Save template
// @Description Save or update a scan template
// @Tags Template management
// @Accept json
// @Produce json
// @Param request body SaveRequest true "Request parameters"
// @Success 200 {object} response.Response "Success"
// @Router /api/task/template/save [post]
func Save(c *gin.Context) {
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	_, err := templateService.Save(c.Request.Context(), req.ID, &req.Result)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

type DeleteRequest struct {
	IDs []string `json:"ids" binding:"required" example:"['507f1f77bcf86cd799439011']"`
}

// @Summary Delete template
// @Description Delete scan template
// @Tags Template management
// @Accept json
// @Produce json
// @Param request body DeleteRequest true "Request parameters"
// @Success 200 {object} response.Response "Success"
// @Router /api/task/template/delete [post]
func Delete(c *gin.Context) {
	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	err := templateService.Delete(c, req.IDs)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}
