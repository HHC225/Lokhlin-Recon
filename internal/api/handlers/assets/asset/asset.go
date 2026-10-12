package asset

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"recon/internal/api/response"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/services/assets/asset"
)

var assetService asset.Service

func init() {
	assetService = asset.NewService()
}

// GetAssets godoc
// @Summary List assets
// @Description List assets，Supports pagination, sorting and filtering
// @Tags Assets
// @Accept json
// @Produce json
// @Param query body models.SearchRequest true "Query parameters"
// @Success 200 {object} response.SuccessResponse{data=object{list=[]models.Asset}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/asset [post]
func GetAssets(c *gin.Context) {
	var query models.SearchRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	assets, err := assetService.GetAssets(c, query)
	if err != nil {
		response.InternalServerError(c, "api.asset.list.failed", err)
		return
	}

	response.Success(c, gin.H{
		"list": assets,
	}, "api.asset.list.success")
}

// GetAssetByID godoc
// @Summary Get asset
// @Description Get asset details by ID
// @Tags Assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Success 200 {object} response.SuccessResponse{data=models.Asset}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/asset/detail [get]
func GetAssetByID(c *gin.Context) {
	var query models.IdRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	asset, err := assetService.GetAssetByID(c, query.ID)
	if err != nil {
		logger.Error(fmt.Sprintf("%v", err))
		response.InternalServerError(c, "api.asset.get.failed", err)
		return
	}

	if asset == nil {
		response.NotFound(c, "api.asset.not_found", nil)
		return
	}

	response.Success(c, asset, "api.asset.get.success")
}

// CreateAsset godoc
// @Summary Create asset
// @Description Create an asset
// @Tags Assets
// @Accept json
// @Produce json
// @Param asset body models.Asset true "Asset information"
// @Success 201 {object} response.SuccessResponse{data=models.Asset}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets [post]
func CreateAsset(c *gin.Context) {
	var asset models.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := assetService.CreateAsset(c, &asset); err != nil {
		response.InternalServerError(c, "api.asset.create.failed", err)
		return
	}

	response.Success(c, asset, "api.asset.create.success")
}

// UpdateAsset godoc
// @Summary Update asset
// @Description Update asset information by ID
// @Tags Assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Param asset body models.Asset true "Asset information"
// @Success 200 {object} response.SuccessResponse{data=models.Asset}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/{id} [put]
func UpdateAsset(c *gin.Context) {
	id := c.Param("id")
	var asset models.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := assetService.UpdateAsset(c, id, &asset); err != nil {
		response.InternalServerError(c, "api.asset.update.failed", err)
		return
	}

	response.Success(c, asset, "api.asset.update.success")
}

// DeleteAsset godoc
// @Summary Delete asset
// @Description Delete an asset by ID
// @Tags Assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/{id} [delete]
func DeleteAsset(c *gin.Context) {
	id := c.Param("id")
	if err := assetService.DeleteAsset(c, id); err != nil {
		response.InternalServerError(c, "api.asset.delete.failed", err)
		return
	}

	response.Success(c, nil, "api.asset.delete.success")
}

// GetScreenshot godoc
// @Summary Get asset screenshot
// @Description Get an asset screenshot by ID
// @Tags Assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Success 200 {object} response.SuccessResponse{data=string}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/{id}/screenshot [get]
func GetScreenshot(c *gin.Context) {
	id := c.Param("id")
	screenshot, err := assetService.GetScreenshot(c, id)
	if err != nil {
		response.InternalServerError(c, "api.asset.screenshot.failed", err)
		return
	}

	if screenshot == "" {
		response.NotFound(c, "api.asset.screenshot.not_found", nil)
		return
	}

	response.Success(c, screenshot, "api.asset.screenshot.success")
}

// GetChangeLog godoc
// @Summary Get asset change log
// @Description Get asset change history by ID
// @Tags Assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Success 200 {object} response.SuccessResponse{data=[]models.AssetChangeLog}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/asset/changelog [get]
func GetChangeLog(c *gin.Context) {
	var query models.IdRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	logs, err := assetService.GetChangeLog(c, query.ID)
	if err != nil {
		response.InternalServerError(c, "api.asset.changelog.failed", err)
		return
	}

	response.Success(c, logs, "")
}

