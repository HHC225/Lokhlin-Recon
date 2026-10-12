// common-------------------------------------
// @file      : common.go
// @time      : 2025/5/10 23:33
// -------------------------------------------

package common

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/services/assets/common"
)

var commonService common.Service

func init() {
	commonService = common.NewService()
}

// @Summary Delete data
// @Description Delete documents from the specified collection by IDs
// @Tags Common operations
// @Accept json
// @Produce json
// @Param request body common.DeleteRequest true "Deletion request parameters"
// @Success 200 {object} response.Response
// @Router /api/asset/common/delete [post]
func DeleteData(c *gin.Context) {
	var req models.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := commonService.DeleteData(c, &req); err != nil {
		switch err {
		case common.ErrInvalidCollection:
			response.NotFound(c, "api.not_found", err)
		case common.ErrNoValidIDs, common.ErrNoDocumentsDeleted:
			response.NotFound(c, "api.not_found", err)
		default:
			response.InternalServerError(c, "api.error", err)
		}
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Add tag
// @Description Add a tag to the specified document
// @Tags Common operations
// @Accept json
// @Produce json
// @Param request body common.TagRequest true "Tag request parameters"
// @Success 200 {object} response.Response
// @Router /api/assets/common/add_tag [post]
func AddTag(c *gin.Context) {
	var req models.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := commonService.AddTag(c, &req); err != nil {
		switch err {
		case common.ErrInvalidRequest, common.ErrInvalidID:
			response.NotFound(c, "api.not_found", err)
		default:
			logger.Error(fmt.Sprintf("add_tag error: %v", err))
			response.InternalServerError(c, "api.error", err)
		}
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Delete tag
// @Description Remove a tag from the specified document
// @Tags Common operations
// @Accept json
// @Produce json
// @Param request body common.TagRequest true "Tag request parameters"
// @Success 200 {object} response.Response
// @Router /api/common/delete_tag [post]
func DeleteTag(c *gin.Context) {
	var req models.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := commonService.DeleteTag(c, &req); err != nil {
		switch err {
		case common.ErrInvalidRequest, common.ErrInvalidID:
			response.NotFound(c, "api.not_found", err)
		default:
			response.InternalServerError(c, "api.error", err)
		}
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Update status
// @Description Update the specified document status
// @Tags Common operations
// @Accept json
// @Produce json
// @Param request body common.StatusRequest true "Status request parameters"
// @Success 200 {object} response.Response
// @Router /api/common/update_status [post]
func UpdateStatus(c *gin.Context) {
	var req models.StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := commonService.UpdateStatus(c, &req); err != nil {
		switch err {
		case common.ErrInvalidRequest, common.ErrInvalidID:
			response.NotFound(c, "api.not_found", err)
		default:
			response.InternalServerError(c, "api.error", err)
		}
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Count records
// @Description Count documents in the specified collection
// @Tags Common operations
// @Accept json
// @Produce json
// @Param request body common.TotalRequest true "Count request parameters"
// @Success 200 {object} response.Response
// @Router /api/assets/common/total [post]
func TotalData(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	count, err := commonService.TotalData(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case common.ErrInvalidCollection:
			response.NotFound(c, "api.not_found", err)
		default:
			response.InternalServerError(c, "api.error", err)
		}
		return
	}

	response.Success(c, gin.H{"total": count}, "api.success")
}
