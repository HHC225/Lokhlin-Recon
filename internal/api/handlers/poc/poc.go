package poc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/services/poc"
	"recon/internal/utils/random"
)

var pocService poc.Service

func init() {
	pocService = poc.NewService()
}

// GetPocList godoc
// @Summary List POCs
// @Description List POCs，Supports pagination and search
// @Tags POC management
// @Accept json
// @Produce json
// @Param request body poc.PocListRequest true "Request parameters"
// @Success 200 {object} response.Response{data=poc.PocListResponse}
// @Router /poc [post]
func GetPocList(c *gin.Context) {
	var req models.PocListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(fmt.Sprintf("GetPocList bad request%v", err))
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := pocService.GetPocList(c, &req)
	if err != nil {
		logger.Error(fmt.Sprintf("GetPocList error %v", err))
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

// GetPocDetail godoc
// @Summary Get POC details
// @Description Get POC details by ID
// @Tags POC management
// @Accept json
// @Produce json
// @Param request body poc.PocDetailRequest true "Request parameters"
// @Success 200 {object} response.Response{data=poc.Poc}
// @Router /poc/detail [post]
func GetPocDetail(c *gin.Context) {
	var req models.PocDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := pocService.GetPocDetail(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

// UpdatePoc godoc
// @Summary Update POC
// @Description Update POC information
// @Tags POC management
// @Accept json
// @Produce json
// @Param request body poc.PocUpdateRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /poc/update [post]
func UpdatePoc(c *gin.Context) {
	var req models.PocUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pocService.UpdatePoc(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// AddPoc godoc
// @Summary Add POC
// @Description Create a POC
// @Tags POC management
// @Accept json
// @Produce json
// @Param request body poc.PocAddRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/poc/add [post]
func AddPoc(c *gin.Context) {
	var req models.PocAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if _, err := pocService.AddPoc(c.Request.Context(), &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// DeletePoc godoc
// @Summary Delete POC
// @Description Delete specified POC
// @Tags POC management
// @Accept json
// @Produce json
// @Param request body poc.PocDeleteRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /poc/delete [post]
func DeletePoc(c *gin.Context) {
	var req models.PocDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pocService.DeletePoc(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// GetAllPocData godoc
// @Summary List all POC data
// @Description List all POCs with id, name, time and tags
// @Tags POC management
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /poc/data/all [get]
func GetAllPocData(c *gin.Context) {
	result, err := pocService.GetAllPocData(context.Background())
	if err != nil {
		logger.Error(fmt.Sprintf("GetAllPocData error %v", err))
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, map[string]interface{}{
		"list": result,
	}, "api.success")
}

// ImportPoc godoc
// @Summary Import POC files
// @Description Upload ZIP and import POC files
// @Tags POC management
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "POCFile(ZIP format)"
// @Success 200 {object} response.Response{data=models.PocImportResponse}
// @Router /poc/data/import [post]
func ImportPoc(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		logger.Error(fmt.Sprintf("ImportPoc get file error %v", err))
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".zip") {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("file must be zip format"))
		return
	}

	if strings.Contains(file.Filename, "..") || strings.Contains(file.Filename, "/") || strings.Contains(file.Filename, "\\") {
		response.BadRequest(c, "api.bad_request", fmt.Errorf("unsafe filename"))
		return
	}

	tempDir := os.TempDir()
	fileName := random.GenerateString(5) + ".zip"
	filePath := filepath.Join(tempDir, fileName)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		logger.Error(fmt.Sprintf("ImportPoc save file error %v", err))
		response.InternalServerError(c, "api.error", err)
		return
	}

	go func() {
		result, err := pocService.ImportPoc(c, filePath)
		if err != nil {
			logger.Error(fmt.Sprintf("ImportPoc process error %v", err))
		} else {
			logger.Info(fmt.Sprintf("ImportPoc success: %+v", result))
		}
	}()

	response.Success(c, map[string]interface{}{
		"message": "Importing in progress",
	}, "api.success")
}
