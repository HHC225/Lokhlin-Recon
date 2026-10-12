package project

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/models"
	"recon/internal/services/project"
)

var projectService project.Service

func init() {
	projectService = project.NewService()
}

// GetProjectsByTag godoc
// @Summary List projects grouped by tag
// @Description List all projects grouped by tag
// @Tags Projects
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]project.TagGroup}
// @Router /api/projects/all [get]
func GetProjectsByTag(c *gin.Context) {
	result, err := projectService.GetProjectsByTag(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, gin.H{
		"list": result,
	}, "")
}

// @Summary Get project data
// @Description Return paginated projects and project counts for each tag
// @Tags Projects
// @Accept json
// @Produce json
// @Param data body models.ProjectListRequest true "Query parameters"
// @Success 200 {object} response.Response{data=models.ProjectListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/project/data [post]
func GetProjectsData(c *gin.Context) {
	var req models.ProjectListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if req.PageIndex <= 0 {
		req.PageIndex = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	data, err := projectService.GetProjectsData(c, req.Search, req.PageIndex, req.PageSize)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, data, "")
}

// @Summary Get project targets and basic information
// @Tags Projects
// @Accept json
// @Produce json
// @Param data body models.ProjectContentRequest true "Project ID"
// @Success 200 {object} response.Response{data=models.ProjectContentResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/project/content [post]
func GetProjectContent(c *gin.Context) {
	var req models.ProjectContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	data, err := projectService.GetProjectContent(c, req.ID)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	if data == nil {
		response.NotFound(c, "api.not_found", nil)
		return
	}
	response.Success(c, data, "")
}

func AddProject(c *gin.Context) {
	var req models.Project
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	err := projectService.AddProject(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
	return
}

func UpdatePorject(c *gin.Context) {
	var req models.UpdateProject
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	err := projectService.UpdateProject(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
	return
}

func DeleteProject(c *gin.Context) {
	var req models.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	err := projectService.DeleteProjects(c, req.IDs, req.DeleateAssets)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
}