// DeduplicateAssets godoc
// @Summary Deduplicate assets
// @Description Deduplicate assets of the specified type
// @Tags Assets
// @Accept json
// @Produce json
// @Param assetType path string true "Asset type"
// @Param filter body map[string]interface{} true "Filter conditions"
// @Param groupFields body []string true "Grouping fields"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/deduplicate/{assetType} [post]
func DeduplicateAssets(c *gin.Context) {
	assetType := c.Param("assetType")
	var request struct {
		Filter      bson.M   `json:"filter" binding:"required"`
		GroupFields []string `json:"groupFields" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := assetService.DeduplicateAssets(c, assetType, request.Filter, request.GroupFields); err != nil {
		response.InternalServerError(c, "api.asset.deduplicate.failed", err)
		return
	}

	response.Success(c, nil, "api.asset.deduplicate.success")
}

// GetAssetCardData godoc
// @Summary Get asset card data
// @Description List asset card data，Supports pagination and search conditions
// @Tags Assets
// @Accept json
// @Produce json
// @Param query body models.SearchRequest true "Query parameters"
// @Success 200 {object} response.SuccessResponse{data=object{list=[]models.Asset}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/asset/card [post]
func GetAssetCardData(c *gin.Context) {
	var query models.SearchRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	assets, err := assetService.GetAssetCardData(c, query)
	if err != nil {
		response.InternalServerError(c, "api.asset.card.failed", err)
		return
	}

	response.Success(c, gin.H{
		"list": assets,
	}, "api.asset.card.success")
}

// GetHostsByRootDomain godoc
// @Summary List hosts by root domain
// @Description Match rootDomain exactly and return a paginated list of hosts
// @Tags Assets
// @Accept json
// @Produce json
// @Param query body object{rootDomain=string,pageIndex=int,pageSize=int} true "Query parameters"
// @Success 200 {object} response.SuccessResponse{data=object{list=[]string,total=int64}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/asset/hosts [post]
func GetHostsByRootDomain(c *gin.Context) {
	var request struct {
		RootDomain string `json:"rootDomain" binding:"required"`
		PageIndex  int    `json:"pageIndex"`
		PageSize   int    `json:"pageSize"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if request.PageSize <= 0 {
		request.PageSize = 10
	}
	if request.PageIndex <= 0 {
		request.PageIndex = 1
	}

	hosts, total, err := assetService.GetHostsByRootDomain(c, request.RootDomain, request.PageIndex, request.PageSize)
	if err != nil {
		response.InternalServerError(c, "api.asset.hosts.failed", err)
		return
	}

	response.Success(c, gin.H{
		"list":  hosts,
		"total": total,
	}, "api.asset.hosts.success")
}

// GetAssetsByHost godoc
// @Summary List assets by host
// @Description Query assets by host and return ip, port, service, statuscode and faviconmmh3 with pagination
// @Tags Assets
// @Accept json
// @Produce json
// @Param query body object{host=string,pageIndex=int,pageSize=int} true "Query parameters"
// @Success 200 {object} response.SuccessResponse{data=object{list=[]models.Asset,total=int64}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/asset/by-host [post]
func GetAssetsByHost(c *gin.Context) {
	var request struct {
		Host      string `json:"host" binding:"required"`
		PageIndex int    `json:"pageIndex"`
		PageSize  int    `json:"pageSize"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if request.PageSize <= 0 {
		request.PageSize = 10
	}
	if request.PageIndex <= 0 {
		request.PageIndex = 1
	}

	assets, total, err := assetService.GetAssetsByHost(c, request.Host, request.PageIndex, request.PageSize)
	if err != nil {
		response.InternalServerError(c, "api.asset.by_host.failed", err)
		return
	}

	response.Success(c, gin.H{
		"list":  assets,
		"total": total,
	}, "api.asset.by_host.success")
}
