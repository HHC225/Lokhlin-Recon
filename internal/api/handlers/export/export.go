// export-------------------------------------
// @file      : export.go
// @time      : 2025/5/25 14:34
// -------------------------------------------

package export

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"path/filepath"
	"recon/internal/api/response"
	"recon/internal/config"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/services/export"
	"recon/internal/utils/helper"
)

var exportService export.Service

func init() {
	exportService = export.NewService()
}

// @Summary Export data
// @Description Export requested data to a file
// @Tags export
// @Accept json
// @Produce json
// @Param request body ExportRequest true "Export request parameters"
// @Success 200 {object} ExportResponse
// @Router /api/export [post]
func Export(c *gin.Context) {
	var req models.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if req.Index == "" || req.Quantity == 0 || req.Type == "" {
		response.BadRequest(c, "api.invalid_params", fmt.Errorf("index, quantity, and type are required"))
		return
	}

	if req.FileType != "xlsx" && req.FileType != "json" {
		response.BadRequest(c, "api.invalid_file_type", fmt.Errorf("filetype must be csv or json"))
		return
	}

	if len(req.Field) == 0 {
		response.BadRequest(c, "api.invalid_fields", fmt.Errorf("fields cannot be empty"))
		return
	}
	searchQuery, err := helper.GetSearchQuery(models.SearchRequest{
		Index:            req.Index,
		SearchExpression: req.SearchExpression,
		Filter:           req.Filter,
		FuzzyQuery:       req.FuzzyQuery,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("get search query error:%s", err.Error()))
		response.InternalServerError(c, "api.export_failed", err)
		return
	}
	if req.Index == "PageMonitoring" {
		searchQuery["hash"] = map[string]interface{}{"$size": 2}
	}
	filter := bson.M(searchQuery)

	fileName := fmt.Sprintf("%s.%s", uuid.New().String(), req.FileType)

	err = exportService.CreateExportTask(c, &models.ExportTask{
		FileName:   fileName,
		CreateTime: helper.GetNowTimeString(),
		Quantity:   req.Quantity,
		DataType:   req.Index,
		FileType:   req.FileType,
		State:      0,
	})

	if err != nil {
		response.InternalServerError(c, "api.export_failed", err)
		return
	}
	run := models.ExportRun{
		Filter:   filter,
		FileName: fileName,
		Index:    req.Index,
		FileType: req.FileType,
		Field:    req.Field,
		Quantity: req.Quantity,
	}

	go exportService.ProcessExportTask(context.Background(), run)

	response.Success(c, models.ExportResponse{
		Message: "Successfully added data export task",
		Code:    200,
	}, "api.success")
}

// @Summary List export fields
// @Description List exportable fields for an index
// @Tags export
// @Accept json
// @Produce json
// @Param request body map[string]string true "Request parameters"
// @Success 200 {object} map[string]interface{}
// @Router /api/export/fields [post]
func GetExportFields(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	index := req["index"]
	fields, err := exportService.GetExportFields(c, index)
	if err != nil {
		response.InternalServerError(c, "api.get_fields_failed", err)
		return
	}

	response.Success(c, map[string]interface{}{
		"field": fields,
	}, "api.success")
}

// @Summary List export records
// @Description List all export task records
// @Tags export
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/export/records [get]
func GetExportRecords(c *gin.Context) {
	records, err := exportService.GetExportRecords(c)
	if err != nil {
		response.InternalServerError(c, "api.get_records_failed", err)
		return
	}

	response.Success(c, map[string]interface{}{
		"list": records,
	}, "api.success")
}

// @Summary Delete export file
// @Description Delete the specified export file
// @Tags export
// @Accept json
// @Produce json
// @Param request body map[string][]string true "Request parameters"
// @Success 200 {object} ExportResponse
// @Router /api/export/delete [post]
func DeleteExport(c *gin.Context) {
	var req map[string][]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	ids := req["ids"]
	if len(ids) == 0 {
		response.NotFound(c, "api.no_files_selected", nil)
		return
	}

	err := exportService.DeleteExportFiles(c, ids)
	if err != nil {
		response.InternalServerError(c, "api.delete_failed", err)
		return
	}

	response.Success(c, models.ExportResponse{
		Message: "Export files deleted successfully",
		Code:    200,
	}, "api.success")
}

// @Summary Download export file
// @Description Download the specified export file
// @Tags export
// @Produce application/octet-stream
// @Param filename query string true "Filename"
// @Success 200 {file} binary
// @Router /api/export/download [get]
func DownloadExport(c *gin.Context) {
	fileName := c.Query("filename")
	if !exportService.IsValidFileName(fileName) {
		response.BadRequest(c, "api.invalid_filename", nil)
		return
	}
	filePath := filepath.Join(config.GlobalConfig.System.ExeDir, "files", "export", fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.NotFound(c, "api.file_not_found", nil)
		return
	}

	c.File(filePath)
}
